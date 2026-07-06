package training

import (
	"database/sql"
	"net/http"
	"strconv"

	"teacher-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

type achievementRequest struct {
	AchievementURL string `json:"achievementUrl"`
}

type auditRequest struct {
	TeacherID          int64 `json:"teacherId"`
	ApplyStatus        int   `json:"applyStatus"`
	AchievementStatus  int   `json:"achievementStatus"`
	HasApplyStatus     bool
	HasAchievementStat bool
}

type createRequest struct {
	Title              string `json:"title"`
	Type               string `json:"type"`
	Level              int    `json:"level"`
	SponsorUnit        string `json:"sponsorUnit"`
	OrganizerUnit      string `json:"organizerUnit"`
	StartTime          string `json:"startTime"`
	EndTime            string `json:"endTime"`
	Location           string `json:"location"`
	Quota              int    `json:"quota"`
	Requirements       string `json:"requirements"`
	AchievementRequire string `json:"achievementRequire"`
	Status             int    `json:"status"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("", list(db))
	rg.POST("", create(db))
	rg.PUT("/:id", update(db))
	rg.DELETE("/:id", remove(db))
	rg.GET("/:id/records", trainingRecords(db))
	rg.POST("/:id/enroll", enroll(db))
	rg.DELETE("/:id/enroll", cancelEnroll(db))
	rg.POST("/:id/audit", audit(db))
	rg.POST("/:id/learning-records", submitAchievement(db))
	rg.GET("/ledgers", records(db))
	rg.GET("/statistics", statistics(db))
}

func list(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 10)
		offset := (page - 1) * size
		where, args, err := listScope(c, db)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "user scope query failed")
			return
		}
		if rawStatus := c.Query("status"); rawStatus != "" {
			status, err := strconv.Atoi(rawStatus)
			if err != nil {
				response.Fail(c, http.StatusBadRequest, "invalid status")
				return
			}
			where = appendWhere(where, "status = ?")
			args = append(args, status)
		}

		var total int
		countArgs := append([]any{}, args...)
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM training `+where, countArgs...).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training count failed")
			return
		}

		currentEmployeeNo := ""
		_ = db.QueryRowContext(c.Request.Context(), `SELECT user_id FROM teacher WHERE id = ?`, c.GetInt64("userID")).Scan(&currentEmployeeNo)
		args = append(args, size, offset)
		queryArgs := append([]any{currentEmployeeNo}, args...)
		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT tr.id, tr.title, tr.type, tr.level, COALESCE(tr.location, ''), tr.quota, tr.status,
			 COALESCE(DATE_FORMAT(tr.start_time, '%Y-%m-%d %H:%i:%s'), ''), COALESCE(DATE_FORMAT(tr.end_time, '%Y-%m-%d %H:%i:%s'), ''),
			 COALESCE(tr.sponsor_unit, ''), COALESCE(tr.organizer_unit, ''), COALESCE(tr.requirements, ''), COALESCE(tr.achievement_require, ''),
			 (SELECT COUNT(*) FROM training_record r WHERE r.training_id = tr.id),
			 COALESCE((SELECT r.apply_status FROM training_record r JOIN teacher et ON et.id = r.teacher_id WHERE r.training_id = tr.id AND et.user_id = ? LIMIT 1), -1)
			 FROM training tr `+where+`
			 ORDER BY start_time DESC, id DESC
			 LIMIT ? OFFSET ?`,
			queryArgs...,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training list failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var item trainingItem
			if err := rows.Scan(
				&item.ID,
				&item.Title,
				&item.Type,
				&item.Level,
				&item.Location,
				&item.Quota,
				&item.Status,
				&item.StartTime,
				&item.EndTime,
				&item.SponsorUnit,
				&item.OrganizerUnit,
				&item.Requirements,
				&item.AchievementRequire,
				&item.EnrolledCount,
				&item.ApplyStatus,
			); err != nil {
				response.Fail(c, http.StatusInternalServerError, "training scan failed")
				return
			}
			items = append(items, item.toJSON())
		}
		if err := rows.Err(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training rows failed")
			return
		}

		response.OK(c, pageData(page, size, total, items))
	}
}

func create(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
			response.Fail(c, http.StatusBadRequest, "title is required")
			return
		}
		if req.Type == "" {
			req.Type = "training"
		}

		result, err := db.ExecContext(
			c.Request.Context(),
			`INSERT INTO training (
				title, type, level, sponsor_unit, organizer_unit, start_time, end_time,
				location, quota, requirements, achievement_require, status, create_by
			) VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?)`,
			req.Title, req.Type, req.Level, nullString(req.SponsorUnit), nullString(req.OrganizerUnit),
			req.StartTime, req.EndTime, nullString(req.Location), req.Quota, nullString(req.Requirements),
			nullString(req.AchievementRequire), req.Status, c.GetInt64("userID"),
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training create failed")
			return
		}
		id, _ := result.LastInsertId()
		response.OK(c, gin.H{"trainingId": id})
	}
}

func update(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
			response.Fail(c, http.StatusBadRequest, "title is required")
			return
		}
		if req.Type == "" {
			req.Type = "training"
		}

		_, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE training SET
				title = ?, type = ?, level = ?, sponsor_unit = ?, organizer_unit = ?,
				start_time = NULLIF(?, ''), end_time = NULLIF(?, ''), location = ?,
				quota = ?, requirements = ?, achievement_require = ?, status = ?
			 WHERE id = ?`,
			req.Title,
			req.Type,
			req.Level,
			nullString(req.SponsorUnit),
			nullString(req.OrganizerUnit),
			req.StartTime,
			req.EndTime,
			nullString(req.Location),
			req.Quota,
			nullString(req.Requirements),
			nullString(req.AchievementRequire),
			req.Status,
			pathID(c),
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training update failed")
			return
		}
		response.OK(c, gin.H{"trainingId": pathID(c), "updated": true})
	}
}

func remove(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		trainingID := pathID(c)
		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training delete failed")
			return
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(c.Request.Context(), `DELETE FROM training_record WHERE training_id = ?`, trainingID); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training record delete failed")
			return
		}
		result, err := tx.ExecContext(c.Request.Context(), `DELETE FROM training WHERE id = ?`, trainingID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training delete failed")
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			response.Fail(c, http.StatusNotFound, "training not found")
			return
		}
		if err := tx.Commit(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training delete failed")
			return
		}
		response.OK(c, gin.H{"trainingId": trainingID, "deleted": true})
	}
}

func enroll(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		trainingID := pathID(c)
		teacherID := c.GetInt64("userID")
		var employeeNo string
		if err := db.QueryRowContext(c.Request.Context(), `SELECT user_id FROM teacher WHERE id = ?`, teacherID).Scan(&employeeNo); err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher query failed")
			return
		}

		var existing int
		if err := db.QueryRowContext(
			c.Request.Context(),
			`SELECT COUNT(*)
			 FROM training_record r
			 JOIN teacher t ON t.id = r.teacher_id
			 WHERE r.training_id = ? AND t.user_id = ?`,
			trainingID,
			employeeNo,
		).Scan(&existing); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training enroll check failed")
			return
		}
		if existing > 0 {
			response.Fail(c, http.StatusConflict, "该工号已报名，请勿重复报名")
			return
		}

		_, err := db.ExecContext(
			c.Request.Context(),
			`INSERT INTO training_record (training_id, teacher_id, apply_status)
			 VALUES (?, ?, 0)`,
			trainingID, teacherID,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training enroll failed")
			return
		}

		var enrolledCount int
		_ = db.QueryRowContext(
			c.Request.Context(),
			`SELECT COUNT(*) FROM training_record WHERE training_id = ?`,
			trainingID,
		).Scan(&enrolledCount)
		response.OK(c, gin.H{"trainingId": trainingID, "enrolled": true, "enrolledCount": enrolledCount})
	}
}

func cancelEnroll(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		trainingID := pathID(c)
		teacherID := c.GetInt64("userID")
		var employeeNo string
		if err := db.QueryRowContext(c.Request.Context(), `SELECT user_id FROM teacher WHERE id = ?`, teacherID).Scan(&employeeNo); err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher query failed")
			return
		}

		result, err := db.ExecContext(
			c.Request.Context(),
			`DELETE r
			 FROM training_record r
			 JOIN teacher t ON t.id = r.teacher_id
			 WHERE r.training_id = ? AND t.user_id = ?`,
			trainingID,
			employeeNo,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training cancel enroll failed")
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			response.Fail(c, http.StatusNotFound, "未找到该工号的报名记录")
			return
		}

		var enrolledCount int
		_ = db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM training_record WHERE training_id = ?`, trainingID).Scan(&enrolledCount)
		response.OK(c, gin.H{"trainingId": trainingID, "enrolled": false, "enrolledCount": enrolledCount})
	}
}

func trainingRecords(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 20)
		offset := (page - 1) * size
		trainingID := pathID(c)

		var total int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM training_record WHERE training_id = ?`, trainingID).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training record count failed")
			return
		}

		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT r.id, r.training_id, r.teacher_id, t.name, t.user_id, COALESCE(t.department, ''), t.college,
			 COALESCE(t.phone, ''), COALESCE(t.email, ''), r.apply_status,
			 COALESCE(DATE_FORMAT(r.sign_in_time, '%Y-%m-%d %H:%i:%s'), ''), r.study_hours,
			 r.achievement_status, COALESCE(r.achievement_url, ''),
			 DATE_FORMAT(r.create_time, '%Y-%m-%d %H:%i:%s')
			 FROM training_record r
			 JOIN teacher t ON t.id = r.teacher_id
			 WHERE r.training_id = ?
			 ORDER BY r.create_time DESC
			 LIMIT ? OFFSET ?`,
			trainingID,
			size,
			offset,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training record query failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var item trainingRecordItem
			if err := rows.Scan(
				&item.ID,
				&item.TrainingID,
				&item.TeacherID,
				&item.TeacherName,
				&item.EmployeeNo,
				&item.Department,
				&item.College,
				&item.Phone,
				&item.Email,
				&item.ApplyStatus,
				&item.SignInTime,
				&item.StudyHours,
				&item.AchievementStatus,
				&item.AchievementURL,
				&item.ApplyTime,
			); err != nil {
				response.Fail(c, http.StatusInternalServerError, "training record scan failed")
				return
			}
			items = append(items, item.toJSON())
		}
		if err := rows.Err(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "training record rows failed")
			return
		}

		response.OK(c, pageData(page, size, total, items))
	}
}

func audit(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		var payload map[string]any
		if err := c.ShouldBindJSON(&payload); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid audit payload")
			return
		}
		req := parseAudit(payload)
		if req.TeacherID == 0 {
			response.Fail(c, http.StatusBadRequest, "teacherId is required")
			return
		}

		if req.HasApplyStatus {
			if _, err := db.ExecContext(c.Request.Context(), `UPDATE training_record SET apply_status = ? WHERE training_id = ? AND teacher_id = ?`, req.ApplyStatus, pathID(c), req.TeacherID); err != nil {
				response.Fail(c, http.StatusInternalServerError, "training apply audit failed")
				return
			}
		}
		if req.HasAchievementStat {
			if _, err := db.ExecContext(c.Request.Context(), `UPDATE training_record SET achievement_status = ? WHERE training_id = ? AND teacher_id = ?`, req.AchievementStatus, pathID(c), req.TeacherID); err != nil {
				response.Fail(c, http.StatusInternalServerError, "training achievement audit failed")
				return
			}
		}
		response.OK(c, gin.H{"trainingId": pathID(c), "audited": true})
	}
}

func submitAchievement(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req achievementRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.AchievementURL == "" {
			response.Fail(c, http.StatusBadRequest, "achievementUrl is required")
			return
		}
		_, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE training_record
			 SET achievement_url = ?, achievement_status = 0
			 WHERE training_id = ? AND teacher_id = ?`,
			req.AchievementURL, pathID(c), c.GetInt64("userID"),
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "training achievement submit failed")
			return
		}
		response.OK(c, gin.H{"trainingId": pathID(c), "submitted": true})
	}
}

func records(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 10)
		offset := (page - 1) * size
		userID := c.GetInt64("userID")

		var total int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM training_record WHERE teacher_id = ?`, userID).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "ledger count failed")
			return
		}

		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT r.id, r.training_id, t.title, COALESCE(DATE_FORMAT(r.sign_in_time, '%Y-%m-%d %H:%i:%s'), ''),
			 r.study_hours, r.apply_status, r.achievement_status, COALESCE(r.achievement_url, '')
			 FROM training_record r
			 JOIN training t ON t.id = r.training_id
			 WHERE r.teacher_id = ?
			 ORDER BY r.create_time DESC
			 LIMIT ? OFFSET ?`,
			userID, size, offset,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "ledger query failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var item ledgerItem
			if err := rows.Scan(&item.ID, &item.TrainingID, &item.Title, &item.CompletedAt, &item.LearningHour, &item.ApplyStatus, &item.AchievementStatus, &item.AchievementURL); err != nil {
				response.Fail(c, http.StatusInternalServerError, "ledger scan failed")
				return
			}
			items = append(items, item.toJSON())
		}
		if err := rows.Err(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "ledger rows failed")
			return
		}

		response.OK(c, pageData(page, size, total, items))
	}
}

func statistics(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		where, args, err := listScope(c, db)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "user scope query failed")
			return
		}
		var open, completed int
		var hours float64
		_ = db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM training `+where+statusClause(where, "status IN (1, 2)"), args...).Scan(&open)
		_ = db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM training `+where+statusClause(where, "status IN (3, 4)"), args...).Scan(&completed)
		_ = db.QueryRowContext(c.Request.Context(), `SELECT COALESCE(SUM(study_hours), 0) FROM training_record`).Scan(&hours)
		response.OK(c, gin.H{"open": open, "completed": completed, "hours": hours})
	}
}

type trainingItem struct {
	ID                 int64
	Title              string
	Type               string
	Level              int
	Location           string
	Quota              int
	Status             int
	StartTime          string
	EndTime            string
	SponsorUnit        string
	OrganizerUnit      string
	Requirements       string
	AchievementRequire string
	EnrolledCount      int
	ApplyStatus        int
}

func (i trainingItem) toJSON() gin.H {
	return gin.H{
		"id":                 i.ID,
		"title":              i.Title,
		"type":               i.Type,
		"level":              i.Level,
		"location":           i.Location,
		"quota":              i.Quota,
		"hours":              0,
		"status":             i.Status,
		"statusText":         statusText(i.Status),
		"startTime":          i.StartTime,
		"endTime":            i.EndTime,
		"sponsorUnit":        i.SponsorUnit,
		"organizerUnit":      i.OrganizerUnit,
		"requirements":       i.Requirements,
		"achievementRequire": i.AchievementRequire,
		"enrolledCount":      i.EnrolledCount,
		"enrolled":           i.ApplyStatus >= 0,
		"applyStatus":        i.ApplyStatus,
	}
}

type ledgerItem struct {
	ID                int64
	TrainingID        int64
	Title             string
	CompletedAt       string
	LearningHour      float64
	ApplyStatus       int
	AchievementStatus int
	AchievementURL    string
}

type trainingRecordItem struct {
	ID                int64
	TrainingID        int64
	TeacherID         int64
	TeacherName       string
	EmployeeNo        string
	Department        string
	College           string
	Phone             string
	Email             string
	ApplyStatus       int
	SignInTime        string
	StudyHours        float64
	AchievementStatus int
	AchievementURL    string
	ApplyTime         string
}

func (i trainingRecordItem) toJSON() gin.H {
	return gin.H{
		"id":                    i.ID,
		"trainingId":            i.TrainingID,
		"teacherId":             i.TeacherID,
		"teacherName":           i.TeacherName,
		"employeeNo":            i.EmployeeNo,
		"department":            i.Department,
		"college":               i.College,
		"phone":                 i.Phone,
		"email":                 i.Email,
		"applyStatus":           i.ApplyStatus,
		"applyStatusText":       applyStatusText(i.ApplyStatus),
		"signInTime":            i.SignInTime,
		"studyHours":            i.StudyHours,
		"achievementStatus":     i.AchievementStatus,
		"achievementStatusText": applyStatusText(i.AchievementStatus),
		"achievementUrl":        i.AchievementURL,
		"applyTime":             i.ApplyTime,
	}
}

func (i ledgerItem) toJSON() gin.H {
	return gin.H{
		"id":                i.ID,
		"trainingId":        i.TrainingID,
		"title":             i.Title,
		"completedAt":       i.CompletedAt,
		"learningHour":      i.LearningHour,
		"applyStatus":       i.ApplyStatus,
		"achievementStatus": i.AchievementStatus,
		"achievementUrl":    i.AchievementURL,
	}
}

func parseAudit(payload map[string]any) auditRequest {
	req := auditRequest{}
	if value, ok := payload["teacherId"].(float64); ok {
		req.TeacherID = int64(value)
	}
	if value, ok := payload["applyStatus"].(float64); ok {
		req.ApplyStatus = int(value)
		req.HasApplyStatus = true
	}
	if value, ok := payload["achievementStatus"].(float64); ok {
		req.AchievementStatus = int(value)
		req.HasAchievementStat = true
	}
	return req
}

func pageData(page int, size int, total int, list []gin.H) gin.H {
	pages := 0
	if size > 0 {
		pages = (total + size - 1) / size
	}
	return gin.H{"page": page, "size": size, "total": total, "pages": pages, "list": list}
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func pathID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
}

func statusText(status int) string {
	switch status {
	case 0:
		return "draft"
	case 1:
		return "open"
	case 2:
		return "in_progress"
	case 3:
		return "ended"
	case 4:
		return "archived"
	default:
		return "unknown"
	}
}

func applyStatusText(status int) string {
	switch status {
	case 0:
		return "pending"
	case 1:
		return "approved"
	case 2:
		return "rejected"
	default:
		return "unknown"
	}
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func isAdmin(c *gin.Context) bool {
	role := c.GetString("role")
	return role == "party_admin" || role == "school_admin"
}

func listScope(_ *gin.Context, _ *sql.DB) (string, []any, error) {
	return "", nil, nil
}

func statusClause(where string, clause string) string {
	return appendWhere(where, clause)
}

func appendWhere(where string, clause string) string {
	if where == "" {
		return "WHERE " + clause
	}
	return " AND " + clause
}

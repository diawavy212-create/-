package treehole

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"teacher-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

type submitRequest struct {
	AnonymousType  int    `json:"anonymousType"`
	Category       string `json:"category"`
	SubCategory    string `json:"subCategory"`
	InfluenceScope int    `json:"influenceScope"`
	EmergencyLevel int    `json:"emergencyLevel"`
	Description    string `json:"description"`
	ExpectedMethod int    `json:"expectedMethod"`
	ContactWay     string `json:"contactWay"`
	AttachmentURL  string `json:"attachmentUrl"`
	Content        string `json:"content"`
}

type assignRequest struct {
	HandlerUnit string `json:"handlerUnit"`
	HandlerID   int64  `json:"handlerId"`
}

type feedbackRequest struct {
	HandlerUnit   string `json:"handlerUnit"`
	HandleContent string `json:"handleContent"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("", list(db))
	rg.POST("", submit(db))
	rg.POST("/uploads", uploadAttachment())
	rg.POST("/:id/accept", accept(db))
	rg.POST("/:id/assign", assign(db))
	rg.POST("/:id/feedback", feedback(db))
	rg.POST("/:id/complete", complete(db))
	rg.POST("/:id/satisfaction", evaluate(db))
	rg.DELETE("/:id", remove(db))
	rg.GET("/statistics", statistics(db))
}

func submit(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req submitRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid treehole payload")
			return
		}
		if req.Description == "" {
			req.Description = req.Content
		}
		req.SubCategory = strings.TrimSpace(req.SubCategory)
		if req.Description == "" {
			response.Fail(c, http.StatusBadRequest, "description is required")
			return
		}
		if req.Category == "" {
			req.Category = "general"
		}

		teacherID := c.GetInt64("userID")

		result, err := db.ExecContext(
			c.Request.Context(),
			`INSERT INTO appeal (
				teacher_id, anonymous_type, category, sub_category, influence_scope,
				emergency_level, description, expected_method, contact_way, attachment_url, status
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
			teacherID, req.AnonymousType, req.Category, nullString(req.SubCategory), req.InfluenceScope,
			req.EmergencyLevel, req.Description, req.ExpectedMethod, nullString(req.ContactWay), nullString(req.AttachmentURL),
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole submit failed")
			return
		}
		id, _ := result.LastInsertId()
		response.OK(c, gin.H{"id": id})
	}
}

func uploadAttachment() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "file is required")
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			response.Fail(c, http.StatusBadRequest, "only image attachment is allowed")
			return
		}
		dir := filepath.Join("uploads", "treeholes")
		if err := os.MkdirAll(dir, 0755); err != nil {
			response.Fail(c, http.StatusInternalServerError, "upload directory create failed")
			return
		}
		name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		dst := filepath.Join(dir, name)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			response.Fail(c, http.StatusInternalServerError, "attachment upload failed")
			return
		}
		response.OK(c, gin.H{"url": "/uploads/treeholes/" + name})
	}
}

func list(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		current, err := currentUser(c, db, userID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "user query failed")
			return
		}

		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 10)
		offset := (page - 1) * size

		where := ""
		args := []any{}
		switch current.Role {
		case "teacher":
			where = "WHERE a.teacher_id = ?"
			args = append(args, userID)
		}

		var total int
		countArgs := append([]any{}, args...)
		if err := db.QueryRowContext(
			c.Request.Context(),
			"SELECT COUNT(*) FROM appeal a LEFT JOIN teacher t ON t.id = a.teacher_id "+where,
			countArgs...,
		).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole count failed")
			return
		}

		args = append(args, size, offset)
		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT a.id, COALESCE(t.name, 'anonymous'), COALESCE(t.college, ''), a.category,
			 COALESCE(a.sub_category, ''), a.description, a.status, COALESCE(a.handle_content, ''),
			 COALESCE(a.attachment_url, ''), a.anonymous_type, a.emergency_level, a.influence_scope,
			 COALESCE(a.satisfaction, -1)
			 FROM appeal a
			 LEFT JOIN teacher t ON t.id = a.teacher_id `+where+`
			 ORDER BY a.create_time DESC
			 LIMIT ? OFFSET ?`,
			args...,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole list failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var item treeholeItem
			if err := rows.Scan(
				&item.ID, &item.TeacherName, &item.College, &item.Category, &item.SubCategory,
				&item.Description, &item.Status, &item.HandleContent, &item.AttachmentURL,
				&item.AnonymousType, &item.EmergencyLevel, &item.InfluenceScope, &item.Satisfaction,
			); err != nil {
				response.Fail(c, http.StatusInternalServerError, "treehole scan failed")
				return
			}
			items = append(items, item.toJSON())
		}
		if err := rows.Err(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole rows failed")
			return
		}

		response.OK(c, pageData(page, size, total, items))
	}
}

func accept(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := db.ExecContext(c.Request.Context(), `UPDATE appeal SET status = 1 WHERE id = ?`, pathID(c)); err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole accept failed")
			return
		}
		response.OK(c, gin.H{"id": pathID(c), "accepted": true})
	}
}

func assign(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req assignRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.HandlerUnit == "" {
			response.Fail(c, http.StatusBadRequest, "handlerUnit is required")
			return
		}
		if _, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE appeal SET handler_unit = ?, handler_id = ?, status = 1 WHERE id = ?`,
			req.HandlerUnit, nullInt64(req.HandlerID), pathID(c),
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole assign failed")
			return
		}
		response.OK(c, gin.H{"id": pathID(c), "assigned": true})
	}
}

func feedback(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req feedbackRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.HandleContent == "" {
			response.Fail(c, http.StatusBadRequest, "handleContent is required")
			return
		}
		if _, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE appeal SET handler_unit = COALESCE(NULLIF(?, ''), handler_unit), handle_content = ?, status = 2 WHERE id = ?`,
			req.HandlerUnit, req.HandleContent, pathID(c),
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole feedback failed")
			return
		}
		response.OK(c, gin.H{"id": pathID(c), "feedbackSaved": true})
	}
}

func complete(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := db.ExecContext(c.Request.Context(), `UPDATE appeal SET status = 4 WHERE id = ?`, pathID(c)); err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole complete failed")
			return
		}
		response.OK(c, gin.H{"id": pathID(c), "completed": true})
	}
}

func remove(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		appealID := pathID(c)
		var attachmentURL string
		_ = db.QueryRowContext(c.Request.Context(), `SELECT COALESCE(attachment_url, '') FROM appeal WHERE id = ?`, appealID).Scan(&attachmentURL)
		result, err := db.ExecContext(c.Request.Context(), `DELETE FROM appeal WHERE id = ?`, appealID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole delete failed")
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			response.Fail(c, http.StatusNotFound, "treehole not found")
			return
		}
		removeAttachmentFile(attachmentURL)
		response.OK(c, gin.H{"id": appealID, "deleted": true})
	}
}

func removeAttachmentFile(url string) {
	if url == "" || strings.Contains(url, "://") {
		return
	}
	cleanURL := strings.TrimPrefix(filepath.ToSlash(url), "/")
	if !strings.HasPrefix(cleanURL, "uploads/treeholes/") {
		return
	}
	root, err := filepath.Abs(filepath.Join("uploads", "treeholes"))
	if err != nil {
		return
	}
	target, err := filepath.Abs(filepath.FromSlash(cleanURL))
	if err != nil {
		return
	}
	if target == root || !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return
	}
	_ = os.Remove(target)
}

func evaluate(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload map[string]any
		if err := c.ShouldBindJSON(&payload); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid satisfaction payload")
			return
		}
		satisfaction, ok := parseSatisfaction(payload)
		if !ok {
			response.Fail(c, http.StatusBadRequest, "satisfaction is required")
			return
		}
		if satisfaction < 0 || satisfaction > 3 {
			response.Fail(c, http.StatusBadRequest, "satisfaction must be 0-3")
			return
		}
		appealID := pathID(c)
		_, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE appeal SET satisfaction = ?, status = 3 WHERE id = ?`,
			satisfaction, appealID,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole satisfaction failed")
			return
		}

		var saved int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COALESCE(satisfaction, -1) FROM appeal WHERE id = ?`, appealID).Scan(&saved); err != nil {
			response.Fail(c, http.StatusNotFound, "treehole not found")
			return
		}
		if saved < 0 || saved > 3 {
			response.Fail(c, http.StatusInternalServerError, "treehole satisfaction was not saved")
			return
		}
		response.OK(c, gin.H{
			"id":               appealID,
			"evaluated":        true,
			"satisfaction":     saved,
			"satisfactionText": satisfactionText(saved),
		})
	}
}

func parseSatisfaction(payload map[string]any) (int, bool) {
	keys := []string{"satisfaction", "satisfactionScore", "score", "rating"}
	for _, key := range keys {
		value, exists := payload[key]
		if !exists {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed), true
		case int:
			return typed, true
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(typed))
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func statistics(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.QueryContext(c.Request.Context(), `SELECT status, COUNT(*) FROM appeal GROUP BY status`)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "treehole statistics failed")
			return
		}
		defer rows.Close()

		stats := gin.H{"pending": 0, "processing": 0, "feedback": 0, "evaluated": 0, "archived": 0}
		for rows.Next() {
			var status int
			var count int
			if err := rows.Scan(&status, &count); err != nil {
				response.Fail(c, http.StatusInternalServerError, "treehole statistics scan failed")
				return
			}
			switch status {
			case 0:
				stats["pending"] = count
			case 1:
				stats["processing"] = count
			case 2:
				stats["feedback"] = count
			case 3:
				stats["evaluated"] = count
			case 4:
				stats["archived"] = count
			}
		}
		response.OK(c, stats)
	}
}

type treeholeItem struct {
	ID             int64
	TeacherName    string
	College        string
	Category       string
	SubCategory    string
	Description    string
	Status         int
	HandleContent  string
	AttachmentURL  string
	AnonymousType  int
	EmergencyLevel int
	InfluenceScope int
	Satisfaction   int
}

func (i treeholeItem) toJSON() gin.H {
	teacherName := i.TeacherName
	if i.AnonymousType != 0 {
		teacherName = "匿名"
	}
	title := i.SubCategory
	if title == "" {
		title = i.Description
		if len([]rune(title)) > 24 {
			title = string([]rune(title)[:24]) + "..."
		}
	}
	return gin.H{
		"id":               i.ID,
		"title":            title,
		"teacherName":      teacherName,
		"college":          i.College,
		"category":         i.Category,
		"subCategory":      i.SubCategory,
		"description":      i.Description,
		"content":          i.Description,
		"status":           i.Status,
		"statusText":       statusText(i.Status),
		"handleContent":    i.HandleContent,
		"attachmentUrl":    i.AttachmentURL,
		"anonymousType":    i.AnonymousType,
		"anonymousText":    anonymousText(i.AnonymousType),
		"emergencyLevel":   i.EmergencyLevel,
		"emergencyText":    emergencyText(i.EmergencyLevel),
		"influenceScope":   i.InfluenceScope,
		"evaluated":        i.Status == 3,
		"satisfaction":     i.Satisfaction,
		"satisfactionText": satisfactionText(i.Satisfaction),
	}
}

type currentUserInfo struct {
	Role    string
	College string
}

func currentUser(c *gin.Context, db *sql.DB, userID int64) (currentUserInfo, error) {
	var user currentUserInfo
	err := db.QueryRowContext(c.Request.Context(), `SELECT role, college FROM teacher WHERE id = ?`, userID).Scan(&user.Role, &user.College)
	return user, err
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
		return "待受理"
	case 1:
		return "处理中"
	case 2:
		return "已反馈"
	case 3:
		return "已评价"
	case 4:
		return "已处理"
	default:
		return "未知"
	}
}

func anonymousText(value int) string {
	switch value {
	case 0:
		return "实名"
	case 1:
		return "匿名"
	case 2:
		return "匿名可回访"
	default:
		return "未知"
	}
}

func emergencyText(value int) string {
	switch value {
	case 0:
		return "普通"
	case 1:
		return "较急"
	case 2:
		return "紧急"
	default:
		return "未知"
	}
}

func satisfactionText(value int) string {
	switch value {
	case 0:
		return "不满意"
	case 1:
		return "基本满意"
	case 2:
		return "满意"
	case 3:
		return "非常满意"
	default:
		return "未评价"
	}
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

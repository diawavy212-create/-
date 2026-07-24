package survey

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"teacher-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

type createRequest struct {
	Title     string            `json:"title"`
	Type      int               `json:"type"`
	Scope     string            `json:"scope"`
	College   string            `json:"college"`
	Group     string            `json:"group"`
	StartTime string            `json:"startTime"`
	EndTime   string            `json:"endTime"`
	Status    int               `json:"status"`
	Questions []questionRequest `json:"questions"`
}

type questionRequest struct {
	ID       int64           `json:"id"`
	Title    string          `json:"title"`
	Type     string          `json:"type"`
	Required bool            `json:"required"`
	Options  []optionRequest `json:"options"`
}

type optionRequest struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
	Score int    `json:"score"`
}

type submitRequest struct {
	SurveyID        int64           `json:"surveyId"`
	DurationSeconds int             `json:"durationSeconds"`
	Answers         []answerRequest `json:"answers"`
}

type answerRequest struct {
	QuestionID int64  `json:"questionId"`
	OptionID   int64  `json:"optionId"`
	Content    string `json:"content"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	ensureSchema(db)
	rg.GET("/list", list(db))
	rg.POST("/answer/submit", submit(db))
	rg.GET("/records", records(db))
	rg.POST("", create(db))
	rg.PUT("/:id", update(db))
	rg.DELETE("/:id", remove(db))
	rg.POST("/create", create(db))
	rg.GET("/report/:id", report(db))
	rg.GET("/:id/report", report(db))
}

func list(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 10)
		offset := (page - 1) * size
		where, args, err := scopeWhere(c, db)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey scope query failed")
			return
		}
		if raw := c.Query("status"); raw != "" {
			status, err := strconv.Atoi(raw)
			if err != nil {
				response.Fail(c, http.StatusBadRequest, "invalid status")
				return
			}
			where = appendWhere(where, "s.status = ?")
			args = append(args, status)
		}

		var total int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM survey s `+where, args...).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey count failed")
			return
		}

		queryArgs := append([]any{c.GetInt64("userID")}, args...)
		queryArgs = append(queryArgs, size, offset)
		groupColumn := surveyGroupColumn(c, db)
		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT s.id, s.title, s.type, s.scope, COALESCE(s.college, ''), COALESCE(s.`+groupColumn+`, ''),
			 COALESCE(DATE_FORMAT(s.start_time, '%Y-%m-%d %H:%i:%s'), ''), COALESCE(DATE_FORMAT(s.end_time, '%Y-%m-%d %H:%i:%s'), ''),
			 s.status, COALESCE(r.is_valid, -1), COALESCE(DATE_FORMAT(r.submit_time, '%Y-%m-%d %H:%i:%s'), ''),
			 (SELECT COUNT(*) FROM survey_question q WHERE q.survey_id = s.id)
			 FROM survey s
			 LEFT JOIN survey_response r ON r.survey_id = s.id AND r.teacher_id = ? `+where+`
			 ORDER BY s.start_time DESC, s.id DESC
			 LIMIT ? OFFSET ?`,
			queryArgs...,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey list failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var item surveyItem
			if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.Scope, &item.College, &item.Group, &item.StartTime, &item.EndTime, &item.Status, &item.ValidState, &item.SubmitTime, &item.QuestionCount); err != nil {
				response.Fail(c, http.StatusInternalServerError, "survey scan failed")
				return
			}
			questions, err := loadQuestions(c, db, item.ID)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, "survey questions failed")
				return
			}
			items = append(items, item.toJSON(questions))
		}
		if err := rows.Err(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey rows failed")
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
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
			response.Fail(c, http.StatusBadRequest, "title is required")
			return
		}
		if len(req.Questions) == 0 {
			req.Questions = defaultQuestions()
		}

		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey create failed")
			return
		}
		defer tx.Rollback()

		groupColumn := surveyGroupColumn(c, db)
		result, err := tx.ExecContext(
			c.Request.Context(),
			`INSERT INTO survey (title, type, scope, college, `+groupColumn+`, start_time, end_time, status, create_by)
			 VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)`,
			req.Title, req.Type, defaultString(req.Scope, "全校"), nullString(req.College), nullString(req.Group), req.StartTime, req.EndTime, req.Status, c.GetInt64("userID"),
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey create failed")
			return
		}
		id, _ := result.LastInsertId()
		if err := saveQuestions(c, tx, id, req.Questions); err != nil {
			response.Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		if err := tx.Commit(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey create failed")
			return
		}
		response.OK(c, gin.H{"surveyId": id})
	}
}

func update(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
			response.Fail(c, http.StatusBadRequest, "title is required")
			return
		}
		surveyID := pathID(c)
		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey update failed")
			return
		}
		defer tx.Rollback()

		groupColumn := surveyGroupColumn(c, db)
		if _, err := tx.ExecContext(
			c.Request.Context(),
			`UPDATE survey SET title = ?, type = ?, scope = ?, college = ?, `+groupColumn+` = ?,
			 start_time = NULLIF(?, ''), end_time = NULLIF(?, ''), status = ?
			 WHERE id = ?`,
			req.Title, req.Type, defaultString(req.Scope, "全校"), nullString(req.College), nullString(req.Group), req.StartTime, req.EndTime, req.Status, surveyID,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey update failed")
			return
		}
		if len(req.Questions) > 0 {
			if _, err := tx.ExecContext(c.Request.Context(), `DELETE FROM survey_option WHERE question_id IN (SELECT id FROM survey_question WHERE survey_id = ?)`, surveyID); err != nil {
				response.Fail(c, http.StatusInternalServerError, "survey options reset failed")
				return
			}
			if _, err := tx.ExecContext(c.Request.Context(), `DELETE FROM survey_question WHERE survey_id = ?`, surveyID); err != nil {
				response.Fail(c, http.StatusInternalServerError, "survey questions reset failed")
				return
			}
			if err := saveQuestions(c, tx, surveyID, req.Questions); err != nil {
				response.Fail(c, http.StatusInternalServerError, err.Error())
				return
			}
		}
		if err := tx.Commit(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey update failed")
			return
		}
		response.OK(c, gin.H{"surveyId": surveyID, "updated": true})
	}
}

func remove(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		result, err := db.ExecContext(c.Request.Context(), `DELETE FROM survey WHERE id = ?`, pathID(c))
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey delete failed")
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			response.Fail(c, http.StatusNotFound, "survey not found")
			return
		}
		response.OK(c, gin.H{"surveyId": pathID(c), "deleted": true})
	}
}

func submit(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req submitRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.SurveyID == 0 {
			response.Fail(c, http.StatusBadRequest, "surveyId is required")
			return
		}
		if len(req.Answers) == 0 {
			response.Fail(c, http.StatusBadRequest, "answers are required")
			return
		}

		var active int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM survey WHERE id = ? AND status = 1 AND (end_time IS NULL OR end_time >= NOW())`, req.SurveyID).Scan(&active); err != nil || active == 0 {
			response.Fail(c, http.StatusBadRequest, "survey is not open")
			return
		}

		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey submit failed")
			return
		}
		defer tx.Rollback()

		valid := req.DurationSeconds >= 10
		result, err := tx.ExecContext(
			c.Request.Context(),
			`INSERT INTO survey_response (survey_id, teacher_id, duration_seconds, is_valid)
			 VALUES (?, ?, ?, ?)`,
			req.SurveyID, c.GetInt64("userID"), req.DurationSeconds, valid,
		)
		if err != nil {
			response.Fail(c, http.StatusConflict, "survey already submitted")
			return
		}
		responseID, _ := result.LastInsertId()
		for _, ans := range req.Answers {
			if ans.QuestionID == 0 {
				continue
			}
			if _, err := tx.ExecContext(
				c.Request.Context(),
				`INSERT INTO survey_answer (response_id, survey_id, question_id, option_id, content)
				 VALUES (?, ?, ?, ?, ?)`,
				responseID, req.SurveyID, ans.QuestionID, nullInt64(ans.OptionID), nullString(strings.TrimSpace(ans.Content)),
			); err != nil {
				response.Fail(c, http.StatusInternalServerError, "survey answer save failed")
				return
			}
		}
		if err := tx.Commit(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey submit failed")
			return
		}
		response.OK(c, gin.H{"surveyId": req.SurveyID, "submitted": true, "valid": valid})
	}
}

func records(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 10)
		offset := (page - 1) * size
		userID := c.GetInt64("userID")

		var total int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM survey_response WHERE teacher_id = ?`, userID).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey record count failed")
			return
		}
		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT r.id, r.survey_id, s.title, r.is_valid, r.duration_seconds, DATE_FORMAT(r.submit_time, '%Y-%m-%d %H:%i:%s')
			 FROM survey_response r JOIN survey s ON s.id = r.survey_id
			 WHERE r.teacher_id = ?
			 ORDER BY r.submit_time DESC
			 LIMIT ? OFFSET ?`,
			userID, size, offset,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "survey records failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var id, surveyID int64
			var title, submitTime string
			var valid bool
			var duration int
			if err := rows.Scan(&id, &surveyID, &title, &valid, &duration, &submitTime); err != nil {
				response.Fail(c, http.StatusInternalServerError, "survey record scan failed")
				return
			}
			items = append(items, gin.H{"id": id, "surveyId": surveyID, "title": title, "valid": valid, "durationSeconds": duration, "submitTime": submitTime})
		}
		response.OK(c, pageData(page, size, total, items))
	}
}

func report(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		surveyID := pathID(c)
		var title string
		if err := db.QueryRowContext(c.Request.Context(), `SELECT title FROM survey WHERE id = ?`, surveyID).Scan(&title); err != nil {
			response.Fail(c, http.StatusNotFound, "survey not found")
			return
		}

		targetTotal := teacherCount(c, db)
		validTotal := scalarInt(c, db, `SELECT COUNT(*) FROM survey_response WHERE survey_id = ? AND is_valid = 1`, surveyID)
		invalidTotal := scalarInt(c, db, `SELECT COUNT(*) FROM survey_response WHERE survey_id = ? AND is_valid = 0`, surveyID)
		optionStats := loadOptionStats(c, db, surveyID)
		collegeCompare := loadCollegeCompare(c, db, surveyID)
		openTopics := loadOpenTopics(c, db, surveyID)
		riskList := buildRiskList(targetTotal, validTotal, invalidTotal, optionStats, openTopics)

		response.OK(c, gin.H{
			"surveyId":          surveyID,
			"title":             title,
			"targetTotal":       targetTotal,
			"validTotal":        validTotal,
			"invalidTotal":      invalidTotal,
			"participationRate": percent(validTotal, targetTotal),
			"optionStats":       optionStats,
			"collegeCompare":    collegeCompare,
			"trend":             loadTrend(c, db, surveyID),
			"openTopics":        openTopics,
			"riskList":          riskList,
			"summary":           buildSummary(title, validTotal, targetTotal, riskList),
		})
	}
}

func saveQuestions(c *gin.Context, tx *sql.Tx, surveyID int64, questions []questionRequest) error {
	for index, q := range questions {
		q.Title = strings.TrimSpace(q.Title)
		if q.Title == "" {
			return errText("survey question title is required")
		}
		if q.Type == "" {
			q.Type = "single"
		}
		result, err := tx.ExecContext(
			c.Request.Context(),
			`INSERT INTO survey_question (survey_id, title, question_type, required, sort_order)
			 VALUES (?, ?, ?, ?, ?)`,
			surveyID, q.Title, q.Type, q.Required, index+1,
		)
		if err != nil {
			return err
		}
		questionID, _ := result.LastInsertId()
		for optionIndex, opt := range q.Options {
			if strings.TrimSpace(opt.Label) == "" {
				continue
			}
			if _, err := tx.ExecContext(
				c.Request.Context(),
				`INSERT INTO survey_option (question_id, label, score, sort_order)
				 VALUES (?, ?, ?, ?)`,
				questionID, opt.Label, opt.Score, optionIndex+1,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadQuestions(c *gin.Context, db *sql.DB, surveyID int64) ([]gin.H, error) {
	rows, err := db.QueryContext(
		c.Request.Context(),
		`SELECT id, title, question_type, required FROM survey_question WHERE survey_id = ? ORDER BY sort_order, id`,
		surveyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]gin.H, 0)
	for rows.Next() {
		var id int64
		var title, qtype string
		var required bool
		if err := rows.Scan(&id, &title, &qtype, &required); err != nil {
			return nil, err
		}
		options, err := loadOptions(c, db, id)
		if err != nil {
			return nil, err
		}
		questions = append(questions, gin.H{"id": id, "title": title, "type": qtype, "required": required, "options": options})
	}
	return questions, rows.Err()
}

func loadOptions(c *gin.Context, db *sql.DB, questionID int64) ([]gin.H, error) {
	rows, err := db.QueryContext(c.Request.Context(), `SELECT id, label, score FROM survey_option WHERE question_id = ? ORDER BY sort_order, id`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]gin.H, 0)
	for rows.Next() {
		var id int64
		var label string
		var score int
		if err := rows.Scan(&id, &label, &score); err != nil {
			return nil, err
		}
		options = append(options, gin.H{"id": id, "label": label, "score": score})
	}
	return options, rows.Err()
}

type surveyItem struct {
	ID            int64
	Title         string
	Type          int
	Scope         string
	College       string
	Group         string
	StartTime     string
	EndTime       string
	Status        int
	ValidState    int
	SubmitTime    string
	QuestionCount int
}

func (i surveyItem) toJSON(questions []gin.H) gin.H {
	return gin.H{
		"id":            i.ID,
		"title":         i.Title,
		"type":          i.Type,
		"typeText":      typeText(i.Type),
		"scope":         i.Scope,
		"college":       i.College,
		"group":         i.Group,
		"startTime":     i.StartTime,
		"endTime":       i.EndTime,
		"status":        i.Status,
		"statusText":    statusText(i.Status),
		"submitted":     i.ValidState >= 0,
		"valid":         i.ValidState == 1,
		"submitTime":    i.SubmitTime,
		"questionCount": i.QuestionCount,
		"questions":     questions,
	}
}

func loadOptionStats(c *gin.Context, db *sql.DB, surveyID int64) []gin.H {
	rows, err := db.QueryContext(
		c.Request.Context(),
		`SELECT q.title, o.label, COUNT(r.id) AS count
		 FROM survey_question q
		 JOIN survey_option o ON o.question_id = q.id
		 LEFT JOIN survey_answer a ON a.option_id = o.id
		 LEFT JOIN survey_response r ON r.id = a.response_id AND r.is_valid = 1
		 WHERE q.survey_id = ?
		 GROUP BY q.id, o.id
		 ORDER BY q.sort_order, o.sort_order`,
		surveyID,
	)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()

	items := make([]gin.H, 0)
	for rows.Next() {
		var question, option string
		var count int
		if err := rows.Scan(&question, &option, &count); err == nil {
			items = append(items, gin.H{"question": question, "option": option, "count": count})
		}
	}
	return items
}

func loadCollegeCompare(c *gin.Context, db *sql.DB, surveyID int64) []gin.H {
	rows, err := db.QueryContext(
		c.Request.Context(),
		`SELECT t.college, COUNT(r.id)
		 FROM teacher t
		 LEFT JOIN survey_response r ON r.teacher_id = t.id AND r.survey_id = ? AND r.is_valid = 1
		 WHERE t.role = 'teacher'
		 GROUP BY t.college
		 ORDER BY COUNT(r.id) DESC, t.college`,
		surveyID,
	)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()

	items := make([]gin.H, 0)
	for rows.Next() {
		var college string
		var count int
		if err := rows.Scan(&college, &count); err == nil {
			items = append(items, gin.H{"college": college, "count": count})
		}
	}
	return items
}

func loadTrend(c *gin.Context, db *sql.DB, surveyID int64) []gin.H {
	rows, err := db.QueryContext(
		c.Request.Context(),
		`SELECT DATE_FORMAT(submit_time, '%Y-%m-%d'), COUNT(*)
		 FROM survey_response
		 WHERE survey_id = ? AND is_valid = 1
		 GROUP BY DATE_FORMAT(submit_time, '%Y-%m-%d')
		 ORDER BY DATE_FORMAT(submit_time, '%Y-%m-%d')`,
		surveyID,
	)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()

	items := make([]gin.H, 0)
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err == nil {
			items = append(items, gin.H{"date": date, "count": count})
		}
	}
	return items
}

func loadOpenTopics(c *gin.Context, db *sql.DB, surveyID int64) []gin.H {
	rows, err := db.QueryContext(
		c.Request.Context(),
		`SELECT a.content
		 FROM survey_answer a
		 JOIN survey_response r ON r.id = a.response_id AND r.is_valid = 1
		 WHERE a.survey_id = ? AND a.content IS NOT NULL AND a.content <> ''`,
		surveyID,
	)
	if err != nil {
		return []gin.H{}
	}
	defer rows.Close()

	counter := map[string]int{}
	stop := map[string]bool{"的": true, "了": true, "和": true, "是": true, "建议": true, "希望": true, "加强": true, "工作": true}
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			continue
		}
		for _, token := range strings.FieldsFunc(content, func(r rune) bool {
			return r == ' ' || r == ',' || r == '，' || r == '.' || r == '。' || r == ';' || r == '；' || r == '\n' || r == '\r' || r == '、'
		}) {
			token = strings.TrimSpace(token)
			if len([]rune(token)) < 2 || stop[token] {
				continue
			}
			counter[token]++
		}
	}
	type pair struct {
		Word  string
		Count int
	}
	pairs := make([]pair, 0, len(counter))
	for word, count := range counter {
		pairs = append(pairs, pair{Word: word, Count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count == pairs[j].Count {
			return pairs[i].Word < pairs[j].Word
		}
		return pairs[i].Count > pairs[j].Count
	})
	items := make([]gin.H, 0)
	for i, p := range pairs {
		if i >= 8 {
			break
		}
		items = append(items, gin.H{"topic": p.Word, "count": p.Count})
	}
	return items
}

func buildRiskList(targetTotal int, validTotal int, invalidTotal int, optionStats []gin.H, topics []gin.H) []gin.H {
	risks := make([]gin.H, 0)
	if percent(validTotal, targetTotal) < 60 {
		risks = append(risks, gin.H{"level": "high", "content": "参与率低于60%，建议继续提醒未填教师。"})
	}
	if invalidTotal > 0 {
		risks = append(risks, gin.H{"level": "medium", "content": "存在完成时长过短等无效问卷，统计时已自动过滤。"})
	}
	for _, item := range optionStats {
		option, _ := item["option"].(string)
		count, _ := item["count"].(int)
		if count > 0 && (strings.Contains(option, "压力") || strings.Contains(option, "不满意") || strings.Contains(option, "较差")) {
			risks = append(risks, gin.H{"level": "medium", "content": "风险选项出现集中反馈：" + option})
		}
	}
	for _, item := range topics {
		topic, _ := item["topic"].(string)
		if strings.Contains(topic, "压力") || strings.Contains(topic, "负担") || strings.Contains(topic, "焦虑") {
			risks = append(risks, gin.H{"level": "medium", "content": "开放题高频主题需关注：" + topic})
		}
	}
	if len(risks) == 0 {
		risks = append(risks, gin.H{"level": "normal", "content": "暂未发现明显集中风险，建议持续跟踪趋势。"})
	}
	return risks
}

func buildSummary(title string, valid int, target int, risks []gin.H) string {
	payload, _ := json.Marshal(risks)
	return title + " 已完成有效问卷 " + strconv.Itoa(valid) + " 份，参与率 " + strconv.Itoa(percent(valid, target)) + "%，风险提示：" + string(payload)
}

func teacherCount(c *gin.Context, db *sql.DB) int {
	return scalarInt(c, db, `SELECT COUNT(*) FROM teacher WHERE role = 'teacher'`)
}

func scalarInt(c *gin.Context, db *sql.DB, query string, args ...any) int {
	var value int
	_ = db.QueryRowContext(c.Request.Context(), query, args...).Scan(&value)
	return value
}

func ensureSchema(db *sql.DB) {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS survey (
			id BIGINT NOT NULL AUTO_INCREMENT,
			title VARCHAR(100) NOT NULL,
			type TINYINT NOT NULL DEFAULT 0,
			scope VARCHAR(50) NOT NULL DEFAULT '全校',
			college VARCHAR(50) NULL,
			survey_group VARCHAR(50) NULL,
			start_time DATETIME NULL,
			end_time DATETIME NULL,
			status TINYINT NOT NULL DEFAULT 0,
			create_by BIGINT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_survey_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS survey_question (
			id BIGINT NOT NULL AUTO_INCREMENT,
			survey_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			question_type VARCHAR(20) NOT NULL DEFAULT 'single',
			required TINYINT(1) NOT NULL DEFAULT 1,
			sort_order INT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_survey_question_survey_id (survey_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS survey_option (
			id BIGINT NOT NULL AUTO_INCREMENT,
			question_id BIGINT NOT NULL,
			label VARCHAR(100) NOT NULL,
			score INT NOT NULL DEFAULT 0,
			sort_order INT NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			KEY idx_survey_option_question_id (question_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS survey_response (
			id BIGINT NOT NULL AUTO_INCREMENT,
			survey_id BIGINT NOT NULL,
			teacher_id BIGINT NOT NULL,
			duration_seconds INT NOT NULL DEFAULT 0,
			is_valid TINYINT(1) NOT NULL DEFAULT 1,
			submit_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_survey_response_teacher (survey_id, teacher_id),
			KEY idx_survey_response_survey_id (survey_id),
			KEY idx_survey_response_teacher_id (teacher_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS survey_answer (
			id BIGINT NOT NULL AUTO_INCREMENT,
			response_id BIGINT NOT NULL,
			survey_id BIGINT NOT NULL,
			question_id BIGINT NOT NULL,
			option_id BIGINT NULL,
			content TEXT NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_survey_answer_response_id (response_id),
			KEY idx_survey_answer_survey_id (survey_id),
			KEY idx_survey_answer_question_id (question_id),
			KEY idx_survey_answer_option_id (option_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, statement := range statements {
		_, _ = db.Exec(statement)
	}
	if !hasColumn(db, "survey", "survey_group") && !hasColumn(db, "survey", "group") {
		_, _ = db.Exec(`ALTER TABLE survey ADD COLUMN survey_group VARCHAR(50) NULL`)
	}
}

func surveyGroupColumn(c *gin.Context, db *sql.DB) string {
	if hasColumn(db, "survey", "survey_group") {
		return "survey_group"
	}
	if hasColumn(db, "survey", "group") {
		return "`group`"
	}
	_, _ = db.ExecContext(c.Request.Context(), `ALTER TABLE survey ADD COLUMN survey_group VARCHAR(50) NULL`)
	return "survey_group"
}

func hasColumn(db *sql.DB, table string, column string) bool {
	var count int
	_ = db.QueryRow(
		`SELECT COUNT(*)
		 FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table,
		column,
	).Scan(&count)
	return count > 0
}

func scopeWhere(c *gin.Context, db *sql.DB) (string, []any, error) {
	role := c.GetString("role")
	if role == "school_admin" {
		return "", nil, nil
	}
	var college string
	if err := db.QueryRowContext(c.Request.Context(), `SELECT college FROM teacher WHERE id = ?`, c.GetInt64("userID")).Scan(&college); err != nil {
		return "", nil, err
	}
	if role == "party_admin" {
		return "WHERE (s.scope = '全校' OR s.college = ? OR s.college IS NULL OR s.college = '')", []any{college}, nil
	}
	return "WHERE (s.scope = '全校' OR s.college = ? OR s.college IS NULL OR s.college = '')", []any{college}, nil
}

func defaultQuestions() []questionRequest {
	return []questionRequest{
		{Title: "近期工作压力感受", Type: "single", Required: true, Options: []optionRequest{{Label: "较轻", Score: 1}, {Label: "适中", Score: 2}, {Label: "压力较大", Score: 3}}},
		{Title: "对学院支持保障的满意度", Type: "single", Required: true, Options: []optionRequest{{Label: "满意", Score: 1}, {Label: "基本满意", Score: 2}, {Label: "不满意", Score: 3}}},
		{Title: "希望学校重点改进的问题", Type: "text", Required: false},
	}
}

func percent(value int, total int) int {
	if total <= 0 {
		return 0
	}
	return value * 100 / total
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

func typeText(value int) string {
	if value == 1 {
		return "年度长测"
	}
	return "常态短测"
}

func statusText(value int) string {
	switch value {
	case 0:
		return "未发布"
	case 1:
		return "进行中"
	case 2:
		return "已结束"
	default:
		return "未知"
	}
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func isAdmin(c *gin.Context) bool {
	role := c.GetString("role")
	return role == "party_admin" || role == "school_admin"
}

func appendWhere(where string, clause string) string {
	if where == "" {
		return "WHERE " + clause
	}
	return where + " AND " + clause
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

type errText string

func (e errText) Error() string {
	return string(e)
}

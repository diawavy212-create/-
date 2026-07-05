package profile

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"teacher-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/teachers", listTeachers(db))
	rg.POST("/teachers", createTeacher(db))
	rg.PUT("/teachers/:id", updateTeacher(db))
	rg.DELETE("/teachers/:id", removeTeacher(db))

	rg.GET("/me", func(c *gin.Context) {
		userID := c.GetInt64("userID")
		detail, ok := queryProfile(c, db, userID)
		if !ok {
			return
		}

		response.OK(c, detail)
	})

	rg.PUT("/me", func(c *gin.Context) {
		userID := c.GetInt64("userID")
		var req struct {
			Name       string `json:"name"`
			Department string `json:"department"`
			Phone      string `json:"phone"`
			Email      string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid profile payload")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Department = strings.TrimSpace(req.Department)
		req.Phone = strings.TrimSpace(req.Phone)
		req.Email = strings.TrimSpace(req.Email)
		if req.Name == "" {
			response.Fail(c, http.StatusBadRequest, "name is required")
			return
		}

		_, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE teacher SET name = ?, department = ?, phone = ?, email = ? WHERE id = ?`,
			req.Name,
			req.Department,
			req.Phone,
			req.Email,
			userID,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "profile update failed")
			return
		}

		detail, ok := queryProfile(c, db, userID)
		if !ok {
			return
		}
		response.OK(c, detail)
	})
}

func createTeacher(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		var req struct {
			Name       string `json:"name"`
			EmployeeNo string `json:"employeeNo"`
			College    string `json:"college"`
			Department string `json:"department"`
			Phone      string `json:"phone"`
			Email      string `json:"email"`
			Role       string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid teacher payload")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.EmployeeNo = strings.TrimSpace(req.EmployeeNo)
		req.College = strings.TrimSpace(req.College)
		req.Department = strings.TrimSpace(req.Department)
		req.Phone = strings.TrimSpace(req.Phone)
		req.Email = strings.TrimSpace(req.Email)
		req.Role = strings.TrimSpace(req.Role)
		if req.Role == "" {
			req.Role = "teacher"
		}
		if req.Name == "" || req.EmployeeNo == "" || req.College == "" {
			response.Fail(c, http.StatusBadRequest, "name, employeeNo and college are required")
			return
		}

		result, err := db.ExecContext(
			c.Request.Context(),
			`INSERT INTO teacher (name, user_id, college, department, phone, email, role)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			req.Name,
			req.EmployeeNo,
			req.College,
			req.Department,
			req.Phone,
			req.Email,
			req.Role,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher create failed")
			return
		}
		id, _ := result.LastInsertId()
		detail, ok := queryProfile(c, db, id)
		if !ok {
			return
		}
		response.OK(c, detail.toJSON())
	}
}

func listTeachers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		page := queryInt(c, "page", 1)
		size := queryInt(c, "size", 20)
		offset := (page - 1) * size

		var total int
		if err := db.QueryRowContext(c.Request.Context(), `SELECT COUNT(*) FROM teacher`).Scan(&total); err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher count failed")
			return
		}

		rows, err := db.QueryContext(
			c.Request.Context(),
			`SELECT id, name, user_id, COALESCE(department, ''), college, COALESCE(phone, ''), COALESCE(email, ''), role
			 FROM teacher
			 ORDER BY id DESC
			 LIMIT ? OFFSET ?`,
			size,
			offset,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher list failed")
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0)
		for rows.Next() {
			var item profileDetail
			if err := rows.Scan(&item.ID, &item.Name, &item.EmployeeNo, &item.Department, &item.College, &item.Phone, &item.Email, &item.Role); err != nil {
				response.Fail(c, http.StatusInternalServerError, "teacher scan failed")
				return
			}
			items = append(items, item.toJSON())
		}
		if err := rows.Err(); err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher rows failed")
			return
		}

		response.OK(c, pageData(page, size, total, items))
	}
}

func updateTeacher(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		teacherID := pathID(c)
		var req struct {
			Name       string `json:"name"`
			EmployeeNo string `json:"employeeNo"`
			College    string `json:"college"`
			Department string `json:"department"`
			Phone      string `json:"phone"`
			Email      string `json:"email"`
			Role       string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid teacher payload")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.EmployeeNo = strings.TrimSpace(req.EmployeeNo)
		req.College = strings.TrimSpace(req.College)
		req.Department = strings.TrimSpace(req.Department)
		req.Phone = strings.TrimSpace(req.Phone)
		req.Email = strings.TrimSpace(req.Email)
		req.Role = strings.TrimSpace(req.Role)
		if req.Name == "" || req.EmployeeNo == "" || req.College == "" || req.Role == "" {
			response.Fail(c, http.StatusBadRequest, "name, employeeNo, college and role are required")
			return
		}

		_, err := db.ExecContext(
			c.Request.Context(),
			`UPDATE teacher SET name = ?, user_id = ?, college = ?, department = ?, phone = ?, email = ?, role = ? WHERE id = ?`,
			req.Name,
			req.EmployeeNo,
			req.College,
			req.Department,
			req.Phone,
			req.Email,
			req.Role,
			teacherID,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher update failed")
			return
		}

		detail, ok := queryProfile(c, db, teacherID)
		if !ok {
			return
		}
		response.OK(c, detail.toJSON())
	}
}

func removeTeacher(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}
		teacherID := pathID(c)
		if teacherID == c.GetInt64("userID") {
			response.Fail(c, http.StatusBadRequest, "不能删除当前登录账号")
			return
		}

		result, err := db.ExecContext(c.Request.Context(), `DELETE FROM teacher WHERE id = ?`, teacherID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "teacher delete failed")
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			response.Fail(c, http.StatusNotFound, "teacher not found")
			return
		}
		response.OK(c, gin.H{"id": teacherID, "deleted": true})
	}
}

type profileDetail struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	EmployeeNo string `json:"employeeNo"`
	Department string `json:"department"`
	College    string `json:"college"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

func (d profileDetail) toJSON() gin.H {
	return gin.H{
		"id":         d.ID,
		"name":       d.Name,
		"employeeNo": d.EmployeeNo,
		"department": d.Department,
		"college":    d.College,
		"phone":      d.Phone,
		"email":      d.Email,
		"role":       d.Role,
	}
}

func queryProfile(c *gin.Context, db *sql.DB, userID int64) (profileDetail, bool) {
	var detail profileDetail
	err := db.QueryRowContext(
		c.Request.Context(),
		`SELECT id, name, user_id, COALESCE(department, ''), college, COALESCE(phone, ''), COALESCE(email, ''), role
		 FROM teacher WHERE id = ?`,
		userID,
	).Scan(&detail.ID, &detail.Name, &detail.EmployeeNo, &detail.Department, &detail.College, &detail.Phone, &detail.Email, &detail.Role)
	if err == sql.ErrNoRows {
		response.Fail(c, http.StatusNotFound, "profile not found")
		return detail, false
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "profile query failed")
		return detail, false
	}

	return detail, true
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

func isAdmin(c *gin.Context) bool {
	role := c.GetString("role")
	return role == "party_admin" || role == "school_admin"
}

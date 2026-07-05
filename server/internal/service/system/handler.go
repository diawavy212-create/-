package system

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"teacher-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

type summaryData struct {
	Teachers           int    `json:"teachers"`
	Appeals            int    `json:"appeals"`
	Trainings          int    `json:"trainings"`
	TrainingRecords    int    `json:"trainingRecords"`
	AttachmentCount    int    `json:"attachmentCount"`
	AttachmentBytes    int64  `json:"attachmentBytes"`
	AttachmentSizeText string `json:"attachmentSizeText"`
	AttachmentPath     string `json:"attachmentPath"`
	DatabaseBytes      int64  `json:"databaseBytes"`
	DatabaseSizeText   string `json:"databaseSizeText"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/summary", summary(db))
}

func summary(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdmin(c) {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}

		attachmentPath := filepath.Join("uploads", "treeholes")
		attachmentCount, attachmentBytes := attachmentUsage(attachmentPath)
		databaseBytes := databaseUsage(c, db)

		response.OK(c, summaryData{
			Teachers:           countRows(c, db, "teacher"),
			Appeals:            countRows(c, db, "appeal"),
			Trainings:          countRows(c, db, "training"),
			TrainingRecords:    countRows(c, db, "training_record"),
			AttachmentCount:    attachmentCount,
			AttachmentBytes:    attachmentBytes,
			AttachmentSizeText: formatBytes(attachmentBytes),
			AttachmentPath:     filepath.ToSlash(attachmentPath),
			DatabaseBytes:      databaseBytes,
			DatabaseSizeText:   formatBytes(databaseBytes),
		})
	}
}

func countRows(c *gin.Context, db *sql.DB, table string) int {
	var total int
	query := map[string]string{
		"teacher":         "SELECT COUNT(*) FROM teacher",
		"appeal":          "SELECT COUNT(*) FROM appeal",
		"training":        "SELECT COUNT(*) FROM training",
		"training_record": "SELECT COUNT(*) FROM training_record",
	}[table]
	if query == "" {
		return 0
	}
	if err := db.QueryRowContext(c.Request.Context(), query).Scan(&total); err != nil {
		return 0
	}
	return total
}

func databaseUsage(c *gin.Context, db *sql.DB) int64 {
	var bytes int64
	err := db.QueryRowContext(c.Request.Context(), `
		SELECT COALESCE(SUM(data_length + index_length), 0)
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`).Scan(&bytes)
	if err != nil {
		return 0
	}
	return bytes
}

func attachmentUsage(root string) (int, int64) {
	var count int
	var bytes int64
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, statErr := entry.Info()
		if statErr != nil {
			return nil
		}
		count++
		bytes += info.Size()
		return nil
	})
	return count, bytes
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func isAdmin(c *gin.Context) bool {
	role := c.GetString("role")
	return role == "party_admin" || role == "school_admin"
}

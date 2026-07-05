package auth

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"teacher-platform/server/internal/config"
	"teacher-platform/server/internal/response"
	"teacher-platform/server/internal/security"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Code    string `json:"code"`
	Ticket  string `json:"ticket"`
	Role    string `json:"role"`
	Service string `json:"service"`
}

type loginUser struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	EmployeeNo string `json:"employeeNo"`
	College    string `json:"college"`
}

func RegisterRoutes(rg *gin.RouterGroup, cfg config.Config, db *sql.DB) {
	rg.POST("/wechat-login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid login payload")
			return
		}
		if req.Code == "" {
			response.Fail(c, http.StatusBadRequest, "wechat code is required")
			return
		}

		openID, err := exchangeWechatOpenID(c, cfg, req.Code)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, err.Error())
			return
		}

		user, err := userByWechatOpenID(c, db, openID)
		if err == sql.ErrNoRows && cfg.DevAuthEnabled {
			user, err = firstUser(c, db, "teacher")
		}
		if err == sql.ErrNoRows {
			response.Fail(c, http.StatusForbidden, "wechat account is not linked")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "user query failed")
			return
		}

		writeLogin(c, cfg, user, gin.H{"openidLinked": true})
	})

	rg.POST("/cas-login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid login payload")
			return
		}
		if req.Ticket == "" {
			response.Fail(c, http.StatusBadRequest, "cas ticket is required")
			return
		}

		account, err := validateCASTicket(c, cfg, req.Ticket, req.Service)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, err.Error())
			return
		}

		user, err := userByCASAccount(c, db, account)
		if err == sql.ErrNoRows && cfg.DevAuthEnabled {
			role := req.Role
			if role == "" {
				role = "party_admin"
			}
			user, err = firstUser(c, db, role)
		}
		if err == sql.ErrNoRows {
			response.Fail(c, http.StatusForbidden, "cas account is not linked")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "admin user query failed")
			return
		}
		if user.Role != "party_admin" && user.Role != "school_admin" {
			response.Fail(c, http.StatusForbidden, "admin role is required")
			return
		}

		writeLogin(c, cfg, user, gin.H{"audience": cfg.AdminTokenAudience})
	})
}

func writeLogin(c *gin.Context, cfg config.Config, user loginUser, extra gin.H) {
	userID, _ := strconv.ParseInt(user.ID, 10, 64)
	token, err := security.SignToken(cfg.AuthTokenSecret, userID, user.Role, 8*time.Hour)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "token sign failed")
		return
	}

	body := gin.H{"token": token, "user": user, "expiresIn": 28800}
	for key, value := range extra {
		body[key] = value
	}
	response.OK(c, body)
}

func exchangeWechatOpenID(c *gin.Context, cfg config.Config, code string) (string, error) {
	if cfg.DevAuthEnabled && (cfg.WeChatAppID == "" || cfg.WeChatAppSecret == "") {
		if code == "dev" {
			return "dev-wechat-openid", nil
		}
		return code, nil
	}
	if cfg.WeChatAppID == "" || cfg.WeChatAppSecret == "" {
		return "", errors.New("wechat app is not configured")
	}

	endpoint := "https://api.weixin.qq.com/sns/jscode2session"
	values := url.Values{}
	values.Set("appid", cfg.WeChatAppID)
	values.Set("secret", cfg.WeChatAppSecret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.New("wechat login request failed")
	}
	defer resp.Body.Close()

	var body struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", errors.New("wechat login response invalid")
	}
	if body.ErrCode != 0 || body.OpenID == "" {
		if body.ErrMsg == "" {
			body.ErrMsg = "wechat login failed"
		}
		return "", errors.New(body.ErrMsg)
	}
	return body.OpenID, nil
}

func validateCASTicket(c *gin.Context, cfg config.Config, ticket string, service string) (string, error) {
	if cfg.DevAuthEnabled && cfg.CASEndpoint == "https://cas.example.edu" {
		return ticket, nil
	}
	if service == "" {
		service = cfg.CASServiceURL
	}
	if cfg.CASEndpoint == "" || service == "" {
		return "", errors.New("cas is not configured")
	}

	endpoint := stringsTrimRight(cfg.CASEndpoint, "/") + "/serviceValidate"
	values := url.Values{}
	values.Set("ticket", ticket)
	values.Set("service", service)
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.New("cas validation request failed")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("cas validation response invalid")
	}
	account, err := parseCASUser(data)
	if err != nil {
		return "", err
	}
	return account, nil
}

type casServiceResponse struct {
	AuthenticationSuccess *struct {
		User string `xml:"user"`
	} `xml:"authenticationSuccess"`
	AuthenticationFailure *struct {
		Code    string `xml:"code,attr"`
		Message string `xml:",chardata"`
	} `xml:"authenticationFailure"`
}

func parseCASUser(data []byte) (string, error) {
	var body casServiceResponse
	if err := xml.Unmarshal(data, &body); err != nil {
		return "", errors.New("cas validation response invalid")
	}
	if body.AuthenticationSuccess == nil || body.AuthenticationSuccess.User == "" {
		return "", errors.New("cas ticket validation failed")
	}
	return body.AuthenticationSuccess.User, nil
}

func userByWechatOpenID(c *gin.Context, db *sql.DB, openID string) (loginUser, error) {
	return queryUser(c, db, `SELECT id, name, role, user_id, college FROM teacher WHERE wechat_openid = ?`, openID)
}

func userByCASAccount(c *gin.Context, db *sql.DB, account string) (loginUser, error) {
	return queryUser(c, db, `SELECT id, name, role, user_id, college FROM teacher WHERE cas_account = ?`, account)
}

func firstUser(c *gin.Context, db *sql.DB, role string) (loginUser, error) {
	return queryUser(c, db, `SELECT id, name, role, user_id, college FROM teacher WHERE role = ? ORDER BY id LIMIT 1`, role)
}

func queryUser(c *gin.Context, db *sql.DB, query string, args ...any) (loginUser, error) {
	var user loginUser
	var id int64
	err := db.QueryRowContext(c.Request.Context(), query, args...).Scan(&id, &user.Name, &user.Role, &user.EmployeeNo, &user.College)
	user.ID = strconv.FormatInt(id, 10)
	return user, err
}

func stringsTrimRight(value string, cutset string) string {
	for len(value) > 0 && containsRune(cutset, rune(value[len(value)-1])) {
		value = value[:len(value)-1]
	}
	return value
}

func containsRune(value string, target rune) bool {
	for _, current := range value {
		if current == target {
			return true
		}
	}
	return false
}

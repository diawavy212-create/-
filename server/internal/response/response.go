package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: http.StatusOK, Msg: "success", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: http.StatusOK, Msg: "success", Data: data})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, Body{Code: status, Msg: message})
}

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	httpCode := http.StatusOK
	switch {
	case code >= 40000 && code < 40100:
		httpCode = http.StatusBadRequest
	case code >= 40100 && code < 40200:
		httpCode = http.StatusUnauthorized
	case code >= 40400 && code < 40500:
		httpCode = http.StatusNotFound
	case code >= 40900 && code < 41000:
		httpCode = http.StatusConflict
	default:
		httpCode = http.StatusInternalServerError
	}

	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}

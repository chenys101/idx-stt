package app

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, code int, data interface{}) {
	format, _ := c.Get("output_format")
	resp := Response{
		Code:    code,
		Message: "success",
		Data:    data,
	}

	switch format {
	case "text":
		c.String(code, "%+v", resp.Data)
	default:
		c.JSON(code, resp)
	}
}

func AbortWithError(c *gin.Context, code int, msg string) {
	resp := Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	}

	format, _ := c.Get("output_format")
	switch format {
	case "text":
		c.String(code, "Error: %s", msg)
	default:
		c.AbortWithStatusJSON(code, resp)
	}
}

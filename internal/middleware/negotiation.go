package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func ContentNegotiation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 输入处理
		contentType := c.GetHeader("Content-Type")
		switch {
		case strings.Contains(contentType, "application/json"):
			c.Set("input_format", "json")
		case strings.Contains(contentType, "text/plain"):
			c.Set("input_format", "text")
		default:
			c.Set("input_format", "json")
		}

		// 输出处理
		accept := c.GetHeader("Accept")
		switch {
		case strings.Contains(accept, "application/json"):
			c.Set("output_format", "json")
		case strings.Contains(accept, "text/plain"):
			c.Set("output_format", "text")
		default:
			c.Set("output_format", "json")
		}
		c.Next()
	}
}

package controller

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"stt/internal/model"
	"stt/internal/pkg/app"
	"stt/internal/pkg/database"
)

type UserController struct{}

func (uc *UserController) CreateUser(c *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	format, _ := c.Get("input_format")
	if format == "text" {
		// 处理 text/plain 输入
		rawData, _ := c.GetRawData()
		parts := strings.Split(string(rawData), "\n")
		for _, part := range parts {
			if strings.HasPrefix(part, "Name:") {
				input.Name = strings.TrimSpace(strings.TrimPrefix(part, "Name:"))
			}
			if strings.HasPrefix(part, "Email:") {
				input.Email = strings.TrimSpace(strings.TrimPrefix(part, "Email:"))
			}
		}
	} else {
		if err := c.ShouldBind(&input); err != nil {
			app.AbortWithError(c, 400, err.Error())
			return
		}
	}

	user := model.User{
		Name:  input.Name,
		Email: input.Email,
	}

	if result := database.DB.Create(&user); result.Error != nil {
		app.AbortWithError(c, 500, result.Error.Error())
		return
	}

	app.Success(c, 201, user)
}
func (uc *UserController) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		app.AbortWithError(c, 400, "Invalid user ID")
		return
	}

	var user model.User
	err = database.DB.First(&user, userID).Error
	if err != nil {
		if err.Error() == "record not found" {
			app.AbortWithError(c, 404, "User not found")
		} else {
			app.AbortWithError(c, 500, err.Error())
		}
		return
	}

	app.Success(c, 200, user)
}

func (uc *UserController) ListUsers(c *gin.Context) {
	var users []model.User
	if result := database.DB.Find(&users); result.Error != nil {
		app.AbortWithError(c, 500, result.Error.Error())
		return
	}
	app.Success(c, 200, users)
}

// GetUser, UpdateUser, DeleteUser 类似实现...

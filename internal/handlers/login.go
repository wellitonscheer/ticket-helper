package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/email"
)

type Login struct{}

func NewLoginHandlers() *Login {
	return &Login{}
}

func (l *Login) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login", gin.H{})
}

func (l *Login) SendEmailVefificationCode(c *gin.Context) {
	to := c.PostForm("email")

	err := email.SendEmail(to, "verification code", "your code is 343443")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to send verification code: %v", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{})
}

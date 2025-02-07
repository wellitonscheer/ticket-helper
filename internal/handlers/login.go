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
	// email := c.PostForm("email")
	to := "wellitonscheer@gmail.com"

	err := email.SendEmail(to, []byte("hello there"))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to send verification code: %v", err.Error()))
	}
}

package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/email"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
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

	verificationCode := utils.Random6Numbers()

	err := email.SendEmail(to, "verification code", fmt.Sprintf("your code is %d", verificationCode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to send verification code: %v", err.Error())})
		return
	}

	// TODO: retorna a tela de colocar o email só mudando o botao pra desabilitado
	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{"Email": to})
}

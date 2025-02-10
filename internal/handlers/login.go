package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wellitonscheer/ticket-helper/internal/sqlite"
)

type Login struct{}

func NewLoginHandlers() *Login {
	return &Login{}
}

func (l *Login) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login", gin.H{})
}

func (l *Login) InsertAuthorizedEmails(c *gin.Context) {
	sqliteLogin, err := sqlite.NewSqliteLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed initialize login: %v", err.Error())})
		return
	}

	err = sqliteLogin.InsertAuthorizedEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to insert authorized emails: %v", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "done"})
}

func (l *Login) SendEmailVefificationCode(c *gin.Context) {
	to := c.PostForm("email")
	if len(to) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid email"})
		return
	}

	sqliteLogin, err := sqlite.NewSqliteLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed initialize login: %v", err.Error())})
		return
	}

	authorized, err := sqliteLogin.IsAuthorizedEmail()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed verify email: %v", err.Error())})
		return
	}
	if !authorized {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "email not authorized to login"})
		return
	}

	// verificationCode := utils.Random6Numbers()

	// err = email.SendEmail(to, "verification code", fmt.Sprintf("your code is %d", verificationCode))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to send verification code: %v", err.Error())})
	// 	return
	// }

	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{"Email": to})
}

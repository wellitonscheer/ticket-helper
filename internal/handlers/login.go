package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wellitonscheer/ticket-helper/internal/email"
	"github.com/wellitonscheer/ticket-helper/internal/sqlite"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
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

	authorized, err := sqliteLogin.IsAuthorizedEmail(to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed verify email: %v", err.Error())})
		return
	}
	if !authorized {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "email not authorized to login"})
		return
	}

	verificationCode := utils.Random6Numbers()

	err = sqliteLogin.InsertVerificationCode(to, verificationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to save verification code: %v", err.Error())})
		return
	}

	err = email.SendEmail(to, "verification code", fmt.Sprintf("your code is %d", verificationCode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to send verification code: %v", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{"Email": to})
}

func (l *Login) ValidateVefificationCode(c *gin.Context) {
	email := c.PostForm("email")
	code := c.PostForm("code")
	if len(email) == 0 || len(code) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid email or verification code"})
		return
	}

	sqliteLogin, err := sqlite.NewSqliteLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed initialize login: %v", err.Error())})
		return
	}

	intCode, err := strconv.Atoi(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to convert code to number: %v", err.Error())})
		return
	}

	isValid, err := sqliteLogin.IsValidVefificationCode(email, intCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to validate verification code: %v", err.Error())})
		return
	}

	if isValid {
		tokenUuid, err := uuid.NewRandom()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to generate uuid: %v", err.Error())})
			return
		}
		tokenString := tokenUuid.String()

		sqliteLogin.CreateUserSession(email, tokenString)

		c.SetCookie("session_token", tokenString, 60*60*3, "/", "caie.dev", true, true)
	}

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/")
		c.Status(http.StatusOK)
	} else {
		c.Header("Location", "/")
		c.Status(http.StatusSeeOther)
	}
	c.Abort()
}

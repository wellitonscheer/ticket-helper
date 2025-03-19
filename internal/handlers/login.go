package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/liteservi"
	"github.com/wellitonscheer/ticket-helper/internal/sqlite"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
)

type LoginHandlers struct {
	appContext context.AppContext
}

func NewLoginHandlers(appContext context.AppContext) LoginHandlers {
	return LoginHandlers{
		appContext: appContext,
	}
}

func (l LoginHandlers) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login", gin.H{})
}

func (l LoginHandlers) SendEmailVefificationCode(c *gin.Context) {
	to := c.PostForm("email")
	if len(to) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid email"})
		return
	}

	authoEmailService := liteservi.NewAuthorizedEmailsService(l.appContext)

	authorized := authoEmailService.IsAuthorizedEmail(to)
	if !authorized {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "email not authorized to login"})
		return
	}

	verificationCode := utils.Random6Numbers()

	veriCodeService := liteservi.NewVerificationCodeService(l.appContext)

	err := veriCodeService.NewVerificationCode(to, verificationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to save verification code: %v", err.Error())})
		return
	}

	err = utils.SendEmail(l.appContext.Config.Email, to, "verification code", fmt.Sprintf("your code is %d", verificationCode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to send verification code: %v", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{"Email": to})
}

func (l LoginHandlers) ValidateVefificationCode(c *gin.Context) {
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

		c.SetCookie("session_token", tokenString, 60*60*3, "/", "", true, true)
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

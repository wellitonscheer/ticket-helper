package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/liteservi"
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
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid email"})
		return
	}

	authoEmailService := liteservi.NewAuthorizedEmailsService(l.appContext)

	authorized := authoEmailService.IsAuthorizedEmail(email)
	if !authorized {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "email not authorized to login"})
		return
	}

	verificationCode := utils.Random6Numbers()

	veriCodeService := liteservi.NewVerificationCodeService(l.appContext)

	err := veriCodeService.NewVerificationCode(email, verificationCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to save verification code: %v", err.Error())})
		return
	}

	err = utils.SendEmail(l.appContext.Config.Email, email, "verification code", fmt.Sprintf("your code is %d", verificationCode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to send verification code: %v", err.Error())})
		return
	}

	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{"Email": email})
}

func (l LoginHandlers) LoginWithCode(c *gin.Context) {
	email := c.PostForm("email")
	code := c.PostForm("code")
	if email == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid email or verification code"})
		return
	}

	intCode, err := strconv.Atoi(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to convert code to number: %v", err.Error())})
		return
	}

	verificCodeServ := liteservi.NewVerificationCodeService(l.appContext)

	verificCode, err := verificCodeServ.GetByEmailCode(email, intCode)
	if err != nil || !verificCode.IsValid() {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "invalid verification code"})
		return
	}

	sessionServi := liteservi.NewSessionService(l.appContext)
	token, err := sessionServi.NewSessionByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to create session"})
		return
	}

	c.SetCookie("session_token", token, int(l.appContext.Config.Common.SessionLifetimeSec), "/", "", true, true)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/")
		c.Status(http.StatusOK)
	} else {
		c.Header("Location", "/")
		c.Status(http.StatusSeeOther)
	}
	c.Abort()
}

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
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusBadRequest,
			LogMsg:  "invalid email",
			UserMsg: "invalid email",
		})
		return
	}

	authoEmailService := liteservi.NewAuthorizedEmailsService(l.appContext)

	authorized := authoEmailService.IsAuthorizedEmail(email)
	if !authorized {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusForbidden,
			LogMsg:  fmt.Sprintf("email not authorized to login (email=%s)\n", email),
			UserMsg: "email not authorized to login",
		})
		return
	}

	verificationCode := utils.Random6Numbers()

	veriCodeService := liteservi.NewVerificationCodeService(l.appContext)

	err := veriCodeService.NewVerificationCode(email, verificationCode)
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("failed to save verification code (email=%s, verificationCode=%d): %v\n", email, verificationCode, err),
			UserMsg: "failed to send verification code",
		})
		return
	}

	err = utils.SendEmail(l.appContext.Config.Email, email, "verification code", fmt.Sprintf("your code is %d", verificationCode))
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("failed to send email with verification code (email=%s, verificationCode=%d): %v\n", email, verificationCode, err),
			UserMsg: "failed to send email with verification code",
		})
		return
	}

	c.HTML(http.StatusOK, "sent-verification-code-success", gin.H{"Email": email})
}

func (l LoginHandlers) LoginWithCode(c *gin.Context) {
	email := c.PostForm("email")
	code := c.PostForm("code")
	if email == "" || code == "" {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusBadRequest,
			LogMsg:  fmt.Sprintf("invalid email or verification code (email=%s, code=%s)\n", email, code),
			UserMsg: "invalid email or verification code",
		})
		return
	}

	intCode, err := strconv.Atoi(code)
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("failed to convert code to number (code=%s): %v\n", code, err),
			UserMsg: "failed to verify code",
		})
		return
	}

	verificCodeServ := liteservi.NewVerificationCodeService(l.appContext)

	verificCode, err := verificCodeServ.GetByEmailCode(email, intCode)
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("failed to get verification code (email=%s, code=%d): %v\n", email, intCode, err),
			UserMsg: "invalid verification code",
		})
		return
	}

	if !verificCode.IsValid() {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusUnauthorized,
			LogMsg:  fmt.Sprintf("invalid verification code (verificCode=%+v)\n", verificCode),
			UserMsg: "invalid verification code",
		})
		return
	}

	if err = verificCodeServ.DeleteById(verificCode.Id); err != nil {
		fmt.Printf("failed to delete used verification code (verification code entry=%+v): %v", verificCode, err)
	}

	sessionServi := liteservi.NewSessionService(l.appContext)
	token, err := sessionServi.NewSessionByEmail(email)
	if err != nil {
		utils.HandleError(c, utils.HandleErrorInput{
			Code:    http.StatusInternalServerError,
			LogMsg:  fmt.Sprintf("failed to create session (email=%s): %v", email, err),
			UserMsg: "failed to create session",
		})
		return
	}

	c.SetCookie("session_token", token, int(l.appContext.Config.Common.SessionLifetime), "/", "", true, true)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/")
		c.Status(http.StatusOK)
	} else {
		c.Header("Location", "/")
		c.Status(http.StatusSeeOther)
	}
	c.Abort()
}

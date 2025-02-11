package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/sqlite"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_redirect := func() {
			if c.GetHeader("HX-Request") == "true" {
				c.Header("HX-Redirect", "/login")
				c.Status(http.StatusOK)
			} else {
				c.Header("Location", "/login")
				c.Status(http.StatusSeeOther)
			}
			c.Abort()
		}

		authToken, err := c.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				_redirect()
				return
			}

			c.Status(http.StatusBadRequest)
			c.Abort()
			return
		}

		if len(authToken) == 0 {
			_redirect()
			return
		}

		sqliteLogin, err := sqlite.NewSqliteLogin()
		if err != nil {
			fmt.Printf("error to create sqlite login service: %v", err.Error())

			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		isValidSession, err := sqliteLogin.IsValidSession(authToken)
		if err != nil {
			fmt.Printf("error validate session: %v", err.Error())

			_redirect()
			return
		}

		if !isValidSession {
			_redirect()
			return
		}

		c.Next()
	}
}

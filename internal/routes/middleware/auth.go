package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/sqlite/liteservi"
)

func AuthMiddleware(appContext context.AppContext) gin.HandlerFunc {
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
			if errors.Is(err, http.ErrNoCookie) {
				_redirect()
				return
			}

			fmt.Printf("failed to get session cookie: %v", err)
			c.Status(http.StatusBadRequest)
			c.Abort()
			return
		}

		if len(authToken) == 0 {
			_redirect()
			return
		}

		sessionService := liteservi.NewSessionService(appContext)
		session, err := sessionService.GetByToken(authToken)
		if err != nil {
			_redirect()
			return
		}

		if !session.IsValid() {
			_redirect()
			return
		}

		c.Next()
	}
}

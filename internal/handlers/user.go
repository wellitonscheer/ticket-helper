package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserNew(c *gin.Context) {
	user := c.Params.ByName("name")

	c.String(http.StatusOK, user)
}

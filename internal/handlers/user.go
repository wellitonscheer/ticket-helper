package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/internal/milvus"
)

func UserNew(c *gin.Context) {
	user := c.Params.ByName("name")
	err := milvus.User.NewUser(&user)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to create user: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, user)
}

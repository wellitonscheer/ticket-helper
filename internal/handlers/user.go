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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("failed to create user: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user": user})
}

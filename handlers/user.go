package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wellitonscheer/ticket-helper/db"
)

var dbMock = make(map[string]string)

func User(c *gin.Context) {
	user := c.Params.ByName("name")
	err := db.User.Save(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create user: %s", err.Error())})
		return
	}
	value, ok := dbMock[user]
	if ok {
		c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
	} else {
		c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
	}
}

package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type HandleErrorInput struct {
	Code    int
	LogMsg  string
	UserMsg string
}

func HandleError(c *gin.Context, input HandleErrorInput) {
	fmt.Printf(input.LogMsg)
	c.String(input.Code, input.UserMsg)
}

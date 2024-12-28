package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Count struct {
	Count int
}

var count = Count{
	Count: 0,
}

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index", count)
}

func IndexCount(c *gin.Context) {
	count.Count++
	c.HTML(http.StatusOK, "count", count)
}

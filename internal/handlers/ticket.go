package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	entireTicketContent = "entire"
	singleMessages      = "message"
	suggestReply        = "reply"
)

func TicketVectorSearch(c *gin.Context) {
	searchInput := c.PostForm("search-input")
	searchType := c.PostForm("search-type")

	if searchInput == "" || searchType == "" {
		c.String(http.StatusBadRequest, "invalid search data")
		return
	}

	if searchType == entireTicketContent {

	} else if searchType == singleMessages {

	} else if searchType == suggestReply {

	} else {
		c.String(http.StatusBadRequest, "invalid search type")
		return
	}

	c.HTML(http.StatusOK, "results", gin.H{})
}

package html

import "github.com/gin-gonic/gin"

type scEventViewQuery struct {
	ListHref string
	CardHref string
	IsList   bool
	IsCard   bool
}

func buildScEventViewQuery(c *gin.Context) *scEventViewQuery {
	q := c.Request.URL.Query()

	listHref := "/sc-events"
	cardHref := "/sc-events-cards"
	if enc := q.Encode(); enc != "" {
		listHref += "?" + enc
		cardHref += "?" + enc
	}

	return &scEventViewQuery{
		ListHref: listHref,
		CardHref: cardHref,
		IsList:   c.Request.URL.Path == "/sc-events",
		IsCard:   c.Request.URL.Path == "/sc-events-cards",
	}
}

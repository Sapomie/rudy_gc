package common

import (
	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/types"
)

func ParseMovieCardRequest(c *gin.Context, base types.ListMovieFullRequest) (types.ListMovieFullRequest, error) {
	request := base
	if err := c.ShouldBindQuery(&request); err != nil {
		return types.ListMovieFullRequest{}, err
	}
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 || request.PageSize > MaxPageSize {
		request.PageSize = DefaultPageSize
	}
	request.OrderBy = NormalizeOrderBy(c.Query("od"), request.OrderBy)
	request.Order = NormalizeOrder(c.Query("order"))
	return request, nil
}

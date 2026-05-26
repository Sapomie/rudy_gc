package response

import "github.com/gin-gonic/gin"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Data  any        `json:"data"`
	Error *ErrorBody `json:"error"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{
		Data:  data,
		Error: nil,
	})
}

func Fail(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Envelope{
		Data: nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

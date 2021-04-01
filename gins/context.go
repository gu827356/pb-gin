package gins

import "github.com/gin-gonic/gin"

const (
	contextKeyRequest  = "_PB_GIN_REQUEST"
	contextKeyResponse = "_PB_GIN_RESPONSE"
	contextKeyError    = "_PB_GIN_ERROR"
)

func RequestInContext(c *gin.Context) (interface{}, bool) {
	return c.Get(contextKeyRequest)
}

func ResponseInContext(c *gin.Context) (interface{}, bool) {
	return c.Get(contextKeyResponse)
}

func ErrorInContext(c *gin.Context) (interface{}, bool) {
	return c.Get(contextKeyError)
}

package gins

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gu827356/pb-gin/pb_gen/googleapis/api/annotations"
)

func RegisterRoute(route gin.IRouter, rule *annotations.HttpRule, f func(c *gin.Context) (interface{}, error)) {
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		route.GET(pattern.Get, createHandlerFunc(f))
	case *annotations.HttpRule_Post:
		route.POST(pattern.Post, createHandlerFunc(f))
	case *annotations.HttpRule_Put:
		route.PUT(pattern.Put, createHandlerFunc(f))
	default:
		panic(fmt.Errorf("now not support this pattern: %+v", pattern))
	}

	for _, bind := range rule.AdditionalBindings {
		RegisterRoute(route, bind, f)
	}
}

func createHandlerFunc(f func(c *gin.Context) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		out, err := f(c)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		render(c, out)
	}
}

func render(c *gin.Context, resp interface{}) {
	if c.ContentType() == binding.MIMEPROTOBUF {
		c.ProtoBuf(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

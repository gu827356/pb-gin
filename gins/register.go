package gins

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
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
			switch err.(type) {
			case *RequestBindErr:
				c.String(http.StatusBadRequest, err.Error())
			default:
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			return

		}
		render(c, out)
	}
}

func render(c *gin.Context, resp interface{}) {
	if c.ContentType() == binding.MIMEPROTOBUF {
		c.ProtoBuf(http.StatusOK, resp)
		return
	}

	if pbResp, ok := resp.(proto.Message); ok {
		c.Render(http.StatusOK, jsonpbRender{
			Data:      pbResp,
			marshaler: &defaultJSONPBMarshaler,
		})
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

var defaultJSONPBMarshaler = jsonpb.Marshaler{
	OrigName:     true,
	EnumsAsInts:  true,
	EmitDefaults: true,
}

var jsonContentType = []string{"application/json; charset=utf-8"}

type jsonpbRender struct {
	Data      proto.Message
	marshaler *jsonpb.Marshaler
}

func (j jsonpbRender) Render(writer http.ResponseWriter) error {
	writeContentType(writer, jsonContentType)
	d, err := j.marshaler.MarshalToString(j.Data)
	if err != nil {
		panic(err)
	}
	_, err = writer.Write([]byte(d))
	return err
}

func (j jsonpbRender) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

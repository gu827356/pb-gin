package gins

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gu827356/pb-gin/pb_gen/googleapis/api/annotations"
)

var (
	protoForm      = protoFormBinding{}
	protoMultipart = protoMultipartBinding{}
	protoJSON      = protoJSONBinding{
		Unmarshaler: jsonpb.Unmarshaler{AllowUnknownFields: true},
	}
	protoFormPost = protoFormPostBinding{}
)

func ParseRequest(c *gin.Context, in interface{}, _ *annotations.HttpRule) error {
	err := defaultBinding(c.Request.Method, c.ContentType()).Bind(c.Request, in)
	if err == nil {
		c.Set(contextKeyRequest, in)
	} else {
		c.Set(contextKeyError, err)
	}
	return err
}

// 由于使用 proto 生成的 struct 没有 form tag，因此在这里 hack 下，使用 json tag 进行解析
func defaultBinding(method, contentType string) binding.Binding {
	if method == http.MethodGet {
		return protoForm
	}
	switch contentType {
	case binding.MIMEJSON:
		return &protoJSON
	case binding.MIMEPOSTForm:
		return protoFormPost
	case binding.MIMEMultipartPOSTForm:
		return protoMultipart
	}
	return binding.Default(method, contentType)
}

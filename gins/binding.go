package gins

import (
	"errors"
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

const defaultMemory = 32 << 20

type protoFormBinding struct{}

func (protoFormBinding) Name() string {
	return "proto-form"
}

func (protoFormBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}
	if err := mapFormByTag(obj, req.Form, "json"); err != nil {
		return err
	}
	return validate(obj)
}

type protoJSONBinding struct {
	jsonpb.Unmarshaler
}

func (b *protoJSONBinding) Name() string {
	return "proto-json"
}

func (b *protoJSONBinding) Bind(req *http.Request, obj interface{}) error {
	msg, ok := obj.(proto.Message)
	if ok {
		return b.Unmarshal(req.Body, msg)
	} else {
		return binding.JSON.Bind(req, obj)
	}
}

type multipartRequest http.Request

// TrySet tries to set a value by the multipart request with the binding a form file
func (r *multipartRequest) TrySet(value reflect.Value, field reflect.StructField, key string, opt setOptions) (isSetted bool, err error) {
	if files := r.MultipartForm.File[key]; len(files) != 0 {
		return setByMultipartFormFile(value, field, files)
	}

	return setByForm(value, field, r.MultipartForm.Value, key, opt)
}
func setByMultipartFormFile(value reflect.Value, field reflect.StructField, files []*multipart.FileHeader) (isSetted bool, err error) {
	switch value.Kind() {
	case reflect.Ptr:
		switch value.Interface().(type) {
		case *multipart.FileHeader:
			value.Set(reflect.ValueOf(files[0]))
			return true, nil
		}
	case reflect.Struct:
		switch value.Interface().(type) {
		case multipart.FileHeader:
			value.Set(reflect.ValueOf(*files[0]))
			return true, nil
		}
	case reflect.Slice:
		slice := reflect.MakeSlice(value.Type(), len(files), len(files))
		isSetted, err = setArrayOfMultipartFormFiles(slice, field, files)
		if err != nil || !isSetted {
			return isSetted, err
		}
		value.Set(slice)
		return true, nil
	case reflect.Array:
		return setArrayOfMultipartFormFiles(value, field, files)
	}
	return false, errors.New("unsupported field type for multipart.FileHeader")
}

func setArrayOfMultipartFormFiles(value reflect.Value, field reflect.StructField, files []*multipart.FileHeader) (isSetted bool, err error) {
	if value.Len() != len(files) {
		return false, errors.New("unsupported len of array for []*multipart.FileHeader")
	}
	for i := range files {
		setted, err := setByMultipartFormFile(value.Index(i), field, files[i:i+1])
		if err != nil || !setted {
			return setted, err
		}
	}
	return true, nil
}

type protoMultipartBinding struct{}

func (protoMultipartBinding) Name() string {
	return "proto-multipart/form-data"
}

func (protoMultipartBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	if err := mappingByPtr(obj, (*multipartRequest)(req), "json"); err != nil {
		return err
	}

	return validate(obj)
}

type protoFormPostBinding struct{}

func (protoFormPostBinding) Name() string {
	return "proto-form-urlencoded"
}

func (protoFormPostBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := mapFormByTag(obj, req.PostForm, "json"); err != nil {
		return err
	}
	return validate(obj)
}

func validate(obj interface{}) error {
	if binding.Validator == nil {
		return nil
	}
	return binding.Validator.ValidateStruct(obj)
}

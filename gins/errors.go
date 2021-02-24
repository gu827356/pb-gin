package gins

import "fmt"

type RequestBindErr struct {
	source error
}

func NewRequestBindErr(source error) *RequestBindErr {
	return &RequestBindErr{source: source}
}

func (e *RequestBindErr) Error() string {
	return fmt.Sprintf("fail to bind request: %v", e.source)
}

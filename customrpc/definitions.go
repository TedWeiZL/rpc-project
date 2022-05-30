package customrpc

import (
	"context"
	"reflect"
)

var (
	Method2Handler map[string]func(context.Context, interface{}) (interface{}, error)
	Method2ReqType map[string]reflect.Type
)

// error
type WrappedErr struct {
	Msg string `json:"msg"`
}

func (we *WrappedErr) Error() string {
	return we.Msg
}

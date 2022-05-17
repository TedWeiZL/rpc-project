package customrpc

import (
	"context"
	"reflect"
)

var method2Handler = make(map[string]func(context.Context, interface{}) (interface{}, error), 1)

func registerhandler(handler ServerToImplement) {
	// technique learnt from grpc official package: how to register a handler function
	method2Handler["TestConn"] = func(ctx context.Context, req interface{}) (interface{}, error) {
		return handler.TestConn(ctx, req.(*TestConnReq))
	}
}

var (
	method2ReqType = map[string]reflect.Type{"TestConn": reflect.TypeOf(TestConnReq{})}
)

// server methods
type ServerToImplement interface {
	TestConn(ctx context.Context, req *TestConnReq) (*TestConnRsp, error)
}

// client methods
type ClientMethods interface {
	TestConn(ctx context.Context, req *TestConnReq) (*TestConnRsp, error)
}

type TestConnRsp struct {
	ClientIP  string `json:"client_ip"`
	ClientMsg string `json:"client_message"`
}

type TestConnReq struct {
	Msg string `json:"message"`
}

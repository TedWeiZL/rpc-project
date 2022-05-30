package test_service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	rpc "rpc_project/customrpc"
	"strconv"
)

func init() {
}

func registerhandler(handler ServerToImplement) {
	// technique learnt from grpc official package: how to register a handler function
	rpc.Method2Handler = make(map[string]func(context.Context, interface{}) (interface{}, error), 1)
	rpc.Method2Handler["TestConn"] = func(ctx context.Context, req interface{}) (interface{}, error) {
		return handler.TestConn(ctx, req.(*TestConnReq))
	}
}

func registerReqTypes() {
	rpc.Method2ReqType = map[string]reflect.Type{"TestConn": reflect.TypeOf(TestConnReq{})}
}

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

type Client struct {
	cc *rpc.Client
}

func (c *Client) TestConn(ctx context.Context, req *TestConnReq) (*TestConnRsp, error) {
	rsp := TestConnRsp{}
	err := c.cc.Invoke(ctx, "TestConn", &req, &rsp)
	if err != nil {
		return nil, err
	}

	return &rsp, nil
}

func InitClient(ip string, port int, connNum int) (ClientMethods, error) {
	if port < 0 || connNum <= 0 {
		return nil, errors.New("invalid port number or producerCnt")
	}
	addr := ip + ":" + strconv.Itoa(port)
	c := &rpc.Client{Addr: addr,
		Conns: make(chan net.Conn, connNum),
		Codec: &rpc.JSONCodec{},
	}

	for i := 1; i <= connNum; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Printf("getClient Dial failed, err = %v", err)
			return nil, err
		}
		c.Conns <- conn
	}

	return &Client{cc: c}, nil
}

func InitServer(handler ServerToImplement, port, workerCnt int) {
	// register handlers
	registerhandler(handler)

	// init worker pool
	if workerCnt <= 0 || workerCnt > 40 {
		workerCnt = 5
	}
	rpc.WPool = rpc.InitWorkerPool(workerCnt)

	// init carrier pool
	rpc.CPool = &rpc.CarrierPool{}
}

package main

import (
	"context"
	"rpc_project/test_service"
)

type server struct{}

func (s *server) TestConn(ctx context.Context, req *test_service.TestConnReq) (*test_service.TestConnRsp, error) {
	return &test_service.TestConnRsp{
		ClientIP:  "127.0.0.1",
		ClientMsg: req.Msg,
	}, nil
}

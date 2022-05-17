package main

import (
	"context"
	rpc "rpc_project/customrpc"
)

type server struct{}

func (d *server) TestConn(ctx context.Context, req *rpc.TestConnReq) (*rpc.TestConnRsp, error) {
	return &rpc.TestConnRsp{
		ClientIP:  "127.0.0.1",
		ClientMsg: req.Msg,
	}, nil
}

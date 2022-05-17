package customrpc

import "context"

type dummyHandler struct{}

func (d *dummyHandler) TestConn(ctx context.Context, req *TestConnReq) (*TestConnRsp, error) {
	return &TestConnRsp{
		ClientIP:  "127.0.0.1",
		ClientMsg: req.Msg,
	}, nil
}

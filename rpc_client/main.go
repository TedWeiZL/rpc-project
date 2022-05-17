package main

import (
	"context"
	"rpc_project/clog"
	rpc "rpc_project/customrpc"
	"strconv"
	"time"
)

func main() {
	clog.InitLog("./logs.txt")

	clt, err := rpc.InitClient("127.0.0.1", 2345, 4)
	if err != nil {
		clog.Logger.Printf("rpc.InitClient failed, err=%v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		testConnRsp, err := clt.TestConn(ctx, &rpc.TestConnReq{Msg: "Hello, server. Message " + strconv.Itoa(i)})
		if err != nil {
			clog.Logger.Printf("clt.TestConn failed, err=%v\n", err)
			return
		}
		clog.Logger.Printf("testConnRsp.Msg=%s, testConnRsp.ClientMsg=%s", testConnRsp.ClientMsg, testConnRsp.ClientIP)
		time.Sleep(time.Second)
	}

}

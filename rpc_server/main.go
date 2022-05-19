package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"rpc_project/clog"
	rpc "rpc_project/customrpc"
	"strconv"
)

func main() {
	clog.InitLog("./server_log.txt")

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	// init server
	s := &server{}
	rpc.InitServer(s, 2345, 5)

	// get listener
	lis, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(2345))
	if err != nil {
		clog.Logger.Printf("net.Listen failed, err=%v", err)
		return
	}
	defer lis.Close()

	// serve listener
	rpc.Serve(lis)

}

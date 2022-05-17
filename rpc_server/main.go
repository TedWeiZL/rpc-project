package main

import (
	rpc "rpc_project/customrpc"
)

func main() {
	s := &server{}
	rpc.StartServer(s, 2345)
}

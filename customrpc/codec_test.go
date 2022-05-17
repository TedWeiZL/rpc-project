package customrpc

import (
	"testing"
)

func TestJsonReqEncodeDecode(t *testing.T) {
	methodName := "TestConn"
	msg := "Hello"
	req := TestConnReq{Msg: msg}
	j := &jsonCodec{}

	reqByte, err := j.EncodeReq(methodName, &req)
	if err != nil {
		t.Fatalf("encoder failed, err=%v", err)
	}

	mName, reqOutInterface, err := j.DecodeReq(reqByte)
	if err != nil {
		t.Fatalf("decoder failed, err=%v", err)
	}

	reqOut, ok := reqOutInterface.(*TestConnReq)
	if !ok {
		t.Fatalf("TestConnReq type assertion failed")
	}

	if reqOut.Msg != msg {
		t.Fatalf("Msg is wrong")
	}

	if mName != methodName {
		t.Fatalf("Method Name is wrong, expecting: %s, got: %s", methodName, mName)
	}
}

func TestJsonRspEncodeDecode(t *testing.T) {

	t.Run("TestJsonEncodeDecodeNilErr", func(tt *testing.T) {
		rsp := TestConnRsp{
			ClientIP:  "127.0.0.1",
			ClientMsg: "Hello",
		}

		j := &jsonCodec{}

		rspByte, err := j.EncodeRsp(rsp, nil)
		if err != nil {
			t.Fatalf("j.EncodeRsp failed, err=%v", err)
		}

		var rspOut TestConnRsp

		errOut, err := j.DecodeRsp(rspByte, &rspOut)
		if err != nil {
			t.Fatalf("j.DecodeRsp(rspByte, &rspOut) failed, err=%v", err)
		}

		if rspOut.ClientIP != "127.0.0.1" || rspOut.ClientMsg != "Hello" {
			t.Fatalf("rspOut got wrong value")
		}

		if errOut != nil {
			t.Fatalf("errOut is wrong, expecting nil, got non-nil")
		}
	})

	t.Run("TestJsonEncodeDecodeNonNilErr", func(tt *testing.T) {
		rsp := TestConnRsp{
			ClientIP:  "127.0.0.1",
			ClientMsg: "Hello",
		}
		rappedError := WrappedErr{Msg: "No error"}

		j := &jsonCodec{}

		rspByte, err := j.EncodeRsp(rsp, &rappedError)
		if err != nil {
			t.Fatalf("j.EncodeRsp failed, err=%v", err)
		}

		var rspOut TestConnRsp

		errOut, err := j.DecodeRsp(rspByte, &rspOut)
		if err != nil {
			t.Fatalf("j.DecodeRsp(rspByte, &rspOut) failed, err=%v", err)
		}

		if rspOut.ClientIP != "127.0.0.1" || rspOut.ClientMsg != "Hello" {
			t.Fatalf("rspOut got wrong value")
		}

		if errOut.Error() != "No error" {
			t.Fatalf("errOut got wrong value")
		}
	})

}

// Benchmark Tests
func BenchmarkJsonReqCodec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		methodName := "TestConn"
		msg := "Hello"
		req := TestConnReq{Msg: msg}
		j := &jsonCodec{}

		reqByte, _ := j.EncodeReq(methodName, &req)

		j.DecodeReq(reqByte)
	}
}

func BenchmarkJsonRspCodec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rsp := TestConnRsp{
			ClientIP:  "127.0.0.1",
			ClientMsg: "Hello",
		}
		rappedError := WrappedErr{Msg: "No error"}

		j := &jsonCodec{}

		j.EncodeRsp(rsp, &rappedError)
		// var rspOut TestConnRsp

		// j.DecodeRsp(rspByte, &rspOut)
	}
}

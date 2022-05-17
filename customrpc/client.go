package customrpc

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type client struct {
	addr string
	codec
	conns chan net.Conn
}

func (c *client) TestConn(ctx context.Context, req *TestConnReq) (*TestConnRsp, error) {
	rsp := TestConnRsp{}
	err := c.invoke(ctx, "TestConn", &req, &rsp)
	if err != nil {
		return nil, err
	}

	return &rsp, nil
}

func (c *client) invoke(ctx context.Context, method string, in interface{}, out interface{}) error {
	conn, err := c.getConn(ctx)
	if err != nil {
		fmt.Printf("c.getConn(ctx) failed, err=%v \n", err)
		return err
	}

	payload, err := c.codec.EncodeReq(method, in)
	if err != nil {
		fmt.Printf("codec encode failed err=%v \n", err)
		return err
	}

	err = sendReq(ctx, conn, payload)
	if err != nil {
		fmt.Printf("invoke sendMsg failed, err=%v \n", err)
		return err
	}

	outBytes, err := receiveRsp(ctx, conn)
	if err != nil {
		fmt.Printf("invoke receiveMsg failed, err=%v \n", err)
		return err
	}

	c.returnConn(ctx, conn)

	wrappedErr, _ := c.DecodeRsp(outBytes, out)

	// nil is not nil!
	// https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/
	if wrappedErr == nil {
		return nil
	}
	return wrappedErr
}

func (c *client) getConn(ctx context.Context) (net.Conn, error) {
	for {
		select {
		case con := <-c.conns:
			return con, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
}

func (c *client) returnConn(ctx context.Context, conn net.Conn) {
	c.conns <- conn
}

func InitClient(ip string, port int, connNum int) (ClientMethods, error) {
	if port < 0 || connNum <= 0 {
		return nil, errors.New("invalid port number or producerCnt")
	}
	addr := ip + ":" + strconv.Itoa(port)
	c := &client{addr: addr,
		conns: make(chan net.Conn, connNum),
		codec: &jsonCodec{},
	}

	for i := 1; i <= connNum; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Printf("getClient Dial failed, err = %v", err)
			return nil, err
		}
		c.conns <- conn
	}

	return c, nil
}

func sendReq(ctx context.Context, conn net.Conn, payload []byte) error {
	done := make(chan struct{})
	errs := make(chan error)

	go func() {
		deadline, ok := ctx.Deadline()
		if ok {
			conn.SetDeadline(deadline)
		} else {
			conn.SetDeadline(time.Now().Add(10 * time.Second))
		}
		_, err := conn.Write(payload)
		if err != nil {
			errs <- err
			fmt.Printf("conn.Write(payload) failed, err=%v \n", err)
			return
		}
		close(done)
	}()

	for {
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			return err
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func receiveRsp(ctx context.Context, conn net.Conn) ([]byte, error) {
	done := make(chan struct{})
	errs := make(chan error)
	var res []byte

	go func() {
		deadline, ok := ctx.Deadline()
		if ok {
			conn.SetDeadline(deadline)
		} else {
			conn.SetDeadline(time.Now().Add(10 * time.Second))
		}

		rspLenByte := make([]byte, 4, 4)
		_, err := conn.Read(rspLenByte)
		if err != nil {
			errs <- err
			fmt.Printf("conn.Read(rspLenByte) failed, err=%v", err)
			return
		}
		rspLen := binary.LittleEndian.Uint32(rspLenByte)

		errLenByte := make([]byte, 4, 4)
		_, err = conn.Read(errLenByte)
		if err != nil {
			errs <- err
			fmt.Printf("conn.Read(errLenByte) failed, err=%v", err)
			return
		}
		errLen := binary.LittleEndian.Uint32(errLenByte)

		buf := make([]byte, 8+rspLen+errLen, 8+rspLen+errLen)
		copy(buf[0:4], rspLenByte)
		copy(buf[4:8], errLenByte)

		_, err = conn.Read(buf[8:])
		if err != nil {
			errs <- err
			fmt.Printf("conn.Read(buf[8:]) failed, err=%v", err)
			return
		}

		res = buf

		close(done)
	}()

	for {
		select {
		case <-done:
			return res, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errs:
			return nil, err
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
}

package minirpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Addr string
	Codec
	Conns chan net.Conn
}

func (c *Client) Invoke(ctx context.Context, method string, in interface{}, out interface{}) error {
	conn, err := c.getConn(ctx)
	if err != nil {
		fmt.Printf("c.getConn(ctx) failed, err=%v \n", err)
		return err
	}

	payload, err := c.Codec.EncodeReq(method, in)
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

func (c *Client) getConn(ctx context.Context) (net.Conn, error) {
	for {
		select {
		case con := <-c.Conns:
			return con, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
}

func (c *Client) returnConn(ctx context.Context, conn net.Conn) {
	c.Conns <- conn
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

package customrpc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type codec interface {
	EncodeReq(method string, req interface{}) ([]byte, error)
	DecodeReq(in []byte) (string, interface{}, error)

	EncodeRsp(rsp interface{}, err *WrappedErr) ([]byte, error)
	DecodeRsp(in []byte, rsp interface{}) (*WrappedErr, error)
	// One difficulty is, when server receives a req from client,
	// how can the serve know which type of struct to Unmarshal the req?
	// One solution is to store the (method, req type) in a map...
	// TODO: Study a solution to it
	// TODO: Include ctx in request messsages as well. Study how gRPC does this.
}

type jsonCodec struct{}

func (j *jsonCodec) EncodeReq(method string, v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	methodByte := []byte(method)
	vByte, err := json.Marshal(v)
	if err != nil {
		fmt.Printf("EncodeReq failed, err=%v", err)
		return nil, err
	}

	// binary.Write(buf, binary.BigEndian, uint8(len(methodByte)))
	// binary.Write(buf, binary.BigEndian, uint32(len(vByte)))

	len1 := byte(len(methodByte))
	len2 := uint32(len(vByte))
	buf.Write([]byte{
		len1,

		byte(len2),
		byte(len2 >> 8),
		byte(len2 >> 16),
		byte(len2 >> 24),
	})

	_, err = buf.Write(methodByte)
	if err != nil {
		fmt.Printf("buf.Write(methodByte) failed, err=%v", err)
		return nil, err
	}
	_, err = buf.Write(vByte)
	if err != nil {
		fmt.Printf("buf.Write(vByte) failed, err=%v", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func (j *jsonCodec) DecodeReq(in []byte) (string, interface{}, error) {
	if len(in) < 6 {
		return "", nil, errors.New("length of req must be at least 5 bytes")
	}
	methodLen := uint8(in[0])

	method := string(in[5 : 5+methodLen])
	reqByte := in[5+methodLen:]

	var req interface{}
	if reqType, ok := method2ReqType[method]; !ok {
		return "", nil, errors.New("DecodeReq Failed: Unknown method")
	} else {
		req = reflect.New(reqType).Interface() // req is a pointer
	}

	err := json.Unmarshal(reqByte, req)
	if err != nil {
		fmt.Printf("DecodeReq json.Unmarshal failed, err=%v \n", err)
		return "", nil, err
	}

	return method, req, nil

}

func (j *jsonCodec) EncodeRsp(rsp interface{}, rappedE *WrappedErr) ([]byte, error) {
	buf := new(bytes.Buffer)

	rspByte, err := json.Marshal(rsp)
	if err != nil {
		fmt.Printf("EncodeRsp json Marshal rspByte failed, err=%v \n", err)
		return nil, err
	}

	rappedEByte, err := json.Marshal(rappedE)
	if err != nil {
		fmt.Printf("EncodeRsp json Marshal rappedEByte failed, err=%v \n", err)
		return nil, err
	}

	// TODO: replace binary.Write with a more efficient one. (avoid using reflection)
	// binary.Write(buf, binary.LittleEndian, uint32(len(rspByte)))
	// binary.Write(buf, binary.LittleEndian, uint32(len(rappedEByte)))
	len1 := uint32(len(rspByte))
	len2 := uint32(len(rappedEByte))
	buf.Write([]byte{
		byte(len1),
		byte(len1 >> 8),
		byte(len1 >> 16),
		byte(len1 >> 24),

		byte(len2),
		byte(len2 >> 8),
		byte(len2 >> 16),
		byte(len2 >> 24),
	})

	_, err = buf.Write(rspByte)
	if err != nil {
		fmt.Printf("EncodeRsp buf.Write rspByte failed, err=%v", err)
		return nil, err
	}

	_, err = buf.Write(rappedEByte)
	if err != nil {
		fmt.Printf("EncodeRsp buf.Write rappedEByte failed, err=%v", err)
		return nil, err
	}

	return buf.Bytes(), nil

}

// rsp must be a pointer type
func (j *jsonCodec) DecodeRsp(in []byte, rsp interface{}) (*WrappedErr, error) {

	rspLen := binary.LittleEndian.Uint32(in[0:4])

	err := json.Unmarshal(in[8:8+rspLen], rsp)
	if err != nil {
		fmt.Printf("json.Unmarshal(out[8:8+rspLen], v) failed, err=%v", err)
		return nil, err
	}

	var rappedE WrappedErr

	// p0 := &rappedE
	// p1 := &p0
	// p2 := &p1

	err = json.Unmarshal(in[8+rspLen:], &rappedE)
	if err != nil {
		fmt.Printf("json.Unmarshal(out[8+rspLen:], &rappedE) failed, err=%v", err)
		return nil, err
	}

	if rappedE.Msg == "" {
		return nil, nil
	}

	return &rappedE, nil
}

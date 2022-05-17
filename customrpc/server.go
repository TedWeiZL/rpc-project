package customrpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

var (
	wPool *workerPool
	cPool *carrierPool

	stopServer bool
	mu         sync.RWMutex // protecting stopServer

	jobs = make(chan job, 1024)
)

type job struct {
	payload []byte
	conn    net.Conn
}

type carrierPool struct {
	carriers []*carrier

	// The following fields are for graceful shutdown
	stopCPool bool
	mu        sync.RWMutex // protecting stopPool
	wg        sync.WaitGroup
}

func (cp *carrierPool) Stop() {
	cp.mu.Lock()
	cp.stopCPool = true
	cp.mu.Unlock()

	cp.wg.Wait()

	close(jobs)
}

type carrier struct {
	conn net.Conn

	parent *carrierPool
}

func (c *carrier) carry() {
	c.parent.wg.Add(1)
	defer c.parent.wg.Done()
	for {
		c.parent.mu.RLock()
		if c.parent.stopCPool {
			c.parent.mu.RUnlock()
			// runtime.Goexit()
			return // TODO: which is better Goexit or return?
		}
		c.parent.mu.RUnlock()

		req, err := c.getOneRequest()
		if err != nil {
			fmt.Printf("c.getOneRequest() failed, err=%v\n", err)
			if err.Error() == "EOF" {
				return
			}
			continue
		}

		jobs <- job{payload: req, conn: c.conn}
	}
}

func (c *carrier) getOneRequest() ([]byte, error) {
	methondNameLenByte := make([]byte, 1, 1)
	_, err := c.conn.Read(methondNameLenByte)
	if err != nil {
		fmt.Printf("c.conn.Read(methondNameLenByte) failed, err=%v\n", err)
		return nil, err
	}
	methodNameLen := uint32(methondNameLenByte[0])

	reqLenByte := make([]byte, 4, 4)
	_, err = c.conn.Read(reqLenByte)
	if err != nil {
		fmt.Printf("c.conn.Read(reqLenByte) failed, err=%v\n", err)
		return nil, err
	}
	reqLen := binary.LittleEndian.Uint32(reqLenByte)

	buf := make([]byte, 5+methodNameLen+reqLen, 5+methodNameLen+reqLen)
	copy(buf[0:1], methondNameLenByte)
	copy(buf[1:5], reqLenByte)

	_, err = c.conn.Read(buf[5:])
	if err != nil {
		fmt.Printf("c.conn.Read(buf[5:]) failed, err=%v\n", err)
		return nil, err
	}

	return buf, nil
}

func InitServer(handler ServerToImplement, port, workerCnt int) {
	// register handlers
	registerhandler(handler)

	// init worker pool
	if workerCnt <= 0 || workerCnt > 40 {
		workerCnt = 5
	}
	wPool = initWorkerPool(workerCnt)

	// init carrier pool
	cPool = &carrierPool{}
}

func Serve(lis net.Listener) {
	for {
		// for graceful shutdown
		mu.RLock()
		if stopServer {
			mu.RUnlock()
			lis.Close()
			return
		}
		mu.RLock()

		conn, err := lis.Accept()
		if err != nil {
			// fmt.Printf("lis.Accept failed, err = %v \n", err)
			continue
		}

		c := &carrier{conn: conn, parent: cPool}
		cPool.carriers = append(cPool.carriers, c)

		go c.carry()
	}
}

func initWorkerPool(poolSize int) *workerPool {
	workerPool := workerPool{
		make([]*worker, 0, poolSize),
	}
	for i := 0; i <= poolSize-1; i++ {
		w := &worker{
			index: i,
			codec: &jsonCodec{},
		}
		workerPool.workers = append(workerPool.workers, w)
		go w.work()
	}

	return &workerPool
}

func GracefulShutdownServer() {
	// Gracefully shut down server steps:
	// 1. Close the listener to stop accept-ing new TCP connections
	// 2. Carry the very last message from existing TCP connections to the channel jobs
	// 4. Shutdown carriers
	// 5. close jobs
	// 6. As workers are processing the remainging jobs, a worker will close the job's conn after handling.
	// 7. Shutdown workers.
	mu.Lock()
	stopServer = true
	mu.Unlock()

	cPool.Stop()
}

type workerPool struct {
	workers []*worker
}

type worker struct {
	index int
	codec
}

func (w *worker) work() {
	for {
		job, ok := <-jobs
		if !ok {
			fmt.Printf("jobs has been closed, worker %d exits \n", w.index) // TODO: use log instead of print
			// runtime.Goexit()
			return
		}

		method, req, err := w.codec.DecodeReq(job.payload)
		if err != nil {
			fmt.Printf("w.codec.DecodeReq(job.payload) failed, err = %v \n", err)
			continue
		}

		// Actual Handling
		hanlder := method2Handler[method]
		rsp, err := hanlder(context.Background(), req)

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		encodedRsp, err := w.codec.EncodeRsp(rsp, &WrappedErr{Msg: errMsg})
		if err != nil {
			fmt.Printf("w.codec.EncodeRsp failed, err=%v\n", err)
			continue
		}

		_, err = job.conn.Write(encodedRsp)
		if err != nil {
			fmt.Printf("job.conn.Write(encodedRsp) failed, err=%v \n", err)
			continue
		}

		// Close the TCP connection once we stopServer is true
		//
		// Initially I thought there could be duplicate conns in jobs and hence I must only close the _last_ conn that share the same RemoteAddr.
		// Actually, since our communication is synchronous - the client will not send another request before getting a response from the server,
		// duplicate conns cannot exist in jobs.
		// Therefore, we can simply close the conn once stopServer is true.
		mu.RLock()
		if stopServer {
			job.conn.Close()
		}
		mu.RUnlock()
	}
}

/*
 * The IPC server handles requests from the IPC clients and response in
 * a defined way.
 * Work flow:
 * ---------------------  starts       ----------------------
 * | StartListening    | ------------> | handleRequest      |
 * | - starts server   |  go routine   | - react on changes |
 * | - accept requests |               |   in the req chan  |
 * | - put new req in  |               | - call req         |
 * |   buffered chan   |               |   callback         |
 * ---------------------               ----------------------
 * Remark: StartListening should be called in a go routine as well
 */
package ipc

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dachunky/crestrontcpbridge/pkg/logging"
	"github.com/google/uuid"
)

type clientRequest struct {
	cc   *ClientCommand
	id   string
	conn net.Conn
}

// IpcServer wraps needed interfaces for the IPC server instance
type IpcServer interface {
	// StartListen to a specific port
	StartListening(cmdHdl func(cc *ClientCommand) (*ServerResponse, error))
	// Requests received from IPC clients
	Requests() chan *clientRequest
	// HasError
	HasError() error
	// Close the IPC server
	Close()
}

type ipcServer struct {
	port          int
	clients       map[string]*ClientCommand
	requests      chan *clientRequest
	serverClosing chan bool
	err           error
	listener      net.Listener
	quit          chan bool
	wg            sync.WaitGroup
}

func (is *ipcServer) setError(err error) {
	logging.LogFmt(logging.LOG_ERROR, "[IPCSERVER] detect error: %v", err)
	is.err = err
}

// handleRequests is the central request handler for all clients
func (is *ipcServer) handleRequests(cmdHdl func(cc *ClientCommand) (*ServerResponse, error)) {
	is.wg.Add(1)
	defer is.wg.Done()
	for {
		select {
		case req := <-is.requests:
			logging.LogFmt(logging.LOG_INFO, "[handler]: receive request [%s] on channel --> calling command handler", req.id)
			sr, err := cmdHdl(req.cc)
			if err != nil {
				is.setError(err)
				continue
			}
			is.clients[sr.ID] = req.cc
			respData, err := sr.GetResponse2Send()
			if err != nil {
				is.setError(err)
				continue
			}
			_, err = req.conn.Write(respData)
			if err != nil {
				is.setError(err)
				continue
			}
			logging.LogFmt(logging.LOG_MAIN, "[handler]: request [%s] successfully reponded", req.id)
		case <-is.serverClosing:
			logging.Log(logging.LOG_INFO, "[handler]: receive termination signal")
			return
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// serveClient is started for each client. incoming request were push to the request channel and
// handled by handleRequests method
func (is *ipcServer) serveClient(cr *clientRequest) {
	defer func() {
		logging.LogFmt(logging.LOG_INFO, "[IPCSERVER] stop serving [%s]", cr.id)
		cr.conn.Close()
		logging.Log(logging.LOG_INFO, "[IPCSERVER] connection closed")
		is.wg.Done()
		logging.Log(logging.LOG_INFO, "[IPCSERVER] serve client excaped")
	}()
	for {
		buf, err := ReadUntilEOF(bufio.NewReader(cr.conn)) //.ReadBytes(0)
		if (err != nil) || (len(buf) < 1) {
			is.setError(err)
			return
		}
		if string(buf) == CLIENT_QUIT_COMMAND {
			logging.LogFmt(logging.LOG_INFO, "[IPCSERVER] client [%s] unregister", cr.id)
			return
		}
		cc, err := ClientCommandFromRequest(buf)
		if err != nil {
			is.setError(err)
			return
		}
		cr.cc = cc
		is.requests <- cr
	}
}

func (is *ipcServer) StartListening(cmdHdl func(cc *ClientCommand) (*ServerResponse, error)) {
	connStr := fmt.Sprintf(":%d", is.port) // localhost
	// create tcp server
	l, err := net.Listen("tcp", connStr)
	if err != nil {
		is.err = err
		return
	}
	is.listener = l
	is.wg.Add(1)
	// start request handler
	go is.handleRequests(cmdHdl)
	defer func() {
		logging.Log(logging.LOG_MAIN, "closing IPC server")
		// tell request handler to close
		is.serverClosing <- true
		is.wg.Done()
	}()
	logging.LogFmt(logging.LOG_MAIN, "start listening to: %s", connStr)
	for {
		// waiting for new requests
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-is.quit:
				// accept failed due to l.close() call
				logging.Log(logging.LOG_DEBUG, "[IPCSERVER]: receive quit --> stop listening")
				return
			default:
				// unexpected accept error
				is.setError(err)
			}
			continue
		}
		logging.Log(logging.LOG_DEBUG, "[IPCSERVER]: new client register request")
		cr := new(clientRequest)
		cr.conn = conn
		cr.id = uuid.NewString()
		is.wg.Add(1)
		// start to serve new client
		go is.serveClient(cr)
	}
}

func (is *ipcServer) Requests() chan *clientRequest {
	return is.requests
}

func (is *ipcServer) HasError() error {
	return is.err
}

func (is *ipcServer) Close() {
	logging.Log(logging.LOG_DEBUG, "[IPCSERVER] CLOSE: close ipc server")
	if is.listener == nil {
		return
	}
	logging.Log(logging.LOG_DEBUG, "[IPCSERVER] CLOSE: close listener")
	is.listener.Close()
	logging.Log(logging.LOG_DEBUG, "[IPCSERVER] CLOSE: quit 'is' channel")
	is.quit <- true
	// time.Sleep(100 * time.Millisecond)
	logging.Log(logging.LOG_DEBUG, "[IPCSERVER] CLOSE: wait for dependend routines to finish")
	is.wg.Wait()
	logging.Log(logging.LOG_DEBUG, "[IPCSERVER] CLOSE: close quit channel")
	close(is.quit)
	logging.Log(logging.LOG_DEBUG, "[IPCSERVER] CLOSE: DONE")
}

func NewIpcServer(port int) IpcServer {
	ret := new(ipcServer)
	ret.clients = make(map[string]*ClientCommand)
	ret.port = port
	ret.requests = make(chan *clientRequest, 256)
	ret.serverClosing = make(chan bool)
	ret.quit = make(chan bool)
	return ret
}

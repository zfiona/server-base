package websocket

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/zfiona/server-base/chanrpc"
	"github.com/zfiona/server-base/log"
	"github.com/zfiona/server-base/network"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Addr            string
	MaxConnNum      int32
	PendingWriteNum int32
	MaxMsgLen       uint32
	HTTPTimeout     time.Duration
	CertFile        string
	KeyFile         string
	ln              net.Listener
	handler         *WSHandler

	exitChan  chan struct{}
	waitGroup *sync.WaitGroup

	// msg parser
	AgentChanRPC          *chanrpc.Server       //handle rpc msg
	Processor             network.Processor     //handle json or pb
}

type WSHandler struct {
	server    *Server
	upgrade   websocket.Upgrader
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Debug("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.server.MaxMsgLen))

	if handler.server.getConnsNum()==0{
		_= conn.Close()
		log.Error("too many connections")
		return
	}
	newConn(conn, handler.server).run()
}

func (s *Server) Start() {
	log.Release("Launch Webserver,%v",s.Addr)
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if s.MaxConnNum <= 0 {
		s.MaxConnNum = 100
		log.Release("invalid MaxConnNum, reset to %v", s.MaxConnNum)
	}
	if s.PendingWriteNum <= 0 {
		s.PendingWriteNum = 100
		log.Release("invalid PendingWriteNum, reset to %v", s.PendingWriteNum)
	}
	if s.HTTPTimeout <= 0 {
		s.HTTPTimeout = 10 * time.Second
		log.Release("invalid Timeout, reset to %v", s.HTTPTimeout)
	}

	if s.CertFile != "" || s.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"gameweb/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
		if err != nil {
			log.Fatal("%v", err)
		}

		ln = tls.NewListener(ln, config)
	}

	s.ln = ln
	s.MaxMsgLen = 4096
	s.exitChan = make(chan struct{})
	s.waitGroup = &sync.WaitGroup{}
	s.handler = &WSHandler{
		server: s,
		upgrade: websocket.Upgrader{
			HandshakeTimeout: s.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		Addr:           s.Addr,
		Handler:        s.handler,
		ReadTimeout:    s.HTTPTimeout,
		WriteTimeout:   s.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go func() {
		_=httpServer.Serve(ln)
	}()
}

func (s *Server) Close() {
	close(s.exitChan)
	_= s.ln.Close()
	s.waitGroup.Wait()
}

func (s *Server) setConnsNum(num int32) {
	atomic.AddInt32(&s.MaxConnNum,num)
}

func (s *Server) getConnsNum() int32 {
	return atomic.LoadInt32(&s.MaxConnNum)
}

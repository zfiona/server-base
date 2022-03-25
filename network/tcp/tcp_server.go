package tcp

import (
	"github.com/zfiona/server-base/chanrpc"
	"github.com/zfiona/server-base/log"
	"github.com/zfiona/server-base/network"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Addr            string
	MaxConnNum      int32
	PendingWriteNum int32
	ln              net.Listener
	waitGroup        *sync.WaitGroup
	exitChan         chan struct{}

	// msg parser
	AgentChanRPC          *chanrpc.Server       //handle rpc msg
	Processor             network.Processor     //handle json or pb
	MsgParser             *network.MsgParser    //handle read and write
}

func (s *Server) Start() {
	s.init()
	go s.run()
}

func (s *Server) init() {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if s.PendingWriteNum <= 0 {
		s.PendingWriteNum = 1024
		log.Release("invalid PendingWriteNum, reset to %v", s.PendingWriteNum)
	}
	if s.AgentChanRPC == nil {
		log.Fatal("NewAgent must not be nil")
	}
	if s.MsgParser == nil{
		log.Fatal("should set MsgParser first")
	}
	if s.Processor == nil{
		log.Fatal("should set Processor first")
	}
	s.ln = ln
	s.exitChan = make(chan struct{})
	s.waitGroup = &sync.WaitGroup{}
}

func (s *Server) run() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	var tempDelay time.Duration
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Release("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		if s.getConnsNum()==0{
			log.Error("too many connections")
			_= conn.Close()
			continue
		}
		newConn(conn, s).run()
	}
}

func (s *Server) setConnsNum(num int32) {
	tmp := atomic.LoadInt32(&s.MaxConnNum) + num
	atomic.StoreInt32(&s.MaxConnNum,tmp)
}

func (s *Server) getConnsNum() int32 {
	return atomic.LoadInt32(&s.MaxConnNum)
}

func (s *Server) Close() {
	close(s.exitChan)
	_= s.ln.Close()
	s.waitGroup.Wait()
}

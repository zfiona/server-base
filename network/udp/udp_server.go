package udp

import (
	"github.com/xtaci/kcp-go"
	"github.com/zfiona/server-base/chanrpc"
	"github.com/zfiona/server-base/log"
	"github.com/zfiona/server-base/network"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Addr             string
	MaxConnNum       int32
	SendChanLimit    int32           // the limit of packet send channel
	RecChanLimit     int32           // the limit of packet receive channel
	ConnReadTimeout  time.Duration // read timeout
	ConnWriteTimeout time.Duration // write timeout
	ln               net.Listener
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
	l, err := kcp.Listen(s.Addr)
	if nil != err {
		return
	}
	conn, err := l.Accept()
	if err != nil {
		log.Error(err.Error())
		return
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
	s.ln = l
	s.exitChan = make(chan struct{})
	s.waitGroup = &sync.WaitGroup{}

	//kcp setting
	kcpConn := conn.(*kcp.UDPSession)
	kcpConn.SetNoDelay(1, 50, 1, 1)
	kcpConn.SetStreamMode(true)
	kcpConn.SetACKNoDelay(true)
	kcpConn.SetWindowSize(4096, 4096)
	_=kcpConn.SetReadBuffer(4 * 1024 * 1024)
	_=kcpConn.SetWriteBuffer(4 * 1024 * 1024)
}

func (s *Server) run() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			continue
		}

		if s.getConnsNum()==0{
			log.Error("too many connections")
			_= conn.Close()
			continue
		}
		NewConn(conn, s).run()
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
	_=s.ln.Close()
	s.waitGroup.Wait()
}



package udp

import (
	"github.com/zfiona/server-base/log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	server    *Server
	conn      net.Conn         // the raw connection
	userData  interface{}      // to save extra data
	closeFlag int32            // close flag
	closeOnce sync.Once        // close conn, once, per instance
	closeChan chan struct{}    // close channel
	sendChan  chan []byte      // packet send channel
	recChan   chan interface{} // packet rec channel
}

func NewConn(conn net.Conn, srv *Server) *Conn {
	c := &Conn{
		server:    srv,
		conn:      conn,
		closeChan: make(chan struct{}),
		sendChan:  make(chan []byte, srv.SendChanLimit),
		recChan:   make(chan interface{}, srv.RecChanLimit),
	}
	if srv.AgentChanRPC != nil {
		srv.AgentChanRPC.Go("NewAgent", c)
	}
	c.server.setConnsNum(-1)
	return c
}

func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.sendChan)
		close(c.recChan)
		_=c.conn.Close()
		if c.server.AgentChanRPC != nil {
			_= c.server.AgentChanRPC.Call0("CloseAgent", c)
		}
		c.server.setConnsNum(1)
	})
}

func (c *Conn) run() {
	asyncDo(c.handleLoop, c.server.waitGroup)
	asyncDo(c.readLoop, c.server.waitGroup)
	asyncDo(c.writeLoop, c.server.waitGroup)
}

func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}

func (c *Conn) WriteMsg(msg interface{}) {
	if c.IsClosed() {
		log.Error("ErrConnClosing")
		return
	}

	p, err := c.server.MsgParser.WriteMsg(c.server.Processor,msg)
	if err != nil{
		log.Error("WriteError,%v",err.Error())
		return
	}

	select {
	case c.sendChan <- p:
		return
	case <-c.closeChan:
		log.Error("ErrConnClosing")
		return
	case <-time.After(time.Second):
		log.Error("ErrWriteBlocking")
		return
	}
}

func (c *Conn) writeLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		case p := <-c.sendChan:
			if c.IsClosed() {
				return
			}
			err := c.conn.SetWriteDeadline(time.Now().Add(c.server.ConnWriteTimeout))
			if err != nil {
				log.Error("writeLoop, %d",err.Error())
				return
			}
			_, err = c.conn.Write(p)
			if err != nil {
				log.Error("writeLoop, %d",err.Error())
				return
			}
		}
	}

}

func (c *Conn) readLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		default:
		}

		err := c.conn.SetReadDeadline(time.Now().Add(c.server.ConnReadTimeout))
		if err != nil {
			log.Error("readLoop, %d",err.Error())
			return
		}

		p, err := c.server.MsgParser.ReadMsg(c.server.Processor,c.conn)
		if err != nil {
			log.Error("readLoop, %d",err.Error())
			return
		}
		c.recChan <- p
	}
}

func (c *Conn) handleLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		case p := <-c.recChan:
			if c.IsClosed() {
				return
			}
			err := c.server.Processor.Route(p,c)
			if err != nil {
				log.Error("route message error: %v", err)
				return
			}
		}
	}
}

func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

func (c *Conn) UserData() interface{} {
	return c.userData
}

func (c *Conn) SetUserData(data interface{}) {
	c.userData = data
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
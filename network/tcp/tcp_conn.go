package tcp

import (
	"github.com/zfiona/server-base/log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
)

type Conn struct {
	closeFlag int32            // close flag
	closeOnce sync.Once        // close conn, once, per instance
	closeChan chan struct{}    // close channel
	sendChan  chan []byte
	server    *Server
	conn      net.Conn
	userData  interface{}      // to save extra data
}

func newConn(conn net.Conn,srv *Server) *Conn {
	c := &Conn{
		server:   srv,
		conn:     conn,
		closeChan: make(chan struct{}),
		sendChan: make(chan []byte, srv.PendingWriteNum),
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
		_= c.conn.Close()
		if c.server.AgentChanRPC != nil {
			c.server.AgentChanRPC.Go("CloseAgent", c)
		}
		c.server.setConnsNum(1)
	})
}

func (c *Conn) run() {
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
		return
	}
	data, err := c.server.MsgParser.WriteMsg(c.server.Processor,msg)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
		return
	}
	if len(c.sendChan) == cap(c.sendChan) {
		log.Debug("close conn: channel full")
		c.Close()
		return
	}

	c.sendChan <- data
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
		case d := <-c.sendChan:
			_, err := c.conn.Write(d)
			if err != nil {
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

		msg, err := c.server.MsgParser.ReadMsg(c.server.Processor,c.conn)
		if err != nil {
			log.Error("read message error: %v", err)
			return
		}
		err = c.server.Processor.Route(msg, c)
		if err != nil {
			log.Error("route message error: %v", err)
			return
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

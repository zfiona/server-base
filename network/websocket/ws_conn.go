package websocket

import (
	"github.com/gorilla/websocket"
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
	writeChan chan []byte
	server    *Server
	conn      *websocket.Conn
	userData  interface{}      // to save extra data
}

func newConn(conn *websocket.Conn, srv *Server) *Conn {
	c := &Conn{
		server:   srv,
		conn:     conn,
		closeChan: make(chan struct{}),
		writeChan: make(chan []byte, srv.PendingWriteNum),
	}
	c.server.setConnsNum(-1)
	if srv.AgentChanRPC != nil {
		srv.AgentChanRPC.Go("NewAgent", c)
	}
	return c
}

func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.writeChan)
		_= c.conn.Close()
		if c.server.AgentChanRPC != nil {
			_= c.server.AgentChanRPC.Call0("CloseAgent", c)
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
	if c.IsClosed(){
		return
	}

	data, err := c.server.Processor.Marshal(msg)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(msg), err)
		return
	}
	c.writeChan <- data
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
		case b := <-c.writeChan:
			err := c.conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				break
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

		_, b, err := c.conn.ReadMessage()
		if err != nil {
			log.Error("read message: %v", err)
			break
		}
		msg, err := c.server.Processor.Unmarshal(b)
		if err != nil {
			log.Error("unmarshal message error: %v", err)
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

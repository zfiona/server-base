package gate

import (
	"net"
)

type Agent interface {
	WriteMsg(msg interface{})
	Close()
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	UserData() interface{}
	SetUserData(data interface{})
}

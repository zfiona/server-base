package gate

import (
	"github.com/zfiona/server-base/chanrpc"
	"github.com/zfiona/server-base/network"
	"github.com/zfiona/server-base/network/tcp"
	"github.com/zfiona/server-base/network/udp"
	"github.com/zfiona/server-base/network/websocket"
	"time"
)

type Gate struct {
	MaxConnNum      int32
	PendingWriteNum int32
	Timeout         time.Duration
	Processor       network.Processor
	MsgParser       *network.MsgParser
	AgentChanRPC    *chanrpc.Server

	//websocket
	WSAddr   string
	CertFile string
	KeyFile  string
	//tcp
	TCPAddr  string
	//udp
	UDPAddr  string
}

func (gate *Gate) Run(closeSig chan bool) {
	if gate.WSAddr != "" {
		gate.RunWebServer(closeSig)
	}
	if gate.TCPAddr != "" {
		gate.RunTcpServer(closeSig)
	}
	if gate.UDPAddr != ""{
		gate.RunUdpServer(closeSig)
	}
}

func (gate *Gate) RunWebServer(closeSig chan bool){
	server := new(websocket.Server)
	server.Addr = gate.WSAddr
	server.MaxConnNum = gate.MaxConnNum
	server.PendingWriteNum = gate.PendingWriteNum
	server.HTTPTimeout = gate.Timeout
	server.CertFile = gate.CertFile
	server.KeyFile = gate.KeyFile

	server.AgentChanRPC = gate.AgentChanRPC
	server.Processor = gate.Processor

	server.Start()
	<-closeSig
	server.Close()
}

func (gate *Gate) RunTcpServer(closeSig chan bool){
	server := new(tcp.Server)
	server.Addr = gate.TCPAddr
	server.MaxConnNum = gate.MaxConnNum
	server.PendingWriteNum = gate.PendingWriteNum

	server.AgentChanRPC = gate.AgentChanRPC
	server.Processor = gate.Processor
	server.MsgParser = gate.MsgParser

	server.Start()
	<-closeSig
	server.Close()
}

func (gate *Gate) RunUdpServer(closeSig chan bool){
	server := new(udp.Server)
	server.Addr = gate.UDPAddr
	server.MaxConnNum = gate.MaxConnNum
	server.SendChanLimit = gate.PendingWriteNum
	server.RecChanLimit = gate.PendingWriteNum
	server.ConnReadTimeout = gate.Timeout
	server.ConnWriteTimeout = gate.Timeout

	server.AgentChanRPC = gate.AgentChanRPC
	server.Processor = gate.Processor
	server.MsgParser = gate.MsgParser

	server.Start()
	<-closeSig
	server.Close()
}

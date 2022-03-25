package network

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)
//--------------
// | len + cmd + data |
//    2  +  2  +  n
// --------------

type MsgParser struct {
	littleEndian bool
	lenMsgLen    byte
	lenMsgId     byte
	maxMsgLen    uint32
}

func (m *MsgParser) SetByteOrder(littleEndian bool) {
	m.littleEndian = littleEndian
}

func (m *MsgParser) SetMsgLen(lenMsgLen,lenMsgId int) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		m.lenMsgLen = byte(lenMsgLen)
	}
	if lenMsgId == 1 || lenMsgId == 2 || lenMsgId == 4 {
		m.lenMsgId = byte(lenMsgId)
	}

	switch m.lenMsgLen {
	case 1:
		m.maxMsgLen = math.MaxUint8
	case 2:
		m.maxMsgLen = math.MaxUint16
	case 4:
		m.maxMsgLen = math.MaxUint32
	}
}

func (m *MsgParser) ReadMsg(p Processor,r io.Reader) (interface{}, error) {
	// read len
	bufMsgLen := make([]byte,m.lenMsgLen)
	if _, err := io.ReadFull(r, bufMsgLen); err != nil {
		return nil, err
	}
	// parse len
	var msgLen uint32
	switch m.lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		if m.littleEndian {
			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
		}
	case 4:
		if m.littleEndian {
			msgLen = binary.LittleEndian.Uint32(bufMsgLen)
		} else {
			msgLen = binary.BigEndian.Uint32(bufMsgLen)
		}
	}

	// check len
	if msgLen > m.maxMsgLen {
		return nil, errors.New("message too long")
	}
	// data
	data := make([]byte, msgLen + uint32(m.lenMsgId))
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return p.Unmarshal(data)
}

func (m *MsgParser) WriteMsg(p Processor,msg interface{}) ([]byte, error) {
	//id & data
	data, err := p.Marshal(msg)
	if err != nil {
		return nil,err
	}
	msgLen := uint32(len(data)) - uint32(m.lenMsgId)

	// write len
	buf := make([]byte, m.lenMsgLen)
	switch m.lenMsgLen {
	case 1:
		buf[0] = byte(msgLen)
	case 2:
		if m.littleEndian {
			binary.LittleEndian.PutUint16(buf, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(buf, uint16(msgLen))
		}
	case 4:
		if m.littleEndian {
			binary.LittleEndian.PutUint32(buf, msgLen)
		} else {
			binary.BigEndian.PutUint32(buf, msgLen)
		}
	}
	//buf
	buf = append(buf,data...)
	return buf,nil
}

package status

import (
	"net"
	"strconv"

	"github.com/Tnze/go-mc/net/packet"
)

const (
	HandshakeStatusState = packet.VarInt(1)
	HandshakeLoginState  = packet.VarInt(2)

	ForgeSeparator  = "\x00"
	RealIPSeparator = "///"
)

func UniversalStatusHandshake(addr string) (Handshake, error) {
	return NewHandshake(addr, 0, 1)
}

func NewHandshake(addr string, protocol, state packet.VarInt) (Handshake, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return Handshake{}, err
	}
	port, err := strconv.ParseUint(portStr, 0, 16)
	if err != nil {
		return Handshake{}, err
	}
	return Handshake{
		ProtocolVersion: protocol,
		ServerAddress:   packet.String(host),
		ServerPort:      packet.UnsignedShort(port),
		NextState:       state,
	}, nil
}

type Handshake struct {
	ProtocolVersion packet.VarInt
	ServerAddress   packet.String
	ServerPort      packet.UnsignedShort
	NextState       packet.VarInt
}

func (pk Handshake) Marshal() packet.Packet {
	return packet.Marshal(
		0x00,
		pk.ProtocolVersion,
		pk.ServerAddress,
		pk.ServerPort,
		pk.NextState,
	)
}

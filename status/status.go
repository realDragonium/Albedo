package status

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	mcnet "github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
	"github.com/gofrs/uuid"
)

func StatusProtocolNumber(addr, hostname string) (int, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	port, err := strconv.ParseUint(portStr, 0, 16)
	if err != nil {
		return 0, err
	}
	conn, err := mcnet.DialMCTimeout(addr, time.Second)
	if err != nil {
		return 0, err
	}
	handshakePacket := Handshake{
		ProtocolVersion: 754,
		ServerAddress:   packet.String(hostname),
		ServerPort:      packet.UnsignedShort(port),
		NextState:       1,
	}.Marshal()
	log.Println("Sending handshake")
	err = conn.WritePacket(handshakePacket)
	if err != nil {
		return 0, err
	}

	err = conn.WritePacket(packet.Marshal(
		packetid.PingStart,
	))
	if err != nil {
		return 0, err
	}
	log.Println("Waiting on answer")
	var p packet.Packet
	if err := conn.ReadPacket(&p); err != nil {
		return 0, err
	}
	log.Println("Received answer")
	var s packet.String
	err = p.Scan(&s)
	if err != nil {
		return 0, nil
	}

	//Ping
	startTime := time.Now()
	err = conn.WritePacket(packet.Marshal(
		packetid.PingServerbound,
		packet.Long(startTime.Unix()),
	))
	if err != nil {
		fmt.Printf("bot: send ping packect fail: %v", err)
		return 0, nil
	}

	if err = conn.ReadPacket(&p); err != nil {
		fmt.Printf("bot: recv pong packect fail: %v", err)
		return 0, nil
	}
	var t packet.Long
	err = p.Scan(&t)
	if err != nil {
		fmt.Printf("bot: scan pong packect fail: %v", err)
		return 0, nil
	}
	if t != packet.Long(startTime.Unix()) {
		fmt.Printf("bot: pong packect no match: %v", err)
		return 0, nil
	}

	resp := []byte(s)
	var status Status
	err = json.Unmarshal(resp, &status)
	if err != nil {
		fmt.Print("unmarshal resp fail:", err)
		os.Exit(1)
	}

	return status.Version.Protocol, nil
}

type Status struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	//favicon ignored
}

func SendStatus() {
	addr := getAddr()
	fmt.Printf("MCPING (%s):\n", addr)
	number, err := StatusProtocolNumber(addr, "10.75.135.175")
	if err != nil {
		fmt.Printf("bot: send handshake packect fail: %v", err)
		return
	}
	fmt.Printf("protocol: %d \n", number)

}

func StatusSomething() {
	addr := getAddr()
	fmt.Printf("MCPING (%s):\n", addr)
	resp, delay, err := bot.PingAndList(addr)
	if err != nil {
		fmt.Printf("ping and list server fail: %v", err)
		os.Exit(1)
	}

	var s Status
	err = json.Unmarshal(resp, &s)
	if err != nil {
		fmt.Print("unmarshal resp fail:", err)
		os.Exit(1)
	}

	fmt.Print(s)
	fmt.Println("Delay:", delay)
}

func PrintServerStatus() {
	addr := getAddr()
	fmt.Printf("MCPING (%s):\n", addr)
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Printf("bot: send handshake packect fail: %v", err)
		return
	}
	port, err := strconv.ParseUint(portStr, 0, 16)
	if err != nil {
		fmt.Printf("bot: send handshake packect fail: %v", err)
		return
	}
	conn, err := mcnet.DialMCTimeout(addr, time.Second)
	if err != nil {
		fmt.Printf("bot: send handshake packect fail: %v", err)
		return
	}

	const Handshake = 0x00
	err = conn.WritePacket(packet.Marshal(
		Handshake,           //Handshake packet ID
		packet.VarInt(754),  //Protocol version
		packet.String(host), //Server's address
		packet.UnsignedShort(port),
		packet.Byte(1),
	))
	if err != nil {
		fmt.Printf("bot: send handshake packect fail: %v", err)
		return
	}

	err = conn.WritePacket(packet.Marshal(
		packetid.PingStart,
	))
	if err != nil {
		fmt.Printf("bot: send list packect fail: %v", err)
		return
	}

	var p packet.Packet
	if err := conn.ReadPacket(&p); err != nil {
		fmt.Printf("bot: recv list packect fail: %v", err)
		return
	}

	var s packet.String
	err = p.Scan(&s)
	if err != nil {
		fmt.Printf("bot: scan list packect fail: %v", err)
		return
	}

	//PING
	startTime := time.Now()
	err = conn.WritePacket(packet.Marshal(
		packetid.PingServerbound,
		packet.Long(startTime.Unix()),
	))
	if err != nil {
		fmt.Printf("bot: send ping packect fail: %v", err)
		return
	}

	if err = conn.ReadPacket(&p); err != nil {
		fmt.Printf("bot: recv pong packect fail: %v", err)
		return
	}
	var t packet.Long
	err = p.Scan(&t)
	if err != nil {
		fmt.Printf("bot: scan pong packect fail: %v", err)
		return
	}
	if t != packet.Long(startTime.Unix()) {
		fmt.Printf("bot: pong packect no match: %v", err)
		return
	}
	delay := time.Since(startTime)
	resp := []byte(s)

	var status Status
	err = json.Unmarshal(resp, &status)
	if err != nil {
		fmt.Print("unmarshal resp fail:", err)
		os.Exit(1)
	}

	fmt.Println(s)
	fmt.Println(status)
	fmt.Println("Delay:", delay)
}

func getAddr() string {
	const usage = "Usage: mcping <hostname>[:port]"
	if len(os.Args) < 2 {
		fmt.Println("no host name.", usage)
		os.Exit(1)
	}

	return os.Args[2]
}

func (s Status) String() string {
	var sb strings.Builder
	fmt.Fprintln(&sb, "Server:", s.Version.Name)
	fmt.Fprintln(&sb, "Protocol:", s.Version.Protocol)
	fmt.Fprintln(&sb, "Description:", s.Description)
	fmt.Fprintf(&sb, "Players: %d/%d\n", s.Players.Online, s.Players.Max)
	for _, v := range s.Players.Sample {
		fmt.Fprintf(&sb, "- [%s] %v\n", v.Name, v.ID)
	}
	return sb.String()
}

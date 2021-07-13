package status

import (
	"log"
	"time"

	"github.com/Tnze/go-mc/data/packetid"
	mcnet "github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

func StartSpam(number int, intervals time.Duration) {
	addr := "10.75.135.175"
	addrWithPort := "10.75.135.175:25565"
	protocol, err := StatusProtocolNumber(addrWithPort, addr)
	if err != nil {
		log.Printf("Error during getting protocol: %v", err)
	}

	infoMonitor := InfoMonitor{
		dataRequestCh: make(chan SpamDataRequest),
		intervals:     5 * time.Second,
	}
	log.Println("Starting up info monitor")
	go infoMonitor.Inform()
	log.Println("Creating Spam monitor")
	monitor := NewSpamMonitor(infoMonitor.dataRequestCh)
	go monitor.Monitor()

	handshakePacket := Handshake{
		ProtocolVersion: packet.VarInt(protocol),
		ServerAddress:   packet.String(addr),
		ServerPort:      25565,
		NextState:       1,
	}.Marshal()

	data := StatusSpamData{
		interval:      intervals,
		addr:          addrWithPort,
		handshake:     handshakePacket,
		notifyBeginCh: monitor.notifyBeginCh,
		notifyEndCh:   monitor.notifyEndCh,
		connCreatedCh: monitor.connCreatedCh,
	}
	log.Println("Starting spinning up workers now")
	for i := 0; i < number; i++ {
		go func(d StatusSpamData) {
			d.sendStatusRequestSpam()
		}(data)
	}
	select {}
}

type StatusSpamData struct {
	interval  time.Duration
	handshake packet.Packet
	addr      string

	notifyBeginCh chan struct{}
	notifyEndCh   chan struct{}
	connCreatedCh chan struct{}
}

func (data *StatusSpamData) sendStatusRequestSpam() {
	failCount := 0
	for {
		if failCount > 5 {
			log.Println("Failed to many times, existing for loop")
			return
		}
		time.Sleep(data.interval)
		data.notifyBeginCh <- struct{}{}
		conn, err := mcnet.DialMCTimeout(data.addr, time.Second)
		if err != nil {
			log.Printf("Error while creating connection: %v", err)
			failCount++
			return
			// contin/ue
		}
		data.connCreatedCh <- struct{}{}
		err = conn.WritePacket(data.handshake)
		if err != nil {
			failCount++
			continue
		}
		err = conn.WritePacket(packet.Marshal(
			packetid.PingStart,
		))
		if err != nil {
			// log.Printf("bot: send list packect fail: %v", err)
			failCount++
			continue
		}

		var p packet.Packet
		if err := conn.ReadPacket(&p); err != nil {
			// log.Printf("bot: recv list packect fail: %v", err)
			failCount++
			continue
		}

		var s packet.String
		err = p.Scan(&s)
		if err != nil {
			// log.Printf("bot: scan list packect fail: %v", err)
			failCount++
			continue
		}

		// var status Status
		// err = json.Unmarshal([]byte(s), &status)
		// if err != nil {
		// 	fmt.Print("unmarshal resp fail:", err)
		// 	os.Exit(1)
		// }

		// log.Print(s)

		//PING
		// startTime := time.Now()
		// err = conn.WritePacket(packet.Marshal(
		// 	packetid.PingServerbound,
		// 	packet.Long(startTime.Unix()),
		// ))
		// if err != nil {
		// 	fmt.Printf("bot: send ping packect fail: %v", err)
		// 	continue
		// }

		// if err = conn.ReadPacket(&p); err != nil {
		// 	fmt.Printf("bot: recv pong packect fail: %v", err)
		// 	continue
		// }
		// var t packet.Long
		// err = p.Scan(&t)
		// if err != nil {
		// 	fmt.Printf("bot: scan pong packect fail: %v", err)
		// 	continue
		// }
		// if t != packet.Long(startTime.Unix()) {
		// 	fmt.Printf("bot: pong packect no match: %v", err)
		// 	continue
		// }
		conn.Close()
		// delay := time.Since(startTime)
		data.notifyEndCh <- struct{}{}
	}
}

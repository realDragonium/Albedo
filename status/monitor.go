package status

import (
	"log"
	"time"
)

type SpamDataRequest struct {
	ch chan SpamData
}

type SpamData struct {
	begon        int
	ended        int
	connCreated  int
	hsWritten    int
	pingSend     int
	pingReceived int
}

func NewSpamMonitor(requestCh chan SpamDataRequest) SpamMonitor {
	return SpamMonitor{
		dataRequestCh:  requestCh,
		notifyBeginCh:  make(chan struct{}),
		notifyEndCh:    make(chan struct{}),
		connCreatedCh:  make(chan struct{}),
		hsWrittenCh:    make(chan struct{}),
		pingSendCh:     make(chan struct{}),
		pingReceivedCh: make(chan struct{}),
	}
}

type SpamMonitor struct {
	data          SpamData
	dataRequestCh chan SpamDataRequest

	notifyBeginCh  chan struct{}
	notifyEndCh    chan struct{}
	connCreatedCh  chan struct{}
	hsWrittenCh    chan struct{}
	pingSendCh     chan struct{}
	pingReceivedCh chan struct{}
}

func (monitor *SpamMonitor) Monitor() {
	for {
		select {
		case request := <-monitor.dataRequestCh:
			request.ch <- monitor.data
			monitor.data = SpamData{}
		case <-monitor.notifyBeginCh:
			monitor.data.begon++
		case <-monitor.notifyEndCh:
			monitor.data.ended++
		case <-monitor.connCreatedCh:
			monitor.data.connCreated++
		case <-monitor.hsWrittenCh:
			monitor.data.hsWritten++
		case <-monitor.pingSendCh:
			monitor.data.pingSend++
		case <-monitor.pingReceivedCh:
			monitor.data.pingReceived++
		}
	}
}

func (s SpamData) Display() {
	log.Printf("Request begon: %d", s.begon)
	log.Printf("Request ended: %d", s.ended)
	log.Printf("Connections opened: %d", s.connCreated)
}

type InfoMonitor struct {
	dataRequestCh chan SpamDataRequest
	intervals     time.Duration
}

func (info *InfoMonitor) Inform() {
	for {
		time.Sleep(info.intervals)
		dataCh := make(chan SpamData)
		info.dataRequestCh <- SpamDataRequest{
			ch: dataCh,
		}
		data := <-dataCh
		log.Printf("Since the last time (%s) there has been:", info.intervals.String())
		data.Display()

	}
}

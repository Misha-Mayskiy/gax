package sfu

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Peer struct {
	ID   string
	Conn *websocket.Conn
	Lock sync.Mutex

	PC                 *webrtc.PeerConnection
	NegotiationPending bool
	StreamIDs          []string
}

func (p *Peer) SendJSON(v interface{}) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	return p.Conn.WriteJSON(v)
}

func (p *Peer) Close() {
	if p.PC != nil {
		p.PC.Close()
	}
	if p.Conn != nil {
		p.Conn.Close()
	}
}

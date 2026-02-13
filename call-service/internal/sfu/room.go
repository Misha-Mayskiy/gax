package sfu

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type Room struct {
	ID     string
	Lock   sync.RWMutex
	Peers  map[string]*Peer
	Tracks []*webrtc.TrackLocalStaticRTP
}

func NewRoom(id string) *Room {
	return &Room{
		ID:     id,
		Peers:  make(map[string]*Peer),
		Tracks: make([]*webrtc.TrackLocalStaticRTP, 0),
	}
}

func (r *Room) AddPeer(peer *Peer) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	r.Peers[peer.ID] = peer
	log.Printf("[Room %s] Peer %s joined", r.ID, peer.ID)
}

func (r *Room) RemovePeer(peerID string) {
	r.Lock.Lock()
	peer, exists := r.Peers[peerID]
	if !exists {
		r.Lock.Unlock()
		return
	}
	delete(r.Peers, peerID)
	r.Lock.Unlock()

	// Чистим треки
	r.Lock.Lock()
	newTracks := make([]*webrtc.TrackLocalStaticRTP, 0)
	for _, track := range r.Tracks {
		shouldKeep := true
		for _, usersStreamID := range peer.StreamIDs {
			if track.StreamID() == usersStreamID {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newTracks = append(newTracks, track)
		}
	}
	r.Tracks = newTracks
	r.Lock.Unlock()

	peer.Close()
	log.Printf("[Room %s] Peer %s left", r.ID, peerID)

	// Уведомляем остальных
	r.Lock.RLock()
	for _, p := range r.Peers {
		go func(p *Peer) {
			p.SendJSON(map[string]interface{}{
				"type": "user_left",
				"payload": map[string]interface{}{
					"user_id":    peerID,
					"stream_ids": peer.StreamIDs,
				},
			})
		}(p)
	}
	r.Lock.RUnlock()
}

func (r *Room) Signal(peer *Peer) {
	peer.Lock.Lock()
	defer peer.Lock.Unlock()

	if peer.PC.ConnectionState() == webrtc.PeerConnectionStateClosed {
		return
	}

	if peer.PC.SignalingState() != webrtc.SignalingStateStable {
		log.Printf("[Peer %s] Signaling not stable, marking pending", peer.ID)
		peer.NegotiationPending = true
		return
	}

	offer, err := peer.PC.CreateOffer(nil)
	if err != nil {
		log.Printf("[Peer %s] CreateOffer error: %v", peer.ID, err)
		return
	}

	if err := peer.PC.SetLocalDescription(offer); err != nil {
		log.Printf("[Peer %s] SetLocalDescription error: %v", peer.ID, err)
		return
	}

	peer.NegotiationPending = false

	if err := peer.Conn.WriteJSON(map[string]interface{}{
		"type":    "offer",
		"payload": offer,
	}); err != nil {
		log.Printf("[Peer %s] WriteJSON error: %v", peer.ID, err)
	}
}

func (r *Room) Join(peer *Peer) error {
	// Кодеки
	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return err
	}

	// Сеть
	settingEngine := webrtc.SettingEngine{}

	// IPv4
	settingEngine.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeUDP4})

	publicIP := os.Getenv("PUBLIC_IP")
	if publicIP != "" {
		settingEngine.SetNAT1To1IPs([]string{publicIP}, webrtc.ICECandidateTypeHost)
	}

	// Порты
	settingEngine.SetEphemeralUDPPortRange(50000, 60000)

	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithSettingEngine(settingEngine),
	)

	// PeerConnection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{},
	}

	pc, err := api.NewPeerConnection(config)
	if err != nil {
		return err
	}
	peer.PC = pc

	// Трансиверы
	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})
	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})

	// Подписка на существующих
	r.Lock.RLock()
	for _, track := range r.Tracks {
		sender, err := peer.PC.AddTrack(track)
		if err != nil {
			log.Printf("AddTrack error: %v", err)
			continue
		}
		go func(sender *webrtc.RTPSender) {
			buf := make([]byte, 1500)
			for {
				if _, _, err := sender.Read(buf); err != nil {
					return
				}
			}
		}(sender)
	}
	r.Lock.RUnlock()

	// Обработка входящих
	peer.PC.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Peer %s sent track: %s", peer.ID, remoteTrack.Kind())

		peer.Lock.Lock()
		peer.StreamIDs = append(peer.StreamIDs, remoteTrack.StreamID())
		peer.Lock.Unlock()

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
		)
		if err != nil {
			log.Println("NewTrackLocal error:", err)
			return
		}

		r.Lock.Lock()
		r.Tracks = append(r.Tracks, localTrack)
		r.Lock.Unlock()

		r.Lock.RLock()
		for id, p := range r.Peers {
			if id == peer.ID {
				continue
			}
			if _, err := p.PC.AddTrack(localTrack); err != nil {
				continue
			}
			go r.Signal(p)
		}
		r.Lock.RUnlock()

		// RTP Forwarding
		go func() {
			buf := make([]byte, 1500)
			for {
				n, _, err := remoteTrack.Read(buf)
				if err != nil {
					return
				}
				localTrack.Write(buf[:n])
			}
		}()

		// PLI Ticker (1 sec)
		go func() {
			ticker := time.NewTicker(time.Second * 1)
			defer ticker.Stop()
			for range ticker.C {
				if peer.PC.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}}) != nil {
					return
				}
			}
		}()
	})

	peer.PC.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		peer.SendJSON(map[string]interface{}{"type": "candidate", "payload": c.ToJSON()})
	})

	peer.PC.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		log.Printf("Peer %s connection state: %s", peer.ID, p.String())
		if p == webrtc.PeerConnectionStateFailed || p == webrtc.PeerConnectionStateClosed {
			r.RemovePeer(peer.ID)
		}
	})

	r.AddPeer(peer)
	r.Signal(peer)
	return nil
}

package controller

import (
	"slices"
	"sync"

	"github.com/pion/webrtc/v4"
)

type P2PConnectionState struct {
	conn *webrtc.PeerConnection
	dc   *webrtc.DataChannel

	offer         *webrtc.SessionDescription
	answer        *webrtc.SessionDescription
	iceCandidates []webrtc.ICECandidateInit

	password string

	mu sync.RWMutex
}

// Безопасные методы доступа к состоянию
func (c *P2PConnectionState) Conn() *webrtc.PeerConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *P2PConnectionState) SetConn(conn *webrtc.PeerConnection) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn = conn
}

func (c *P2PConnectionState) DataChannel() *webrtc.DataChannel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dc
}

func (c *P2PConnectionState) SetDataChannel(dc *webrtc.DataChannel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dc = dc
}

func (c *P2PConnectionState) Offer() *webrtc.SessionDescription {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.offer
}

func (c *P2PConnectionState) SetOffer(offer *webrtc.SessionDescription) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offer = offer
}

func (c *P2PConnectionState) Answer() *webrtc.SessionDescription {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.answer
}

func (c *P2PConnectionState) SetAnswer(answer *webrtc.SessionDescription) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.answer = answer
}

func (c *P2PConnectionState) Password() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.password
}

func (c *P2PConnectionState) SetPassword(password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.password = password
}

func (c *P2PConnectionState) AddICECandidate(candidate webrtc.ICECandidateInit) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.iceCandidates = append(c.iceCandidates, candidate)
}

func (c *P2PConnectionState) ICECandidates() []webrtc.ICECandidateInit {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Clone(c.iceCandidates)
}

// Close освобождает ресурсы соединения
func (c *P2PConnectionState) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if c.dc != nil {
		if err := c.dc.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

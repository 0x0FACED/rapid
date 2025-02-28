package controller

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v4"
	"golang.org/x/crypto/bcrypt"
)

type EncodedOffer struct {
	// offer string
	SDP string `json:"sdp"`
	// hash of pass
	Hash string `json:"hash"`
}

type DecodedOffer struct {
	// offer string
	SDP webrtc.SessionDescription `json:"sdp"`
	// hash of pass
	Hash string `json:"hash"`
}

type EncodedAnswer struct {
	SDP  string `json:"sdp"`
	Hash string `json:"hash"`
}

type DecodedAnswer struct {
	SDP  webrtc.SessionDescription
	Hash string
}

var (
	ICEServers = []webrtc.ICEServer{
		{URLs: []string{"stun:stun.l.google.com:19302"}},
		{URLs: []string{"stun:stun.l.google.com:5349"}},
		{URLs: []string{"stun:stun1.l.google.com:3478"}},
		{URLs: []string{"stun:stun1.l.google.com:5349"}},
		{URLs: []string{"stun:stun2.l.google.com:19302"}},
		{URLs: []string{"stun:stun2.l.google.com:5349"}},
		{URLs: []string{"stun:stun3.l.google.com:3478"}},
		{URLs: []string{"stun:stun3.l.google.com:5349"}},
		{URLs: []string{"stun:stun4.l.google.com:19302"}},
	}
)

type P2PConnectionState struct {
	conn *webrtc.PeerConnection
	dc   *webrtc.DataChannel

	offer         *webrtc.SessionDescription
	answer        *webrtc.SessionDescription
	iceCandidates []webrtc.ICECandidateInit

	password string

	isConnected  atomic.Bool
	onConnect    func()
	onDisconnect func()
	onMessage    func([]byte)

	mu sync.RWMutex
}

func NewP2PConnectionState() (*P2PConnectionState, error) {
	// TODO: refactor
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{
		// hardcoded free ice servers list
		ICEServers: ICEServers,
		// tcp + udp
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
	})
	if err != nil {
		return nil, err
	}

	dc, err := conn.CreateDataChannel("rapid", nil)
	if err != nil {
		return nil, err
	}

	return &P2PConnectionState{
		conn:          conn,
		dc:            dc,
		iceCandidates: make([]webrtc.ICECandidateInit, 0),

		onConnect:    func() {},
		onDisconnect: func() {},
		onMessage:    func([]byte) {},
	}, nil
}

func (c *P2PConnectionState) Initialize() error {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers:         ICEServers,
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
	})
	if err != nil {
		return err
	}

	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		c.AddICECandidate(candidate.ToJSON())
	})

	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
			c.isConnected.Store(true)
			c.onConnect()
		case webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateFailed,
			webrtc.PeerConnectionStateClosed:
			c.isConnected.Store(false)
			c.onDisconnect()
		}
	})

	dc, err := conn.CreateDataChannel("rapid", &webrtc.DataChannelInit{
		Negotiated: BoolToPtr(true),
		ID:         Uint16ToPtr(0),
	})
	if err != nil {
		return err
	}

	dc.OnOpen(func() {
		fmt.Println("Data channel opened!")
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		c.onMessage(msg.Data)
	})

	c.conn = conn
	c.dc = dc
	return nil
}

// TODO: refactor
func BoolToPtr(val bool) *bool {
	return &val
}

// TODO: refactor
func Uint16ToPtr(val uint16) *uint16 {
	return &val
}

func (c *P2PConnectionState) Conn() *webrtc.PeerConnection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// TODO: refactor
func (c *P2PConnectionState) CreateEncodedOffer(opts *webrtc.OfferOptions) (string, error) {
	if c.conn != nil {
		err := c.Close()
		if err != nil {
			return "", err
		}
	}

	if err := c.Initialize(); err != nil {
		return "", err
	}

	offer, err := c.conn.CreateOffer(opts)
	if err != nil {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(c.Password()),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	data := EncodedOffer{
		SDP:  offer.SDP,
		Hash: string(hashedPassword),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	if err = c.conn.SetLocalDescription(offer); err != nil {
		return "", err
	}

	c.offer = &offer

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

func (c *P2PConnectionState) DecodeOffer(offer string) (*DecodedOffer, error) {
	decoded, err := base64.URLEncoding.DecodeString(offer)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("gzip reader create failed: %w", err)
	}
	defer gzReader.Close()

	uncompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("gzip decompress failed: %w", err)
	}

	var encodedOffer EncodedOffer
	if err := json.Unmarshal(uncompressed, &encodedOffer); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	if encodedOffer.SDP == "" {
		return nil, errors.New("empty SDP in offer")
	}

	sdp := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  encodedOffer.SDP,
	}

	decodedOffer := &DecodedOffer{
		SDP:  sdp,
		Hash: encodedOffer.Hash,
	}

	return decodedOffer, nil
}

func (c *P2PConnectionState) DecodeAnswer(answer string) (*DecodedAnswer, error) {
	decoded, err := base64.URLEncoding.DecodeString(answer)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("gzip reader create failed: %w", err)
	}
	defer gzReader.Close()

	uncompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("gzip decompress failed: %w", err)
	}

	var encodedAnswer EncodedAnswer
	if err := json.Unmarshal(uncompressed, &encodedAnswer); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	if encodedAnswer.SDP == "" {
		return nil, errors.New("empty SDP in answer")
	}

	return &DecodedAnswer{
		SDP: webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  encodedAnswer.SDP,
		},
		Hash: encodedAnswer.Hash,
	}, nil
}

func (c *P2PConnectionState) CreateEncodedAnswer(opts *webrtc.AnswerOptions) (string, error) {
	if c.conn.RemoteDescription() == nil {
		return "", errors.New("remote description not set")
	}

	answer, err := c.conn.CreateAnswer(opts)
	if err != nil {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(c.Password()),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	data := EncodedAnswer{
		SDP:  answer.SDP,
		Hash: string(hashedPassword),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	if err = c.conn.SetLocalDescription(answer); err != nil {
		return "", err
	}

	c.answer = &answer

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

func (c *P2PConnectionState) SendMessage(msg []byte) error {
	if !c.isConnected.Load() {
		return errors.New("not connected")
	}
	return c.dc.Send(msg)
}

func (c *P2PConnectionState) WaitForConnection(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New("connection timeout")
		default:
			if c.isConnected.Load() {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *P2PConnectionState) ValidatePassword(hash string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(c.Password()),
	)
}

func (c *P2PConnectionState) SetCallbacks(
	onConnect func(),
	onDisconnect func(),
	onMessage func([]byte),
) {
	c.onConnect = onConnect
	c.onDisconnect = onDisconnect
	c.onMessage = onMessage
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

func (c *P2PConnectionState) AddRemoteICECandidates() error {
	for _, candidate := range c.ICECandidates() {
		if err := c.conn.AddICECandidate(candidate); err != nil {
			return err
		}
	}

	return nil
}

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

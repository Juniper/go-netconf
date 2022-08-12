package netconf

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/nemith/go-netconf/v2/transport"
)

type sessionConfig struct {
	capabilities []string
}

type SessionOption interface {
	apply(*sessionConfig)
}

type capabilityOpt []string

func (o capabilityOpt) apply(cfg *sessionConfig) {
	for _, cap := range o {
		cfg.capabilities = append(cfg.capabilities, cap)
	}
}

func WithCapability(capabilities ...string) SessionOption {
	return capabilityOpt(capabilities)
}

var DefaultCapabilities = []string{
	"urn:ietf:params:netconf:base:1.0",
	"urn:ietf:params:netconf:base:1.1",

	// XXX: these seems like server capabilities and i don't see why
	// a client would need to send them

	// "urn:ietf:params:netconf:capability:writable-running:1.0",
	// "urn:ietf:params:netconf:capability:candidate:1.0",
	// "urn:ietf:params:netconf:capability:confirmed-commit:1.0",
	// "urn:ietf:params:netconf:capability:rollback-on-error:1.0",
	// "urn:ietf:params:netconf:capability:startup:1.0",
	// "urn:ietf:params:netconf:capability:url:1.0?scheme=http,ftp,file,https,sftp",
	// "urn:ietf:params:netconf:capability:validate:1.0",
	// "urn:ietf:params:netconf:capability:xpath:1.0",
	// "urn:ietf:params:netconf:capability:notification:1.0",
	// "urn:ietf:params:netconf:capability:interleave:1.0",
	// "urn:ietf:params:netconf:capability:with-defaults:1.0",
}

type Session struct {
	tr        transport.Transport
	sessionID uint64

	clientCaps CapabilitySet
	serverCaps CapabilitySet

	mu      sync.Mutex
	seq     uint64
	reqs    map[uint64]chan RPCReplyMsg
	closing bool
}

func Open(transport transport.Transport, opts ...SessionOption) (*Session, error) {
	cfg := sessionConfig{
		capabilities: DefaultCapabilities,
	}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	s := &Session{
		tr: transport,
		// XXX: fix me.  We doing sets or slices.  Figure it out man
		clientCaps: NewCapabilitySet(cfg.capabilities...),
		reqs:       make(map[uint64]chan RPCReplyMsg),
	}

	// this needs a timeout of some sort.
	if err := s.hello(); err != nil {
		s.tr.Close()
		return nil, err
	}

	go s.recv()
	return s, nil
}

// hello exchanges hello messages and reports if there are any errors.
func (s *Session) hello() error {
	clientMsg := HelloMsg{
		Capabilities: s.clientCaps.All(),
	}
	if err := s.writeMsg(&clientMsg); err != nil {
		return fmt.Errorf("failed to write hello message: %w", err)
	}

	r, err := s.tr.MsgReader()
	if err != nil {
		return err
	}

	var serverMsg HelloMsg
	if err := xml.NewDecoder(r).Decode(&serverMsg); err != nil {
		return fmt.Errorf("failed to read server hello message: %w", err)
	}

	if serverMsg.SessionID == 0 {
		return fmt.Errorf("server did not return a session-id")
	}

	if len(serverMsg.Capabilities) == 0 {
		return fmt.Errorf("server did not return any capabilities")
	}
	s.serverCaps = NewCapabilitySet(serverMsg.Capabilities...)

	// upgrade the transport if we are on a larger version and the transport
	// supports it.
	const baseCap11 = baseCap + ":1.1"
	if s.serverCaps.Has(baseCap11) && s.clientCaps.Has(baseCap11) {
		if upgrader, ok := s.tr.(transport.Upgrader); ok {
			upgrader.Upgrade()
		}
	}

	return nil
}

func (s *Session) SessionID() uint64 {
	return s.sessionID
}

// startElement will walk though a xml.Decode until it finds a start element
// and returns it.
func startElement(d *xml.Decoder) (*xml.StartElement, error) {
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}

		if start, ok := tok.(xml.StartElement); ok {
			return &start, nil
		}
	}
}

// recv is the main receive loop.  It runs concurrently to be able to handle
// interleaved messages (like notifications).
func (s *Session) recv() {
	var (
		r    io.Reader
		dec  *xml.Decoder
		root *xml.StartElement
		err  error
	)
Loop:
	for {
		r, err = s.tr.MsgReader()
		if err != nil {
			break
		}
		dec = xml.NewDecoder(r)

		root, err = startElement(dec)
		if err != nil {
			break
		}

		// FIXME: This should look for a namspaces as well (strict node?)
		switch root.Name.Local {
		case "notification":
			var notif NotificationMsg
			if err := dec.DecodeElement(&notif, root); err != nil {
				log.Printf("failed to decode notification message: %v", err)
			}
			// DO something with this
		case "rpc-reply":
			var reply RPCReplyMsg
			if err := dec.DecodeElement(&reply, root); err != nil {
				log.Printf("failed to decode rpc-reply message: %v", err)
			}
			ok, ch := s.replyChan(reply.MessageID)
			if !ok {
				log.Printf("cannot find reply channel for message-id %d", reply.MessageID)
				continue Loop
			}
			ch <- reply
		default:
			log.Printf("improper xml message type %q", root.Name.Local)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		if s.closing {
			return
		}
	}
	// Close the connection
	log.Fatal(err)
}

func (s *Session) replyChan(msgID uint64) (bool, chan RPCReplyMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, ok := s.reqs[msgID]
	if !ok {
		return false, nil
	}
	delete(s.reqs, msgID)
	return true, ch
}

func (s *Session) writeMsg(v any) error {
	w, err := s.tr.MsgWriter()
	if err != nil {
		return err
	}

	if err := xml.NewEncoder(w).Encode(v); err != nil {
		return err
	}
	return w.Close()
}

func (s *Session) send(msg *RPCMsg) (chan RPCReplyMsg, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	msg.MessageID = s.seq

	if err := s.writeMsg(msg); err != nil {
		return nil, err
	}

	// cap of 1 makes sure we don't block on send
	ch := make(chan RPCReplyMsg, 1)
	s.reqs[msg.MessageID] = ch

	return ch, nil

}

func (s *Session) Do(ctx context.Context, msg *RPCMsg) (*RPCReplyMsg, error) {
	ch, err := s.send(msg)
	if err != nil {
		return nil, err
	}

	select {
	case reply := <-ch:
		return &reply, nil
	case <-ctx.Done():
		// remove any existing request
		s.mu.Lock()
		delete(s.reqs, msg.MessageID)
		s.mu.Unlock()

		return nil, ctx.Err()

		// XXX: stop channel on close?
	}
}

func (s *Session) Call(ctx context.Context, op any, resp any) error {
	msg := &RPCMsg{
		Operation: op,
	}

	reply, err := s.Do(ctx, msg)
	if err != nil {
		return err
	}

	if resp != nil {
		if err := xml.Unmarshal(reply.Data, resp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Close will gracefully close the sessions first by sending a `close-session`
// operation to the remote and then closing the underlying transport
func (s *Session) Close(ctx context.Context) error {
	s.mu.Lock()
	s.closing = true
	s.mu.Unlock()

	// Even is this call fails we want to close the underlying connection but
	// perhaps the error is useful so save it and return it if there isn't
	// any other encounterd.
	rpcErr := s.Call(ctx, &closeSession{}, nil)

	if err := s.tr.Close(); err != nil {
		return err
	}

	return rpcErr
}

func (s *Session) ClientCapabilities() CapabilitySet {
	// XXX: should we clone this? Do we care of someone is careless with the values
	return s.clientCaps
}

func (s *Session) ServerCapabilities() CapabilitySet {
	// XXX: should we clone this? Do we care of someone is careless with the values
	return s.serverCaps
}

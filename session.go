package netconf

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

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
	msgID     uint64
	sessionID uint64

	clientCaps CapabilitySet
	serverCaps CapabilitySet

	mu   sync.Mutex
	reqs map[uint64]chan RPCReplyMsg
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

	if err := s.hello(); err != nil {
		return nil, err
	}

	go s.recv()
	return s, nil
}

func (s *Session) writeMsg(v any) error {
	w := s.tr.MsgWriter()
	io.WriteString(w, xml.Header)
	if err := xml.NewEncoder(w).Encode(v); err != nil {
		return err
	}
	return w.Close()
}

func (s *Session) hello() error {
	clientMsg := HelloMsg{
		Capabilities: s.clientCaps.All(),
	}
	if err := s.writeMsg(&clientMsg); err != nil {
		return fmt.Errorf("failed to write hello message: %w", err)
	}

	var serverMsg HelloMsg
	if err := xml.NewDecoder(s.tr.MsgReader()).Decode(&serverMsg); err != nil {
		return fmt.Errorf("failed to read server hello message: %w", err)
	}

	if serverMsg.SessionID == 0 {
		return fmt.Errorf("server did not return a session-id")
		// XXX: close session
	}

	if len(serverMsg.Capabilities) == 0 {
		return fmt.Errorf("server did not return any capabilities")
	}
	s.serverCaps = NewCapabilitySet(serverMsg.Capabilities...)

	// upgrade the transport if we are on a larger version and the transport
	// supports it.
	const baseCap11 = baseCap + ":1.1"
	if s.serverCaps.Has(baseCap11) && s.clientCaps.Has(baseCap11) {
		log.Println("upgrading")
		if upgrader, ok := s.tr.(transport.Upgrader); ok {
			upgrader.Upgrade()
		}
	}

	return nil
}

func (s *Session) SessionID() uint64 {
	return s.sessionID
}

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

func (s *Session) recv() {
Loop:
	for {
		dec := xml.NewDecoder(s.tr.MsgReader())

		root, err := startElement(dec)
		if err != nil {
			if err != io.EOF {
				log.Printf("failed to read StartElement token: %v", err)
				// XXX panic here or return errors next call?
			}
			continue
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
			ch := s.replyChan(reply.MessageID)
			if ch == nil {
				log.Printf("cannot find reply channel for message-id %d", reply.MessageID)
				continue Loop
			}
			ch <- reply
		}
	}
}

func (s *Session) replyChan(msgID uint64) chan RPCReplyMsg {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, ok := s.reqs[msgID]
	if !ok {
		// FIXME what to do here?  Log Faiil What?
		return nil
	}
	delete(s.reqs, msgID)
	return ch
}

func (s *Session) nextMsgID() uint64 {
	return uint64(atomic.AddUint64(&s.msgID, 1))
}

func (s *Session) Call(ctx context.Context, op any) (*RPCReplyMsg, error) {
	rpc := RPCMsg{
		MessageID: s.nextMsgID(),
		Operation: op,
	}
	if err := s.writeMsg(&rpc); err != nil {
		return nil, err
	}

	// cap of 1 makes sure we don't block on send
	ch := make(chan RPCReplyMsg, 1)
	s.mu.Lock()
	s.reqs[rpc.MessageID] = ch
	s.mu.Unlock()

	select {
	case reply := <-ch:
		return &reply, nil
	case <-ctx.Done():
		// FIXME: need to poision/purge the reply from the map
		return nil, ctx.Err()
	}
}

func (s *Session) Close() error {
	// do the things
	return nil
}

func (s *Session) ClientCapabilities() CapabilitySet {
	// XXX: should we clone this? Do we care of someone is careless with the values
	return s.clientCaps
}

func (s *Session) ServerCapabilities() CapabilitySet {
	// XXX: should we clone this? Do we care of someone is careless with the values
	return s.serverCaps
}

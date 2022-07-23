package netconf

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/nemith/go-netconf/v2/transport"
)

var DefaultCapabilities = []string{
	"urn:ietf:params:netconf:base:1.0",
	"urn:ietf:params:netconf:base:1.1",
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

	clientCaps map[string]struct{}
	serverCaps map[string]struct{}

	mu   sync.Mutex
	reqs map[uint64]chan RPCReplyMsg
}

func Open(transport transport.Transport) (*Session, error) {
	sess := &Session{
		reqs:       make(map[uint64]chan RPCReplyMsg),
		serverCaps: make(map[string]struct{}),
		tr:         transport,
	}

	if err := sess.hello(); err != nil {
		return nil, err
	}

	go sess.recv()

	return sess, nil
}

func (s *Session) ServerCapabilities() []string {
	out := make([]string, 0, len(s.serverCaps))
	for k := range s.serverCaps {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
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
		// TODO: allow for custom client capability lists
		Capabilities: DefaultCapabilities,
	}
	if err := s.writeMsg(&clientMsg); err != nil {
		return fmt.Errorf("failed to write hello message: %w", err)
	}

	var serverMsg HelloMsg
	if err := xml.NewDecoder(s.tr.MsgReader()).Decode(&serverMsg); err != nil {
		return fmt.Errorf("failed to read server hello message: %w", err)
	}
	for _, cap := range serverMsg.Capabilities {
		s.serverCaps[cap] = struct{}{}
	}

	// FIXME: check base capabilities and versions here
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

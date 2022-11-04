package netconf

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
)

type OK bool

func (b *OK) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	v := &struct{}{}
	if err := d.DecodeElement(v, &start); err != nil {
		return err
	}
	*b = v != nil
	return nil
}

type OKResp struct {
	OK OK `xml:"ok"`
}

type Datastore string

func (s Datastore) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if s == "" {
		return fmt.Errorf("datastores cannot be empty")
	}

	// XXX: it would be nice to actually just block names with crap in them
	// instead of escaping them, but we need to find a list of what is allowed.
	escaped, err := escapeXML(string(s))
	if err != nil {
		return fmt.Errorf("invalid string element: %w", err)
	}

	v := struct {
		Elem string `xml:",innerxml"`
	}{Elem: "<" + escaped + "/>"}
	return e.EncodeElement(&v, start)
}

func escapeXML(input string) (string, error) {
	buf := &strings.Builder{}
	if err := xml.EscapeText(buf, []byte(input)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type URL string

func (u URL) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	v := struct {
		URL string `xml:"url"`
	}{string(u)}
	return e.EncodeElement(&v, start)
}

/*type RawConfig []byte

func (c RawConfig) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var v struct {
		Config struct {
			Inner []byte `xml:",innerxml"`
		} `xml:"config,"`
	}
	v.Config.Inner = []byte(c)

	return e.EncodeElement(&v, start)
}*/

// XXX: these should be typed?
const (
	// Running configuration datastore. Required by RFC6241
	Running Datastore = "running"

	// Candidate configuration configuration datastore.  Supported with the
	// `:canidate` capability defined in RFC6241 section 8.3
	Candidate Datastore = "candidate"

	// Startup configuration configuration datastore.  Supported with the
	// `:startup` capability defined in RFC6241 section 8.7
	Startup Datastore = "startup" //
)

type getConfigReq struct {
	XMLName xml.Name  `xml:"get-config"`
	Source  Datastore `xml:"source"`
	// Filter
}

type getConfigResp struct {
	XMLName xml.Name `xml:"data"`
	Config  []byte   `xml:",innerxml"`
}

// GetConfig implments the <get-config> rpc operation defined in [RFC6241 7.1].
// `source` is the datastore to query.
//
// [RFC6241 7.1]: https://www.rfc-editor.org/rfc/rfc6241.html#section-7.1
func (s *Session) GetConfig(ctx context.Context, source Datastore) ([]byte, error) {
	req := getConfigReq{
		Source: source,
	}
	var resp getConfigResp

	if err := s.Call(ctx, &req, &resp); err != nil {
		return nil, err
	}

	return resp.Config, nil
}

// MergeStrategy defines the strategies for merging configuration in a
// `<edit-config> operation`.
//
// *Note*: in RFC6241 7.2 this is called the `operation` attribute and
// `default-operation` parameter.  Since the `operation` term is already
// overloaded this was changed to `MergeStrategy` for a cleaner API.
type MergeStrategy string

const (
	// MergeConfig configuration elements are merged together at the level at
	// which this specified.  Can be used for config elements as well as default
	// defined with [WithDefaultMergeStrategy] option.
	MergeConfig MergeStrategy = "merge"

	// ReplaceConfig defines that the incoming config change should replace the
	// existing config at the level which it is specified.  This can be
	// specified on indivual config elements or set as the default strategy set
	// with [WithDefaultMergeStrategy] option.
	ReplaceConfig MergeStrategy = "replace"

	// NoMergeStategy is only used as a default strategy defined in
	// [WithDefaultMergeStragegy].  Elements must specific one of the other
	// stragegies with the `operation` Attribute on elements in the `<config>`
	// subtree.  Elements without the `operation` attribute are ignored.
	NoMergeStragegy MergeStrategy = "none"

	// CreateConfig allows a subtree element to be created only if it doesn't
	// already exist.
	// This strategy is only used as the `operation` attribute of
	// a `<config>` element and cannot be used as the defaul strategy.
	CreateConfig MergeStrategy = "create"

	// DeleteConfig will completely delete subtree from the config only if it
	// already exists.  This strategy is only used as the `operation` attribute
	// of a `<config>` element and cannot be used as the defaul strategy.
	DeleteConfig MergeStrategy = "delete"

	// Removeonfig will remove subtree from the config.  If the subtree doesn't
	// exist in the datastore then it is siliently skipped.  This strategy is
	// only used as the `operation` attribute of a `<config>` element and cannot
	// be used as the defaul strategy.
	RemoveConfig MergeStrategy = "remove"
)

// TestStrategy defines the beahvior for testing configuration before applying it in a `<edit-config>` operation.
//
// *Note*: in RFC6241 7.2 this is called the `test-option` parameter. Since the `option` term is already
// overloaded this was changed to `TestStrategy` for a cleaner API.
type TestStrategy string

const (
	// TestThenSet will validate the configuration and only if is is valid then
	// apply the configuration to the datastore.
	TestThenSet TestStrategy = "test-then-set"

	// SetOnly will not do any testing before applying it.
	SetOnly TestStrategy = "set"

	// Test only will validation the incoming configiration and return the
	// results without modifying the underlying store.
	TestOnly TestStrategy = "test-only"
)

// ErrorStrategy defines the behavior when an error is encountered during a `<edit-config>` operation.
//
// *Note*: in RFC6241 7.2 this is called the `error-option` parameter. Since the `option` term is already
// overloaded this was changed to `ErrorStrategy` for a cleaner API.
type ErrorStrategy string

const (
	// StopOnError will about the `<edit-config>` operation on the first error.
	StopOnError ErrorStrategy = "stop-on-error"

	// ContinueOnError will continue to parse the configiration data even if an
	// error is encountered.  Errors are still recorded and reported in the
	// reply.
	ContinueOnError ErrorStrategy = "continue-on-error"

	// RollbackOnError will restore the configuration back to before the
	// `<edit-config>` operation took place.  This requires the device to
	// support the `:rollback-on-error` capabilitiy.
	RollbackOnError ErrorStrategy = "rollback-on-error"
)

type (
	defaultMergeStrategy MergeStrategy
	testStrategy         TestStrategy
	errorStrategy        ErrorStrategy
)

func (o defaultMergeStrategy) apply(req *editConfigReq) { req.DefaultMergeStrategy = MergeStrategy(o) }
func (o testStrategy) apply(req *editConfigReq)         { req.TestStrategy = TestStrategy(o) }
func (o errorStrategy) apply(req *editConfigReq)        { req.ErrorStrategy = ErrorStrategy(o) }

// WithDefaultMergeStrategy sets the default config merging strategy for the
// <edit-config> operation.  Only [Merge], [Replace], and [None] are suppored
// (the rest of the strategies are for defining as attributed in individual
// elements inside the `<config>` subtree).
func WithDefaultMergeStrategy(op MergeStrategy) EditConfigOption { return defaultMergeStrategy(op) }

// WithTestStrategy sets the `test-option` in the `<edit-config>“ operation.
// This defines what testing should be done the supplied configuration.  See the
// documenation on [TestStrategy] for details on each strategy.
func WithTestStrategy(op TestStrategy) EditConfigOption { return testStrategy(op) }

// WithErrorStrategy sets the `error-option` in the `<edit-config>` operation.
// This defines the behavior when errors are encountered applying the supplied
// config.  See [ErrorStrategy] for the available options.
func WithErrorStrategy(opt ErrorStrategy) EditConfigOption { return errorStrategy(opt) }

type editConfigReq struct {
	XMLName              xml.Name      `xml:"edit-config"`
	Target               Datastore     `xml:"target"`
	DefaultMergeStrategy MergeStrategy `xml:"default-operation,omitempty"`
	TestStrategy         TestStrategy  `xml:"test-option,omitempty"`
	ErrorStrategy        ErrorStrategy `xml:"error-option,omitempty"`
	// either of these two values
	Config interface{} `xml:"config,omitempty"`
	URL    string      `xml:"url,omitempty"`
}

// EditOption is a optional arguments to [Session.EditConfig] method
type EditConfigOption interface {
	apply(*editConfigReq)
}

// EditConfig issues the `<edit-config>` operation defined in [RFC6241 7.2] for
// updating an existing target config datastore.
//
// [RFC6241 7.2]: https://www.rfc-editor.org/rfc/rfc6241.html#section-7.2
func (s *Session) EditConfig(ctx context.Context, target Datastore, config interface{}, opts ...EditConfigOption) error {
	req := editConfigReq{
		Target: target,
	}

	// XXX: Should we use reflect here?
	switch v := config.(type) {
	case string:
		req.Config = struct {
			Inner []byte `xml:",innerxml"`
		}{Inner: []byte(v)}
	case []byte:
		req.Config = struct {
			Inner []byte `xml:",innerxml"`
		}{Inner: v}
	case URL:
		req.URL = string(v)
	default:
		req.Config = config
	}

	for _, opt := range opts {
		opt.apply(&req)
	}

	var resp OKResp
	return s.Call(ctx, &req, &resp)
}

type copyConfigReq struct {
	XMLName xml.Name    `xml:"copy-config"`
	Source  interface{} `xml:"source"`
	Target  interface{} `xml:"target"`
}

// CopyConfig issues the `<copy-config>` operation as defined in [RFC6241 7.3]
// for copying an entire config to/from a source and target datastore.
//
// A `<config>` element defining a full config can be used as the source.
//
// If a device supports the `:url` capability than a [URL] object can be used
// for the source or target datastore.
//
// [RFC6241 7.3] https://www.rfc-editor.org/rfc/rfc6241.html#section-7.3
func (s *Session) CopyConfig(ctx context.Context, source, target interface{}) error {
	req := copyConfigReq{
		Source: source,
		Target: target,
	}

	var resp OKResp
	return s.Call(ctx, &req, &resp)
}

type deleteConfigReq struct {
	XMLName xml.Name  `xml:"delete-config"`
	Target  Datastore `xml:"target"`
}

func (s *Session) DeleteConfig(ctx context.Context, target Datastore) error {
	req := deleteConfigReq{
		Target: target,
	}

	var resp OKResp
	return s.Call(ctx, &req, &resp)
}

type lockReq struct {
	XMLName xml.Name
	Target  Datastore `xml:"target"`
}

func (s *Session) Lock(ctx context.Context, target Datastore) error {
	req := lockReq{
		XMLName: xml.Name{Local: "lock"},
		Target:  target,
	}
	var resp OKResp

	return s.Call(ctx, &req, &resp)
}

func (s *Session) Unlock(ctx context.Context, target Datastore) error {
	req := lockReq{
		XMLName: xml.Name{Local: "unlock"},
		Target:  target,
	}
	var resp OKResp

	return s.Call(ctx, &req, &resp)
}

func (s *Session) Get(ctx context.Context /* filter */) error {
	panic("unimplemented")
}

type killSessionReq struct {
	XMLName   xml.Name `xml:"kill-session"`
	SessionID uint32   `xml:"session-id"`
}

func (s *Session) KillSession(ctx context.Context, sessionID uint32) error {
	req := killSessionReq{
		SessionID: sessionID,
	}
	var resp OKResp

	return s.Call(ctx, &req, &resp)
}

type validateReq struct {
	XMLName xml.Name    `xml:"validate"`
	Source  interface{} `xml:"source"`
}

func (s *Session) Validate(ctx context.Context, source interface{}) error {
	req := validateReq{
		Source: source,
	}

	var resp OKResp
	return s.Call(ctx, &req, &resp)
}

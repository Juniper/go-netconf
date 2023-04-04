package netconf

const (
	baseCap      = "urn:ietf:params:netconf:base"
	stdCapPrefix = "urn:ietf:params:netconf:capability"
)

// DefaultCapabilities are the capabilities sent by the client during the hello
// exchange by the server.
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

// ExpandCapability will automatically add the standard capability prefix of
// `urn:ietf:params:netconf:capability` if not already present.
func ExpandCapability(s string) string {
	if s == "" {
		return ""
	}

	if s[0] != ':' {
		return s
	}

	return stdCapPrefix + s
}

// XXX: may want to expose this type publicly in the future when the api has
// stabilized?
type capabilitySet struct {
	caps map[string]struct{}
}

func newCapabilitySet(capabilities ...string) capabilitySet {
	cs := capabilitySet{
		caps: make(map[string]struct{}),
	}
	cs.Add(capabilities...)
	return cs
}

func (cs *capabilitySet) Add(capabilities ...string) {
	for _, cap := range capabilities {
		cap = ExpandCapability(cap)
		cs.caps[cap] = struct{}{}
	}
}

func (cs capabilitySet) Has(s string) bool {
	// XXX: need to figure out how to handle versions (i.e always map to 1.0 or
	// map to latest/any?)
	s = ExpandCapability(s)
	_, ok := cs.caps[s]
	return ok
}

func (cs capabilitySet) All() []string {
	out := make([]string, 0, len(cs.caps))
	for cap := range cs.caps {
		out = append(out, cap)
	}
	return out
}

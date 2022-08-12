package netconf

const (
	baseCap      = "urn:ietf:params:netconf:base"
	stdCapPrefix = "urn:ietf:params:netconf:capability"
)

func ExpandCapability(s string) string {
	if s == "" {
		return ""
	}

	if s[0] != ':' {
		return s
	}

	return stdCapPrefix + s
}

type CapabilitySet struct {
	caps map[string]struct{}
}

func NewCapabilitySet(capabilities ...string) CapabilitySet {
	cs := CapabilitySet{
		caps: make(map[string]struct{}),
	}
	cs.Add(capabilities...)
	return cs
}

func (cs *CapabilitySet) Add(capabilities ...string) {
	for _, cap := range capabilities {
		cap = ExpandCapability(cap)
		cs.caps[cap] = struct{}{}
	}
}

func (cs CapabilitySet) Has(s string) bool {
	// XXX: need to figure out how to handle versions (i.e always map to 1.0 or
	// map to latest/any?)
	s = ExpandCapability(s)
	_, ok := cs.caps[s]
	return ok
}

func (cs CapabilitySet) All() []string {
	out := make([]string, 0, len(cs.caps))
	for cap := range cs.caps {
		out = append(out, cap)
	}
	return out
}

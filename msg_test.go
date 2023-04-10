package netconf

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ()

var helloMsgTestTable = []struct {
	name string
	raw  []byte
	msg  HelloMsg
}{
	{
		name: "basic",
		raw:  []byte(`<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"><capabilities><capability>urn:ietf:params:netconf:base:1.0</capability><capability>urn:ietf:params:netconf:base:1.1</capability></capabilities></hello>`),
		msg: HelloMsg{
			XMLName: xml.Name{
				Local: "hello",
				Space: "urn:ietf:params:xml:ns:netconf:base:1.0",
			},
			Capabilities: []string{
				"urn:ietf:params:netconf:base:1.0",
				"urn:ietf:params:netconf:base:1.1",
			},
		},
	},
	{
		name: "junos",
		raw: []byte(`<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
  <capabilities>
      <capability>urn:ietf:params:netconf:base:1.0</capability>
	  <capability>urn:ietf:params:netconf:capability:candidate:1.0</capability>
	  <capability>urn:ietf:params:netconf:capability:confirmed-commit:1.0</capability>
	  <capability>urn:ietf:params:netconf:capability:validate:1.0</capability>
	  <capability>urn:ietf:params:netconf:capability:url:1.0?scheme=http,ftp,file</capability>
	  <capability>urn:ietf:params:xml:ns:netconf:base:1.0</capability>
	  <capability>urn:ietf:params:xml:ns:netconf:capability:candidate:1.0</capability>
	  <capability>urn:ietf:params:xml:ns:netconf:capability:confirmed-commit:1.0</capability>
	  <capability>urn:ietf:params:xml:ns:netconf:capability:validate:1.0</capability>
	  <capability>urn:ietf:params:xml:ns:netconf:capability:url:1.0?scheme=http,ftp,file</capability>
	  <capability>urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring</capability>
	  <capability>http://xml.juniper.net/netconf/jdm/1.0</capability>
  </capabilities>
  <session-id>410</session-id>
</hello>`),
		msg: HelloMsg{
			XMLName: xml.Name{
				Local: "hello",
				Space: "urn:ietf:params:xml:ns:netconf:base:1.0",
			},
			Capabilities: []string{
				"urn:ietf:params:netconf:base:1.0",
				"urn:ietf:params:netconf:capability:candidate:1.0",
				"urn:ietf:params:netconf:capability:confirmed-commit:1.0",
				"urn:ietf:params:netconf:capability:validate:1.0",
				"urn:ietf:params:netconf:capability:url:1.0?scheme=http,ftp,file",
				"urn:ietf:params:xml:ns:netconf:base:1.0",
				"urn:ietf:params:xml:ns:netconf:capability:candidate:1.0",
				"urn:ietf:params:xml:ns:netconf:capability:confirmed-commit:1.0",
				"urn:ietf:params:xml:ns:netconf:capability:validate:1.0",
				"urn:ietf:params:xml:ns:netconf:capability:url:1.0?scheme=http,ftp,file",
				"urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring",
				"http://xml.juniper.net/netconf/jdm/1.0",
			},
			SessionID: 410,
		},
	},
}

func TestUnmarshalHelloMsg(t *testing.T) {
	for _, tc := range helloMsgTestTable {
		t.Run(tc.name, func(t *testing.T) {
			var got HelloMsg
			err := xml.Unmarshal(tc.raw, &got)
			assert.NoError(t, err)
			assert.Equal(t, got, tc.msg)
		})
	}
}
func TestMarshalHelloMsg(t *testing.T) {
	for _, tc := range helloMsgTestTable {
		t.Run(tc.name, func(t *testing.T) {
			out, err := xml.Marshal(tc.msg)
			t.Logf("out: %s", out)
			assert.NoError(t, err)
		})
	}
}

func TestMarshalRPCMsg(t *testing.T) {
	tt := []struct {
		name      string
		operation interface{}
		err       bool
		want      []byte
	}{
		{
			name:      "nil",
			operation: nil,
			err:       true,
		},
		{
			name:      "string",
			operation: "<foo><bar/></foo>",
			want:      []byte(`<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><foo><bar/></foo></rpc>`),
		},
		{
			name:      "byteslice",
			operation: []byte("<baz><qux/></baz>"),
			want:      []byte(`<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><baz><qux/></baz></rpc>`),
		},
		{
			name:      "validate",
			operation: validateReq{Source: Running},
			want:      []byte(`<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><validate><source><running/></source></validate></rpc>`),
		},
		{
			name: "namedStruct",
			operation: struct {
				XMLName xml.Name `xml:"http://xml.juniper.net/junos/22.4R0/junos command"`
				Command string   `xml:",innerxml"`
			}{
				Command: "show bgp neighbors",
			},
			want: []byte(`<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><command xmlns="http://xml.juniper.net/junos/22.4R0/junos">show bgp neighbors</command></rpc>`),
		},
		/*
			{
				name: "unnamedStruct",
				operation: struct {
					Command string `xml:"command"`
				}{
					Command: "show version",
				},
			},*/
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := xml.Marshal(&RPCMsg{
				MessageID: 1,
				Operation: tc.operation,
			})
			t.Logf("out: %s", out)

			if tc.err {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, out, tc.want)
		})
	}
}

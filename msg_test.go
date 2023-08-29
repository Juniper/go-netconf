package netconf

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rawXMLTests = []struct {
	name        string
	element     RawXML
	xml         []byte
	noUnmarshal bool
}{
	{
		name:    "empty",
		element: RawXML(""),
		xml:     []byte("<RawXML></RawXML>"),
	},
	{
		name:    "nil",
		element: nil,
		xml:     []byte("<RawXML></RawXML>"),
		// we will never unmarshal back into a nil (see "empty" testcase)
		noUnmarshal: true,
	},
	{
		name:    "textElement",
		element: RawXML("A man a plan a canal panama"),
		xml:     []byte("<RawXML>A man a plan a canal panama</RawXML>"),
	},
	{
		name:    "xml",
		element: RawXML("<foo><bar>hamburger</bar></foo>"),
		xml:     []byte("<RawXML><foo><bar>hamburger</bar></foo></RawXML>"),
	},
}

func TestRawXMLUnmarshal(t *testing.T) {
	for _, tc := range rawXMLTests {
		if tc.noUnmarshal {
			continue
		}

		t.Run(tc.name, func(t *testing.T) {
			var got RawXML
			err := xml.Unmarshal(tc.xml, &got)
			assert.NoError(t, err)
			assert.Equal(t, tc.element, got)
		})
	}
}

func TestRawXMLMarshal(t *testing.T) {
	for _, tc := range rawXMLTests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := xml.Marshal(&tc.element)
			assert.NoError(t, err)
			assert.Equal(t, tc.xml, got)
		})
	}
}

var helloMsgTestTable = []struct {
	name string
	raw  []byte
	msg  helloMsg
}{
	{
		name: "basic",
		raw:  []byte(`<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"><capabilities><capability>urn:ietf:params:netconf:base:1.0</capability><capability>urn:ietf:params:netconf:base:1.1</capability></capabilities></hello>`),
		msg: helloMsg{
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
		msg: helloMsg{
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
			var got helloMsg
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
		operation any
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
			operation: ValidateReq{Source: Running},
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
			out, err := xml.Marshal(&request{
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

var replyJunosGetConfigError = []byte(`
<rpc-reply xmlns:junos="http://xml.juniper.net/junos/20.3R0/junos" xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1">
<rpc-error>
<error-type>protocol</error-type>
<error-tag>operation-failed</error-tag>
<error-severity>error</error-severity>
<error-message>syntax error, expecting &lt;candidate/&gt; or &lt;running/&gt;</error-message>
<error-info>
<bad-element>non-exist</bad-element>
</error-info>
</rpc-error>
</rpc-reply>
`)

func TestUnmarshalRPCReply(t *testing.T) {
	tt := []struct {
		name  string
		reply []byte
		want  Reply
	}{
		{
			name:  "error",
			reply: replyJunosGetConfigError,
			want: Reply{
				XMLName: xml.Name{
					Space: "urn:ietf:params:xml:ns:netconf:base:1.0",
					Local: "rpc-reply",
				},
				MessageID: 1,
				Errors: []RPCError{
					{
						Type:     ErrTypeProtocol,
						Tag:      ErrOperationFailed,
						Severity: SevError,
						Message:  "syntax error, expecting <candidate/> or <running/>",
						Info: []byte(`
<bad-element>non-exist</bad-element>
`),
					},
				},
				Body: []byte(`
<rpc-error>
<error-type>protocol</error-type>
<error-tag>operation-failed</error-tag>
<error-severity>error</error-severity>
<error-message>syntax error, expecting &lt;candidate/&gt; or &lt;running/&gt;</error-message>
<error-info>
<bad-element>non-exist</bad-element>
</error-info>
</rpc-error>
`),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var got Reply
			err := xml.Unmarshal(tc.reply, &got)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

}

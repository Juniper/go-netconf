package netconf

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

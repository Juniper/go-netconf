package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type testServer struct {
	addr net.Addr
}

const hostkey = `
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEApUSW/7sajldrqHjhNcg6PUhUu8ztmWWfE1myIz9DpvnCgfBCsgQK
CF4uayYmN+FrxrEWovszrcDYxStvAFiDo6YAGI3CdePWOWJvzoAP+qtN018t3Fl3NaNBuB
J1u+aQpoAa7+37hq7i/Hu9Gu7QFSvtA+QVjwt/o9L9gyZMwDCmhxoo+XjdV/SGkxaA7NcR
y6YEPEbDFHdXPF9SPi3xY9+2UxPFSp7LmEREiZCq+KyGgUuNDyGei7Gb+jZUiXgctw30l3
f+mvQt6FKOmvHtevgIF2Eu7AVDIh+Hrj9fd3qpnsKsI+laBza+++2g34oJGPI0zXw9gFbV
yNeliet2uQAAA9C5b2AkuW9gJAAAAAdzc2gtcnNhAAABAQClRJb/uxqOV2uoeOE1yDo9SF
S7zO2ZZZ8TWbIjP0Om+cKB8EKyBAoIXi5rJiY34WvGsRai+zOtwNjFK28AWIOjpgAYjcJ1
49Y5Ym/OgA/6q03TXy3cWXc1o0G4EnW75pCmgBrv7fuGruL8e70a7tAVK+0D5BWPC3+j0v
2DJkzAMKaHGij5eN1X9IaTFoDs1xHLpgQ8RsMUd1c8X1I+LfFj37ZTE8VKnsuYRESJkKr4
rIaBS40PIZ6LsZv6NlSJeBy3DfSXd/6a9C3oUo6a8e16+AgXYS7sBUMiH4euP193eqmewq
wj6VoHNr777aDfigkY8jTNfD2AVtXI16WJ63a5AAAAAwEAAQAAAQAvF1Y3VCcC/CHvBVKW
spD1uVB7mq7xEKW9K8e4h2RNhclIoR8//iqlq8BqQ5qMPa0qFneuxQk6r0KVHAUrAg2wab
KJTItmcB8whr35B0CGWp14Zxx4Nv3iyLwHKStm+RGqf8ItL5CGFfsTmmaN8BJWlgeZHjqO
YeZi1dHqttUTxdM3wrBLF6teZ5q9xmLrzUK7+osq/Es8vghDzorxxtlxpwo/38kZchL67N
TI+GMc5tM0ExqwhPwNhDfEUF9WEIg8rB1YQA+/ehaaj2POnPcvOdGSHDSWdr443FOddbeP
VKJAqv4omh/T/LFzm8sbVdAomm2hAa2CWJkXVk6n3pepAAAAgB1Xq4PBJfUaOOqvzBRMdG
gmjnR/Qw4Pe3HcYC2G/BXvoQ8o8ioYmV7G9vsOY88gRqZisQdjfuaFsczJ+wtCTgePoZw3
foeATvKmm5Ds8tRr33CjL0wt/uDQRuLW8IiOmtN4+bgUQjJyD7Gu8THzg4NQ6dJunZsESr
tN7yojg7YWAAAAgQDj33QvfjabMKpbpRJ+Lk/KKqfHUj0s3OZIx8EQu/YzT+WZW6VwYAFe
pHuvBMN+SfiJW0niMdyDnrrROV/hIPtEJoS/sBBMGNJb/MLP2Cm3nySocvoT7J6OS1B7hJ
2rAjoAijpFi9HbyY00LbfNNXT9HJVrFpJ3uXwsn8OWPXNxRwAAAIEAuariiisP8L+/K9DM
3S3lpP816GQ+91gYGk1fS1jm5ejCieKJw3m/R96loV2wBz1NzSCn5B4sHqXIoVhYt7Ms8k
QXHvb1h5QQyt1p/F5eFC/f+ZEThsSSX6FIHjTazV3OxcvUxoHTG3P4RDNWY6yzo2iZke1R
0Oh3hpZmwH9d1/8AAAAYYnJhbmRvbmJlbm5ldHRAc3Ryb25nc2FkAQID
-----END OPENSSH PRIVATE KEY-----
`

func newTestServer(t *testing.T, handlerFn func(*testing.T, ssh.Channel, <-chan *ssh.Request)) (*testServer, error) {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	key, err := ssh.ParsePrivateKey([]byte(hostkey))
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	config.AddHostKey(key)

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}

	go func() {
		nconn, err := ln.Accept()
		if err != nil {
			t.Logf("failed to accept new conn: %v", err)
			return
		}

		_, chans, reqs, err := ssh.NewServerConn(nconn, config)
		if err != nil {
			t.Logf("failed to create ssh conn: %v", err)
			return
		}
		go ssh.DiscardRequests(reqs)

		for newChannel := range chans {
			if newChannel.ChannelType() != "session" {
				_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				continue
			}

			ch, reqs, err := newChannel.Accept()
			if err != nil {
				t.Logf("failed to accept new channel: %v", err)
				return
			}

			handlerFn(t, ch, reqs)
		}
	}()

	return &testServer{
		addr: ln.Addr(),
	}, nil
}

func TestTransport(t *testing.T) {
	var (
		srvIn bytes.Buffer
	)
	srvDone := make(chan struct{})
	server, err := newTestServer(t, func(t *testing.T, ch ssh.Channel, reqs <-chan *ssh.Request) {
		go func() {
			for req := range reqs {
				if req.Type != "subsystem" || !bytes.Equal(req.Payload[4:], []byte("netconf")) {
					panic(fmt.Sprintf("unknown ssh request: %q: %q", req.Type, req.Payload))
				}
				_ = req.Reply(true, nil)
			}
		}()
		_, _ = io.WriteString(ch, "muffins]]>]]>")
		_, _ = io.Copy(&srvIn, ch)
		close(srvDone)
	})
	require.NoError(t, err)

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	tr, err := Dial(context.Background(), "tcp", server.addr.String(), config)
	require.NoError(t, err)

	// test read
	r, err := tr.MsgReader()
	assert.NoError(t, err)

	_, err = io.ReadAll(r)
	assert.NoError(t, err)

	// test write
	w, err := tr.MsgWriter()
	assert.NoError(t, err)

	out := "a man a plan a canal panama"
	_, _ = io.WriteString(w, out)

	err = w.Close()
	assert.NoError(t, err)

	err = tr.Close()
	assert.NoError(t, err)

	// wait for the server to close
	<-srvDone

	want := out + "\n]]>]]>"
	assert.Equal(t, want, srvIn.String())
}

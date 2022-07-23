# Go `netconf` client library

[![GoDoc](https://godoc.org/github.com/nemith/go-netconf/v2?status.svg)](https://godoc.org/github.com/Juniper/go-netconf/netconf)
[![Report Card](https://goreportcard.com/badge/github.com/nemith/go-netconf/v2)](https://goreportcard.com/report/github.com/nemith/go-netconf/v2)

This library is used to create client applications for connecting to network devices via NETCONF.

## Support

| RFC                                                                               | Support                   |
| --------------------------------------------------------------------------------- | ------------------------- |
| [RFC6241 Network Configuration Protocol (NETCONF)][RFC6241]                       | :construction: inprogress |
| [RFC6242 Using the NETCONF Protocol over Secure Shell (SSH)][RFC6242]             | :heavy_check_mark:        |
| [RFC7589 Using the NETCONF Protocol over Transport Layer Security (TLS)][RFC7589] | planned                   |
| [RFC5277 NETCONF Event Notifications][RFC5277]                                    | planned                   |
| [RFC5717 Partial Lock Remote Procedure Call (RPC) for NETCONF][RFC5717]           | planned                   |
| [RFC6243 With-defaults Capability for NETCONF][RFC6243]                           | maybe                     |
| [RFC4743 Using NETCONF over the Simple Object Access Protocol (SOAP)][RFC4743]    | not planned               |
| [RFC4744 Using the NETCONF Protocol over the BEEP][RFC4744]                       | not planned               |

There are other RFC around YANG integration that will be looked at later.

[RFC4743]: [https://www.rfc-editor.org/rfc/rfc4743.html]
[RFC4744]: [https://www.rfc-editor.org/rfc/rfc4744.html]
[RFC5277]: [https://www.rfc-editor.org/rfc/rfc5277.html]
[RFC5717]: [https://www.rfc-editor.org/rfc/rfc5717.html]
[RFC6241]: [https://www.rfc-editor.org/rfc/rfc6241.html]
[RFC6242]: [https://www.rfc-editor.org/rfc/rfc6242.html]
[RFC6243]: [https://www.rfc-editor.org/rfc/rfc6243.html]
[RFC7589]: [https://www.rfc-editor.org/rfc/rfc7589.html]

See [TODO.md](TODO.md) for a list of what is left to implement these features.

## Differences from v1
* **Much cleaner and idomatic API**
* **Transports are implemented in their own packages.**  This means if you are not using SSH or TLS you don't need to bring in the underlying depdendencies.
* **Stream based transports.**  Should mean less memory usuage and less allocation bringing overall performance


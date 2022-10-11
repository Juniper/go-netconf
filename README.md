# Go `netconf` client library

WARNING: This is currently pre-alpha quality.  The API isn't solid yet and a lot of testing still needs to be done as well as additional feaures.  

[![GoDoc](https://godoc.org/github.com/nemith/go-netconf/v2?status.svg)](https://godoc.org/github.com/nemith/go-netconf/v2)
[![Report Card](https://goreportcard.com/badge/github.com/nemith/go-netconf/v2)](https://goreportcard.com/report/github.com/nemith/go-netconf/v2)

This library is used to create client applications for connecting to network devices via NETCONF.

## Support

| RFC                                                                               | Support                   |
| --------------------------------------------------------------------------------- | ------------------------- |
| [RFC6241 Network Configuration Protocol (NETCONF)][RFC6241]                       | :construction: inprogress |
| [RFC6242 Using the NETCONF Protocol over Secure Shell (SSH)][RFC6242]             | :heavy_check_mark:        |
| [RFC7589 Using the NETCONF Protocol over Transport Layer Security (TLS)][RFC7589] | :heavy_check_mark:        |
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

## Differences from [`github.com/juniper/go-netconf/netconf`](https://pkg.go.dev/github.com/Juniper/go-netconf)
* **Much cleaner, idomatic API, less dumb** I, @nemith, was the original creator of the netconf package and it was my very first Go project and it shows.  There are number of questionable API design, code, and a lot of odd un tested bugs.  Really this rewrite was created to fix this.
* **No impled vendor ownership** Moving the project out of the `Juniper` organization allowes better control over the project, less implied support (or lack there of), and hopefully more contributions.
* **Transports are implemented in their own packages.**  This means if you are not using SSH or TLS you don't need to bring in the underlying depdendencies into your binary.
* **Stream based transports.**  Should mean less memory usage and much less allocation bringing overall performance higher. 

## Differences from [`github.com/scrapli/scrapligo/driver/netconf`](https://pkg.go.dev/github.com/scrapli/scrapligo/driver/netconf)
Scrapligo driver is quite good and way better than the original juniper project.  However this new package concentrates more on RFC correctness and implementing some of the more advanced RFC features like call-home and event notifications.  If there is a desire there could be some callaboration with scrapligo in the future.
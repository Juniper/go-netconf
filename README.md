# Go `netconf` client library

WARNING: This is currently alpha quality.  The API isn't solid yet and a lot of testing still needs to be done as well as additional feaures.  Working for a solid beta API is incoming.

[![Go Reference](https://pkg.go.dev/badge/github.com/nemith/netconf.svg)](https://pkg.go.dev/github.com/nemith/netconf)
[![Report Card](https://goreportcard.com/badge/github.com/nemith/netconf)](https://goreportcard.com/report/github.com/nemith/netconf)
[![stability-unstable](https://img.shields.io/badge/stability-unstable-yellow.svg)](https://github.com/emersion/stability-badges#unstable)
[![Validate](https://github.com/nemith/netconf/actions/workflows/validate.yaml/badge.svg?branch=main&event=push)](https://github.com/nemith/netconf/actions/workflows/validate.yaml)
[![coverage](https://raw.githubusercontent.com/nemith/netconf/coverage/badge.svg)](http://htmlpreview.github.io/?https://github.com/nemith/netconf/blob/coverage/coverage.html)

This library is used to create client applications for connecting to network devices via NETCONF.

Like Go itself, only the latest two Go versions are tested and supported (Go 1.19 or Go 1.20).


## RFC Support

| RFC                                                                               | Support                      |
| --------------------------------------------------------------------------------- | ---------------------------- |
| [RFC6241 Network Configuration Protocol (NETCONF)][RFC6241]                       | :construction: inprogress    |
| [RFC6242 Using the NETCONF Protocol over Secure Shell (SSH)][RFC6242]             | :white_check_mark: supported |
| [RFC7589 Using the NETCONF Protocol over Transport Layer Security (TLS)][RFC7589] | :white_check_mark: beta      |
| [RFC5277 NETCONF Event Notifications][RFC5277]                                    | :bulb: planned               |
| [RFC5717 Partial Lock Remote Procedure Call (RPC) for NETCONF][RFC5717]           | :bulb: planned               |
| [RFC8071 NETCONF Call Home and RESTCONF Call Home][RFC8071]                       | :bulb: planned               |
| [RFC6243 With-defaults Capability for NETCONF][RFC6243]                           | :bulb: planned               |
| [RFC4743 Using NETCONF over the Simple Object Access Protocol (SOAP)][RFC4743]    | :x: not planned              |
| [RFC4744 Using the NETCONF Protocol over the BEEP][RFC4744]                       | :x: not planned              |

There are other RFC around YANG integration that will be looked at later.

[RFC4743]: https://www.rfc-editor.org/rfc/rfc4743.html
[RFC4744]: https://www.rfc-editor.org/rfc/rfc4744.html
[RFC5277]: https://www.rfc-editor.org/rfc/rfc5277.html
[RFC5717]: https://www.rfc-editor.org/rfc/rfc5717.html
[RFC6241]: https://www.rfc-editor.org/rfc/rfc6241.html
[RFC6242]: https://www.rfc-editor.org/rfc/rfc6242.html
[RFC6243]: https://www.rfc-editor.org/rfc/rfc6243.html
[RFC7589]: https://www.rfc-editor.org/rfc/rfc7589.html
[RFC8071]: https://www.rfc-editor.org/rfc/rfc8071.html

See [TODO.md](TODO.md) for a list of what is left to implement these features.

## Comparison

### Differences from [`github.com/juniper/go-netconf/netconf`](https://pkg.go.dev/github.com/Juniper/go-netconf)

* **Much cleaner, idomatic API, less dumb** I, @nemith, was the original creator of the netconf package and it was my very first Go project and it shows.  There are number of questionable API design, code, and a lot of odd un tested bugs.  Really this rewrite was created to fix this.
* **No impled vendor ownership** Moving the project out of the `Juniper` organization allowes better control over the project, less implied support (or lack there of), and hopefully more contributions.
* **Transports are implemented in their own packages.**  This means if you are not using SSH or TLS you don't need to bring in the underlying depdendencies into your binary.
* **Stream based transports.**  Should mean less memory usage and much less allocation bringing overall performance higher.

### Differences from [`github.com/scrapli/scrapligo/driver/netconf`](https://pkg.go.dev/github.com/scrapli/scrapligo/driver/netconf)

Scrapligo driver is quite good and way better than the original juniper project.  However this new package concentrates more on RFC correctness and implementing some of the more advanced RFC features like call-home and event notifications.  If there is a desire there could be some callaboration with scrapligo in the future.

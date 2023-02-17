# TODO

## Before beta

- [X] Upgrade transport based on versions (1.0->1.1 upgrade to chunked)
- [X] Cleanup request/response API (Session.Do, session.Call?)
- [ ] Convert rpc errors to go errors
- [X] shutdown / close
- [ ] all RFC6241 operations (methods + op structs?)
- [ ] unit tests (>80% coverage?)
  - [ ] operations
  - [ ] server close, shutdown, etc
- [ ] More linting
  - [ ] renable errcheck
  - [ ] errorlint

## Before 0.1.0 release

- [ ] benchmark against juniper/netconf / scrapligo
- [ ] filter support
- [~] TLS support
- [ ] Notification handler support
- [ ] Capability creation/query API
- [X] github actions (CI)

### Future

- [ ] Call Home support
- [ ] nccurl command to issue rpc requests from the cli
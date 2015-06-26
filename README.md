A URL shortener written in Go

### TODO

- [ ] query database for id number, not short
- [ ] add checksum to prevent iterating over all urls
- [ ] sanitize for base64 encoded javascript
  - already sanitizes for plain js
- [ ] fix css
  - look into Go template rendering
- [ ] keep http(s) prefix added to URL

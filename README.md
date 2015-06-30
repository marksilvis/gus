A URL shortener written in Go

### TODO

- [ ] add regex to check for valid url
- [ ] add checksum to prevent iterating over all urls
- [ ] sanitize for base64 encoded javascript
  - already sanitizes for plain js
- [ ] keep http(s) prefix added to URL
  - currently redirects to https

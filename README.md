[![Go Report Card](https://goreportcard.com/badge/github.com/onodera-punpun/punfed)](https://goreportcard.com/report/github.com/onodera-punpun/punfed)

punfed - Personalized Caddy HTTP POST upload plugin


## DESCRIPTION

This is my fork of [caddy.upload](https://github.com/wmark/caddy.upload) mainly for *personal use* on https://punpun.xyz/upload.

I mainly removed tons of stuff, made filenames always randomized (like how pomf.se or sr.ht does it) and added the ability to "protect" your server by requireing a key. Again, like how sr.ht does it.

Oh and it also always returns the file URL now.

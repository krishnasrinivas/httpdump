# httpdump
HTTP reverse proxy that dumps the http headers and body

usage: httpdump -l :8000 -f localhost:9000
       This listens on port 8000 and forwards the request to localhost:9000

# Another Cloudflare DDNS
This is a quick and dirty DDNS implementation in Go for Cloudflare.

## Usage
### docker-compose
```
version: '3.4'
services:
  ddns:
    image: orsanawwad/cloudflare-ddns:latest
    environment:
      - CFKEY=YOUR_CLOUDFLARE_API_KEY
      - CFUSER=YOUR_EMAIL_ADDRESS
      - CFZONE=YOUR_DOMAIN_TO_UPDATE
      - CFHOSTS=LIST_OF_SUBDOMAINS
      - TICKTIME=GO_FORMAT_TIME_STRING
```
CFHOSTS format
```
"subdomain1.domain.tld subdomain2.domain.tld"
```
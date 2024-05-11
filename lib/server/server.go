package server

import (
  "strings"
  "net/http"
)

func ParseSubdomain(domain string, r *http.Request) string {
  hostname := strings.Split(r.Host, ":")[0]
  if domain == hostname {
    return ""
  }
  subdomain := strings.Replace(hostname, "." + domain, "", 1)
  return subdomain
}

func HandleSubdomain(domain string, subdomain string, w http.ResponseWriter, r *http.Request, handler func(w http.ResponseWriter, r *http.Request)) {
  if subdomain == ParseSubdomain(domain, r) {
    handler(w, r)
  }
}

func HandlePath(path string, w http.ResponseWriter, r *http.Request, handler func(w http.ResponseWriter, r *http.Request)) {
  if r.URL.Path == path {
    handler(w, r)
  }
}

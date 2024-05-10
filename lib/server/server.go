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

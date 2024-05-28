package server

import (
	"fmt"
	"io"
	"log"
  "regexp"
	"net/http"
  "strings"
  "time"
)

var re = regexp.MustCompile(`^((.*)\.)?localhost(:\d+)$`)
func ParseSubdomain(domain string, r *http.Request) string {
  if domain != "localhost" {
    isLocalhost := re.MatchString(r.Host)
    if isLocalhost {
      domain = "localhost"
    }
  }

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

func HandleSubdomainRegexp(domain string, subdomain *regexp.Regexp, w http.ResponseWriter, r *http.Request, handler func(w http.ResponseWriter, r *http.Request, match []string)) {
  sub := ParseSubdomain(domain, r)
  match := subdomain.FindStringSubmatch(sub)
  if match != nil {
    handler(w, r, match)
  }
}

func HandlePath(path string, w http.ResponseWriter, r *http.Request, handler func(w http.ResponseWriter, r *http.Request)) {
  if r.URL.Path == path {
    handler(w, r)
  }
}

func HandleProxy(w http.ResponseWriter, r *http.Request, targetHost string) {
  client := &http.Client{Timeout: 30 * time.Second}
  
  path := r.URL.Path
  if r.URL.RawQuery != "" {
    path = path + "?" + r.URL.RawQuery
  }
  targetURL := fmt.Sprintf("%s%s", targetHost, path)

  req, err := http.NewRequest(r.Method, targetURL, r.Body)
  if err != nil {
    log.Printf("Error preparing request [%s]: %v", targetURL, err)
    http.Error(w, "Fail preparing request", http.StatusInternalServerError)
    return
  }

  req.Header = r.Header
  response, err := client.Do(req)
  if err != nil {
    log.Printf("Error forwarding request [%s]: %v", targetURL, err)
    http.Error(w, "Fail forwarding request", http.StatusInternalServerError)
    return
  }
  defer response.Body.Close()

  for k, v := range response.Header {
    for _, vv := range v {
      w.Header().Add(k, vv)
    }
  }
  w.WriteHeader(response.StatusCode)
  _, err = io.Copy(w, response.Body)
  if err != nil {
    log.Printf("Error copying response [%s]: %v", targetURL, err)
  }
}

package main

import (
	"fmt"
	"regexp"
	"github.com/navetacandra/oide/lib/server"
	"database/sql"
	"net/http"
	_ "github.com/lib/pq"
	"github.com/navetacandra/oide/lib/server/git" // credits: https://github.com/asim
)

const domain = "navetacandraa.my.id"
var dockerProxyPattern = regexp.MustCompile(`^([\d\w]+)\-(\d+)$`)

func GetConnection(host string, port int, user string, password string, dbname string) *sql.DB {
  db, err := sql.Open(
    "postgres", 
    fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname),
  )
  if err != nil {
    panic(err)
  }
  return db
}

func main() {
  db := GetConnection("localhost", 5432, "postgres", "postgres", "oide")
  defer db.Close()

  // docker.CreateDocker(db, "navetacandra")

  http.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
    handled := false
    server.HandleSubdomain(domain, "", w, r, func(w http.ResponseWriter, r *http.Request) {
      server.HandlePath("/", w, r, func(w http.ResponseWriter, r *http.Request) {   
        fmt.Fprintf(w, "Hello, World!\nAccess in: %s", server.ParseSubdomain(domain, r))
      })
      server.HandlePath("/hello", w, r, func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
      })
      server.HandlePath("/error", w, r, func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "Error aja", http.StatusInternalServerError)
      })
      handled = true
    })

    server.HandleSubdomain(domain, "git", w, r, func(w http.ResponseWriter, r *http.Request) {
      git.Handler(w, r, func(dir string, repo string, branch string) {
        fmt.Printf("Pushed to %s:%s to %s", repo, branch, dir)
      })
      handled = true
    })

    server.HandleSubdomain(domain, "proxy", w, r, func(w http.ResponseWriter, r *http.Request) {
      server.HandleProxy(w, r, "http://localhost:8080")
      handled = true
    })

    if handled {
      return
    }

    subdomain := server.ParseSubdomain(domain, r)
    dockerProxy := dockerProxyPattern.FindStringSubmatch(subdomain)
    
    if dockerProxy != nil {
      var ip string
      row := db.QueryRow("SELECT container_ip FROM containers WHERE user_id=(SELECT id FROM users WHERE username=$1)", dockerProxy[1])
      err := row.Scan(&ip)

      if err != nil {
        http.Error(w, fmt.Sprintf("Failed get container: %s", err.Error()), http.StatusInternalServerError)
        return
      }

      server.HandleProxy(w, r, fmt.Sprintf("http://%s:%s", ip, dockerProxy[2]))
    }
  })

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    panic(err)
  }
}

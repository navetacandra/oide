package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/lib/pq"
	"github.com/navetacandra/oide/lib/cloudshell"
	"github.com/navetacandra/oide/lib/server"
	"github.com/navetacandra/oide/lib/server/git" // credits: https://github.com/asim
)

type FileCache struct {
  path string
  content []byte
}

const domain = "navetacandraa.my.id"
var dockerProxyPattern = regexp.MustCompile(`^([\d\w]+)\-(\d+)$`)
var caches = []FileCache{}

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

func GetFile(path string) ([]byte, error) {
  for _, file := range caches {
    if file.path == path {
      return file.content, nil
    }
  }
  content, err := os.ReadFile(path)
  if err != nil {
    return nil, err
  }
  caches = append(caches, FileCache{path, content})
  return content, nil
}

func PreventDirTree(fs http.Handler) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    subdomain := server.ParseSubdomain(domain, r)
    if subdomain == "" {
      if r.URL.Path[len(r.URL.Path)-1] == '/' {
        http.NotFound(w, r)
        return
      }
      fs.ServeHTTP(w, r)
      return
    }
    http.NotFound(w, r)
  }
}

func RenderHtml(path string) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    content, err := GetFile(path)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    fmt.Fprintf(w, string(content))
  }
}

func main() {
  db := GetConnection("localhost", 5432, "postgres", "postgres", "oide")
  defer db.Close()

  assetsFile := http.FileServer(http.Dir("./web/assets/"))
  http.Handle("/assets/*", PreventDirTree(http.StripPrefix("/assets/", assetsFile)))

  http.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
    handled := false
    server.HandleSubdomain(domain, "", w, r, func(w http.ResponseWriter, r *http.Request) {
      server.HandlePath("/", w, r, RenderHtml("./web/home.html"))
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

    server.HandleSubdomain(domain, "ssh", w, r, func(w http.ResponseWriter, r *http.Request) {
      cloudshell.Handle(w, r, 10 * time.Second, []string{"bash", "-l"})
      handled = true
    })

    server.HandleSubdomain(domain, "proxy", w, r, func(w http.ResponseWriter, r *http.Request) {
      server.HandleProxy(w, r, "http://localhost:8080")
      handled = true
    })

    server.HandleSubdomainRegexp(domain, dockerProxyPattern, w, r, func(w http.ResponseWriter, r *http.Request, match []string) {
      var ip string
      row := db.QueryRow("SELECT container_ip FROM containers WHERE user_id=(SELECT id FROM users WHERE username=$1)", match[1])
      err := row.Scan(&ip)

      if err != nil {
        http.Error(w, fmt.Sprintf("Failed get container: %s", err.Error()), http.StatusInternalServerError)
        return
      }
      server.HandleProxy(w, r, fmt.Sprintf("http://%s:%s", ip, match[2]))
      handled = true
    })

    if handled {
      return
    }

    http.NotFound(w, r)
  })

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    panic(err)
  }
}

package main

import (
  "fmt"
  "github.com/navetacandra/oide/lib/server"
  "net/http"
  "database/sql"
  _ "github.com/lib/pq"
)

const domain = "localhost"

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

  http.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
    server.HandleSubdomain(domain, "", w, r, func(w http.ResponseWriter, r *http.Request) {
      server.HandlePath("/", w, r, func(w http.ResponseWriter, r *http.Request) {   
        fmt.Fprintf(w, "Hello, World!\nAccess in: %s", server.ParseSubdomain("localhost", r))
      })
      server.HandlePath("/hello", w, r, func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
      })
    })
    server.HandleSubdomain(domain, "proxy", w, r, func(w http.ResponseWriter, r *http.Request) {
      server.HandleProxy(w, r, "http://localhost:8080")
    })
  })

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    panic(err)
  }
}

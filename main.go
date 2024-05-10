package main

import (
  "fmt"
  "github.com/navetacandra/oide/lib/server"
  "net/http"
)

func main() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!\nAccess in: %s", server.ParseSubdomain("localhost", r))
  })

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    panic(err)
  }
}

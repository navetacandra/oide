package cloudshell

import (
	"log"
	"os"
	"sync"
	"time"
	"os/exec"
	"net/http"
	"encoding/json"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
  CheckOrigin: func(r *http.Request) bool {
    return true
  },
}

func Handle(w http.ResponseWriter, r *http.Request, keepAliveTimeout time.Duration, command []string) {
  conn, err := wsUpgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println(err)
    return
  }

  cmd := exec.Command(command[0], command[1:]...)
  cmd.Env = os.Environ()
  tty, err := pty.Start(cmd)
  if err != nil {
    log.Println(err)
    return
  }

  defer func(){
    if err := cmd.Process.Kill(); err != nil {
      log.Println(err)
    }
    if _, err := cmd.Process.Wait(); err != nil {
      log.Println(err)
    }
    if err := conn.Close(); err != nil {
      log.Println(err)
    }
  }()

  var connectionClosed bool
  var waiter sync.WaitGroup
  waiter.Add(1)

  // keepalive
  lastPong := time.Now()
  conn.SetPongHandler(func(_ string) error {
    lastPong = time.Now()
    return nil
  })

  go func() {
    for {
      if err := conn.WriteMessage(websocket.PingMessage, []byte("keepalive")); err != nil {
        return
      }
      time.Sleep(keepAliveTimeout/2)
      if time.Now().Sub(lastPong) > keepAliveTimeout {
        waiter.Done()
        return
      }
    }
  }()

  // tty -> conn
  go func() {
    for {
      if connectionClosed {
        return
      }
      buf := make([]byte, 1024)
      n, err := tty.Read(buf)
      if err != nil {
        waiter.Done()
        return
      }
      err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
      if err != nil {
        log.Println(err)
        return
      }
    }
  }()

  // conn -> tty
  go func() {
    for {
      mtype, msg, err := conn.ReadMessage()
      if err != nil {
        return
      }
      if mtype == websocket.BinaryMessage {
        size := &TTYSize{}
        err := json.Unmarshal(msg, size)
        if err == nil {
          pty.Setsize(tty, &pty.Winsize{
            Rows: size.Rows,
            Cols: size.Cols,
          })
        }
        continue
      }
      _, err = tty.Write(msg)
      if err != nil {
        log.Println(err)
        return
      }
    }
  }()
  
  waiter.Wait()
  connectionClosed = true
}

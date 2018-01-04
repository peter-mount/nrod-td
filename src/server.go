// The rest server

package main

import (
  //"encoding/json"
  "fmt"
  "log"
  "net/http"
  "github.com/gorilla/mux"
)

type Server struct {
  // Port to listen to
  Port    int
  // ===== Internal
  router  *mux.Router
}

func serverInit() {
  // If not defined then use port 80
  if settings.Server.Port < 1 || settings.Server.Port > 65534 {
    settings.Server.Port = 8080
  }

  settings.Server.router = mux.NewRouter()
}

func serverStart() {
  log.Printf( "Listening on port %d\n", settings.Server.Port )
  fatalOnError( http.ListenAndServe( fmt.Sprintf( ":%d", settings.Server.Port ), settings.Server.router ) )
}

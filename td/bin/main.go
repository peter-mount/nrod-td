package main

import (
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/nrod-td/td"
  "log"
)

func main() {
  err := kernel.Launch( &td.TD{} )
  if err != nil {
    log.Fatal( err )
  }
}

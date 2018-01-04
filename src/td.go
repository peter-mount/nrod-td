package main

import (
  "flag"
  "log"
  "os"
  "time"
)

func main() {
  log.Println( "TD v0.1" )

  help := flag.Bool( "h", false, "Show help" )
  yamlFile := flag.String( "f", "/config.yaml", "The config file to use" )

  flag.Parse()

  if( *help ) {
    flag.PrintDefaults()
    os.Exit(0)
  }

  // Load config
  loadConfig( yamlFile )

//  if( !settings.Http.enabled && !settings.Stomp.enabled ) {
//    log.Fatal( "No message source configured, bailing out" )
//  }

  amqpConnect()

//  if( settings.Http.enabled ) {
//    stompRun()
//  }

  // Now keep running forever
  for {
    time.Sleep(time.Minute)
  }
}

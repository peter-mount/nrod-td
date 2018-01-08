package main

import (
  "github.com/peter-mount/golib/rabbitmq"
  "github.com/peter-mount/golib/statistics"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "log"
  "path/filepath"
)

type Config struct {
  Debug   bool                    // Debug logging
  Stats   statistics.Statistics   // Statistics
  Amqp    rabbitmq.RabbitMQ       // RabbitMQ config
  Server  Server                  // Rest server
  Td      TD                      // TD
}

var settings Config

func loadConfig( configFile *string ) {
  filename, _ := filepath.Abs( *configFile )
  log.Println( "Loading config:", filename )

  yml, err := ioutil.ReadFile( filename )
  fatalOnError( err )

  //settings := Config{}
  err = yaml.Unmarshal( yml, &settings )
  fatalOnError( err )

  debug( "Config: %+v\n", settings )

  // Call each supported init method so they can play with the config
  serverInit()
  settings.Stats.Configure()
  //statsInit()
  tdInit()
}

// log.Println() only if debug is enabled
func debug( v ...interface{} ) {
  if( settings.Debug ) {
    log.Println( v... )
  }
}

// helper, log fatal if err is not nil
func fatalOnError( err interface{} ) {
  if( err != nil ) {
    log.Fatal( err )
  }
}

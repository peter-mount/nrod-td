// Handle the connection to the remote AMQP server to receive messages

package main

import (
  "github.com/streadway/amqp"
  "log"
  "time"
)

type AMQP struct {
  Url                 string  `yaml:"url"`
  Exchange            string  `yaml:"exchange"`
  ConnectionName      string  `yaml:"connectionName"`
  HeartBeat           int     `yaml:"heartBeat"`
  Product             string  `yaml:"product"`
  Version             string  `yaml:"version"`
  // ===== Internal
  connection     *amqp.Connection  `yaml:"-"`  // amqp connection
  channel        *amqp.Channel     `yaml:"-"`  // amqp channel
}

// called by main() ensure mandatory config is present
func amqpInit( ) {
  if settings.Amqp.Url == "" {
    log.Fatal( "amqp.url is mandatory" )
  }

  if settings.Amqp.Exchange == "" {
    settings.Amqp.Exchange = "amq.topic"
  }

  if settings.Amqp.HeartBeat == 0 {
    settings.Amqp.HeartBeat = 10
  }

  if settings.Amqp.Product == "" {
    settings.Amqp.Product = "Area51 GO"
  }

  if settings.Amqp.Version == "" {
    settings.Amqp.Version = "0.1β"
  }

}

// Connect to amqp server as necessary
func amqpConnect( ) {
  debug( "Connecting to amqp" )

  // Connect using the amqp url
  /*
  if settings.Amqp.ConnectionName == "" {
    connection, err := amqp.Dial( settings.Amqp.Url )
    fatalOnError( err )
    settings.Amqp.connection = connection
  } else {
  */
    // Use the user provided client name
    connection, err := amqp.DialConfig( settings.Amqp.Url, amqp.Config{
      Heartbeat:  time.Duration( settings.Amqp.HeartBeat ) * time.Second,
      Properties: amqp.Table{
        "product": settings.Amqp.Product,
        "version": settings.Amqp.Version,
        "connection_name": settings.Amqp.ConnectionName,
      },
      Locale: "en_US",
      } )
    fatalOnError( err )
    settings.Amqp.connection = connection
  //}

  // To cleanly shutdown by flushing kernel buffers, make sure to close and
  // wait for the response.
  //defer rabbit.connection.Close()

  // Most operations happen on a channel.  If any error is returned on a
  // channel, the channel will no longer be valid, throw it away and try with
  // a different channel.  If you use many channels, it's useful for the
  // server to
  channel, err := settings.Amqp.connection.Channel()
  fatalOnError( err )
  settings.Amqp.channel = channel

  debug( "AMQP Connected" )

  fatalOnError( settings.Amqp.channel.ExchangeDeclare( settings.Amqp.Exchange, "topic", true, false, false, false, nil) )
}

// Publish a message
func amqpPublish( routingKey string, msg []byte ) {
  debug( "Publishing to ", settings.Amqp.Exchange, routingKey )

  fatalOnError( settings.Amqp.channel.Publish(
    settings.Amqp.Exchange,
    routingKey,
    false,
    false,
    amqp.Publishing{
      Body: msg,
    }) )
}

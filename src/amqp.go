// Handle the connection to the remote AMQP server to receive messages

package main

import (
  "log"
  "github.com/streadway/amqp"
)

type AMQP struct {
  Url         string `yaml:"url"`
  Exchange    string `yaml:"exchange"`
  connection  *amqp.Connection  `yaml:"-"`  // amqp connection
  channel     *amqp.Channel     `yaml:"-"`  // amqp channel
}

// called by main() ensure mandatory config is present
func amqpInit( ) {
  if( settings.Amqp.Url == "" ) {
    log.Fatal( "amqp.url is mandatory" )
  }

  if( settings.Amqp.Exchange == "" ) {
    settings.Amqp.Exchange = "amq.topic"
  }
}

// Connect to amqp server as necessary
func amqpConnect( ) {
  debug( "Connecting to amqp" )

  // Connect using the amqp url
  connection, err := amqp.Dial( settings.Amqp.Url )
  fatalOnError( err )
  settings.Amqp.connection = connection

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

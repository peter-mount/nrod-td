package td

import (
  "encoding/json"
  "github.com/peter-mount/golib/statistics"
  "log"
  "strconv"
  "sync"
  "time"
)

type TDFeed struct {
  // Required: The Queue name to declare
  Queue       string      `yaml:"queue"`
  // Required: The routing key to bind to
  RoutingKey  string      `yaml:"routingKey"`
  // The exchange to use, defaults to "amq.topic"
  Exchange    string      `yaml:"exchange"`
  // Declare the queue as durable
  Durable     bool        `yaml:"durable"`
  AutoDelete  bool        `yaml:"autoDelete"`
  Exclusive   bool        `yaml:"exclusive"`
  // Consumer Tag, defaults to "", shows in RabbitMQ management plugin
  ConsumerTag string      `yaml:"consumerTag"`
  // ===== Internal
  areas     map[string]*TDArea
  mutex    *sync.Mutex
  // The timestamp of the last operation
  timestamp int64
  // Timestamp of the last reset
  reset     int64
}

type TDMessage struct {
  CA  *CAMessage  `json:"CA_MSG"`
  CB  *CBMessage  `json:"CB_MSG"`
  CC  *CCMessage  `json:"CC_MSG"`
  CT  *CTMessage  `json:"CT_MSG"`
  SF  *SMessage   `json:"SF_MSG"`
  SG  *SMessage   `json:"SG_MSG"`
  SH  *SMessage   `json:"SH_MSG"`
}

type TDArea struct {
  name      string
  // The timestamp of the last operation
  timestamp int64
  // Map of berths
  berths    map[string]*TDBerth
  // Signal data
  signals   map[string]string
  // heartBeat from CT message
  heartBeat string
}

type TDBerth struct {
  // The timestamp of the last operation
  Timestamp int64   `json:"timestamp"`
  // Descr on this berth
  Descr     string  `json:"descr"`
}

// Return the NR timestamp as a unix timestamp
// t timestamp string
// a Statistic id, "" for none (i.e. berths)
func tdParseTimestamp( t string, a string ) int64 {
  n, err := strconv.ParseInt( t, 10, 64 )
  if err == nil {
    // NR feed is in Java time (millis) so convert to Unix time (seconds)
    n := n / int64(1000)

    if a != "" {
      statistics.Set( "td." + a, time.Now().Unix() - n )
    }

    return n
  }
  return 0
}

func (a *TDFeed) update( t string ) *TDFeed {
  a.timestamp = tdParseTimestamp( t, "all" )
  return a
}

func (a *TDArea) update( t string ) *TDArea {
  a.timestamp = tdParseTimestamp( t, a.name )
  return a
}

func (t *TDFeed) area( a string ) *TDArea {
  if val, ok := t.areas[ a ]; ok {
    return val
  }

  var v *TDArea = new( TDArea )
  v.name = a
  v.berths = make( map[string]*TDBerth )
  v.signals = make( map[string]string )
  t.areas[ a ] = v

  return v
}

func (b *TDBerth) clone() *TDBerth {
  var r = new( TDBerth )
  r.Timestamp = b.Timestamp
  r.Descr = b.Descr
  return r
}

func (b *TDBerth) update( t string, d string ) *TDBerth {
  b.Timestamp = tdParseTimestamp( t, "" )
  b.Descr = d
  return b
}

func (a *TDArea) berth( b string ) *TDBerth {
  if val, ok := a.berths[ b ]; ok {
    return val
  }

  a.berths[ b ] = new( TDBerth )
  return a.berths[ b ]
}

func (s *TD) tdStart() error {
  channel, err := s.Amqp.NewChannel()
  if err != nil {
    return err
  }

  _, err = s.Amqp.QueueDeclare(
    channel,
    s.Td.Queue,
    s.Td.Durable,
    s.Td.AutoDelete,
    s.Td.Exclusive,
    // wait & no args
    false, nil )
  if err != nil {
    return err
  }

  err = s.Amqp.QueueBind(
    channel,
    s.Td.Queue,
    s.Td.RoutingKey,
    s.Td.Exchange,
    // wait & no args
    false, nil )
  if err != nil {
    return err
  }

  queue, err := s.Amqp.Consume(
    channel,
    s.Td.Queue,
    s.Td.ConsumerTag,
    // Don't auto Ack as we do this after processing
    false,
    // Exclusive so we are the only permitted consumer on this queue
    true,
    // noLocal is always false on RabbitMQ as its unsupported
    false,
    // wait & no args
    false, nil )
  if err != nil {
    return err
  }

  go func(  ) {
    for {
      msg, ok := <-queue
      if !ok {
        log.Fatal( "channel closed, see reconnect example" )
      }

      var dat []*TDMessage
      err := json.Unmarshal( msg.Body, &dat )
      if err != nil {
        log.Fatal( err )
      }

      for _, tdmsg := range dat {
        // TODO Must find a better way of handling this
        if tdmsg.CA != nil { tdmsg.CA.handle( s ) }
        if tdmsg.CB != nil { tdmsg.CB.handle( s ) }
        if tdmsg.CC != nil { tdmsg.CC.handle( s ) }
        if tdmsg.CT != nil { tdmsg.CT.handle( s ) }
        if tdmsg.SF != nil { tdmsg.SF.handle( s ) }
        if tdmsg.SG != nil { tdmsg.SG.handle( s ) }
        if tdmsg.SH != nil { tdmsg.SH.handle( s ) }
      }

      msg.Ack( false )
    }
  }()

  return nil
}

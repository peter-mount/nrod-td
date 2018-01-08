// Handles the actual TD feed
package main

import (
  "encoding/json"
  "github.com/gorilla/mux"
  "github.com/peter-mount/golib/statistics"
  "log"
  "net/http"
  "sort"
  "strconv"
  "sync"
  "time"
)

type TD struct {
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

func tdInit() {
  if settings.Td.Queue == "" {
    log.Fatal( "Queue name is required" )
  }

  if settings.Td.RoutingKey == "" {
    log.Fatal( "RoutingKey is required" )
  }

  if settings.Td.Exchange == "" {
    settings.Td.Exchange = "amq.topic"
  }

  settings.Td.areas = make( map[string]*TDArea )
  settings.Td.mutex = &sync.Mutex{}

  settings.Td.reset = time.Now().Unix()

  settings.Server.router.HandleFunc( "/area", tdGetAreas ).Methods( "GET" )
  settings.Server.router.HandleFunc( "/{id}", tdGetArea ).Methods( "GET" )
}

// Return the NR timestamp as a unix timestamp
// t timestamp string
// a Statistic id, "" for none (i.e. berths)
func tdParseTimestamp( t string, a string ) int64 {
  n, err := strconv.ParseInt( t, 10, 64 )
  if err == nil {
    // NR feed is in Java time (millis) so convert to Unix time (seconds)
    n = n / int64(1000)

    if a != "" {
      statistics.Set( "td." + a, time.Now().Unix() - n )
    }

    return n
  }
  return 0
}

func (a *TD) update( t string ) *TD {
  a.timestamp = tdParseTimestamp( t, "all" )
  return a
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

func (a *TDArea) update( t string ) *TDArea {
  a.timestamp = tdParseTimestamp( t, a.name )
  return a
}

func (t *TD) area( a string ) *TDArea {
  if val, ok := t.areas[ a ]; ok {
    return val
  }

  debug( "New Area", a )

  var v *TDArea = new( TDArea )
  v.name = a
  v.berths = make( map[string]*TDBerth )
  v.signals = make( map[string]string )
  t.areas[ a ] = v

  return v
}

type TDBerth struct {
  // The timestamp of the last operation
  Timestamp int64   `json:"timestamp"`
  // Descr on this berth
  Descr     string  `json:"descr"`
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

type TDMessage struct {
  CA  *CAMessage  `json:"CA_MSG"`
  CB  *CBMessage  `json:"CB_MSG"`
  CC  *CCMessage  `json:"CC_MSG"`
  CT  *CTMessage  `json:"CT_MSG"`
  SF  *SMessage   `json:"SF_MSG"`
  SG  *SMessage   `json:"SG_MSG"`
  SH  *SMessage   `json:"SH_MSG"`
}

type SMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  Type  string  `json:"msg_type"`
  Addr  string  `json:"address"`
  Data  string  `json:"data"`
}

func (m *SMessage) handle() {
  settings.Td.mutex.Lock()
  settings.Td.update( m.Time ).area( m.Area ).signals[ m.Addr ] = m.Data
  settings.Td.mutex.Unlock()
}

type CAMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  From  string  `json:"from"`
  To    string  `json:"to"`
  Descr string  `json:"descr"`
}

func (m *CAMessage) handle() {
  settings.Td.mutex.Lock()
  var a *TDArea = settings.Td.update( m.Time ).area( m.Area ).update( m.Time )
  a.berth( m.From ).update( m.Time, "" )
  a.berth( m.To ).update( m.Time, m.Descr )
  settings.Td.mutex.Unlock()
}

type CBMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  From  string  `json:"from"`
  Descr string  `json:"descr"`
}

func (m *CBMessage) handle() {
  settings.Td.mutex.Lock()
  settings.Td.update( m.Time ).area( m.Area ).update( m.Time ).berth( m.From ).update( m.Time, "" )
  settings.Td.mutex.Unlock()
}

type CCMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  To    string  `json:"to"`
  Descr string  `json:"descr"`
}

func (m *CCMessage) handle() {
  settings.Td.mutex.Lock()
  settings.Td.update( m.Time ).area( m.Area ).update( m.Time ).berth( m.To ).update( m.Time, m.Descr )
  settings.Td.mutex.Unlock()
}

type CTMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  RepTM string  `json:"report_time"`
}

func (m *CTMessage) handle() {
  settings.Td.mutex.Lock()
  settings.Td.update( m.Time ).area( m.Area ).heartBeat = m.RepTM
  settings.Td.mutex.Unlock()
}

func tdStart() {
  _, err := settings.Amqp.QueueDeclare(
    settings.Td.Queue,
    settings.Td.Durable,
    settings.Td.AutoDelete,
    settings.Td.Exclusive,
    // wait & no args
    false, nil )
  fatalOnError( err )

  fatalOnError( settings.Amqp.QueueBind(
    settings.Td.Queue,
    settings.Td.RoutingKey,
    settings.Td.Exchange,
    // wait & no args
    false, nil ) )

  queue, err := settings.Amqp.Consume(
    settings.Td.Queue,
    settings.Td.ConsumerTag,
    // Don't auto Ack as we do this after processing
    false,
    // Exclusive so we are the only permitted consumer on this queue
    true,
    // noLocal is always false on RabbitMQ as its unsupported
    false,
    // wait & no args
    false, nil )
  fatalOnError( err )

  go func(  ) {
    for {
      msg, ok := <-queue
      if !ok {
        log.Fatal( "channel closed, see reconnect example" )
      }

      var dat []*TDMessage
      fatalOnError( json.Unmarshal( msg.Body, &dat ) )

      for _, tdmsg := range dat {
        // TODO Must find a better way of handling this
        if tdmsg.CA != nil { tdmsg.CA.handle() }
        if tdmsg.CB != nil { tdmsg.CB.handle() }
        if tdmsg.CC != nil { tdmsg.CC.handle() }
        if tdmsg.CT != nil { tdmsg.CT.handle() }
        if tdmsg.SF != nil { tdmsg.SF.handle() }
        if tdmsg.SG != nil { tdmsg.SG.handle() }
        if tdmsg.SH != nil { tdmsg.SH.handle() }
      }

      msg.Ack( false )
    }
  }(  )
}

type LatencyOut struct {
  Value       int64       `json:"value"`
  Max         int64       `json:"max"`
  Min         int64       `json:"min"`
}

type AreasOut struct {
  Areas     []string      `json:"areas"`
  Timestamp   int64       `json:"lastUpdate"`
  Reset       int64       `json:"reset"`
  Total       int         `json:"total"`
  Berths      int         `json:"berths"`
  Latency    *statistics.Statistic   `json:"latency"`
}

// Return
func tdGetAreas( w http.ResponseWriter, r *http.Request ) {
  var result = new( AreasOut )
  result.Timestamp = settings.Td.timestamp
  result.Reset = settings.Td.reset

  var berths = 0

  settings.Td.mutex.Lock()
  for name, area := range settings.Td.areas {
    result.Areas = append( result.Areas, name )
    berths += len( area.berths )
  }
  settings.Td.mutex.Unlock()

  sort.Strings( result.Areas )

  result.Total = len( result.Areas )
  result.Berths = berths

  result.Latency = statistics.Get( "td.all" )

  settings.Server.setJsonResponse( w, 0, result.Timestamp, 60 )
  json.NewEncoder(w).Encode( result )
}

type AreaOut struct {
  Name        string                `json:"name"`
  Berths      map[string]*TDBerth   `json:"berths"`
  Signals     map[string]string     `json:"signals"`
  Timestamp   int64                 `json:"lastUpdate"`
  HeartBeat   string                `json:"heartBeat"`
  Occupied    int                   `json:"occupied"`
  Total       int                   `json:"total"`
  Latency    *statistics.Statistic  `json:"latency"`
}

func tdGetArea( w http.ResponseWriter, r *http.Request ) {
  var params = mux.Vars( r )

  var result = new( AreaOut )
  result.Name = params[ "id" ]
  result.Berths = make( map[string]*TDBerth )
  result.Signals = make( map[string]string )

  settings.Td.mutex.Lock()
  if area, ok := settings.Td.areas[ result.Name ]; ok {
    result.Timestamp = area.timestamp
    result.HeartBeat = area.heartBeat

    for name, berth := range area.berths {
      if berth.Descr != "" {
        result.Berths[ name ] = berth.clone()
      }
    }

    for addr, data := range area.signals {
      result.Signals[ addr ] = data
    }

    result.Total = len( area.berths )
  }
  settings.Td.mutex.Unlock()

  result.Occupied = len( result.Berths )

  var sc = 200
  if result.Timestamp == 0 {
    sc = 404
  } else {
    result.Latency = statistics.Get( "td." + result.Name )
  }

  settings.Server.setJsonResponse( w, sc, result.Timestamp, 10 )
  json.NewEncoder(w).Encode( result )
}

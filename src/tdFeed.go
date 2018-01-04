// Handles the actual TD feed
package main

import (
  "encoding/json"
  "log"
  "strconv"
  "sync"
  "time"
)

type TD struct {
  areas     map[string]*TDArea
  mutex    *sync.Mutex
  // The timestamp of the last operation
  timestamp int64
}

func tdInit() {
  settings.Td.areas = make( map[string]*TDArea )
  settings.Td.mutex = &sync.Mutex{}
}

func (a *TD) update( t string ) *TD {

  n, err := strconv.ParseInt( t, 10, 64 )
  if err == nil {
    a.timestamp = n

    // Record the latency. count will be the number of messages processed for all
    // Note n is Java time so in milliseconds hence *1000
    statsSet( "td.all", (time.Now().Unix()*1000) - n )
  }

  return a
}

type TDArea struct {
  name      string
  // The timestamp of the last operation
  timestamp int64
  // Map of berths
  berths    map[string]*TDBerth
}

func (a *TDArea) update( t string ) *TDArea {

  n, err := strconv.ParseInt( t, 10, 64 )
  if err == nil {
    a.timestamp = n

    // Record the latency. count will be the number of messages processed for this area
    // Note n is Java time so in milliseconds hence *1000
    statsSet( "td." + a.name, (time.Now().Unix()*1000) - n )
  }

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
  t.areas[ a ] = v

  return v
}

type TDBerth struct {
  // The timestamp of the last operation
  Timestamp int64   `json:"timestamp"`
  // Descr on this berth
  Descr     string  `json:"descr"`
}

func (b *TDBerth) update( t string, d string ) *TDBerth {
  n, err := strconv.ParseInt( t, 10, 64 )
  if err == nil {
    b.Timestamp = n
  }
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
  //SF  *SMessage   `json:"SF_MSG"`
  //SG  *SMessage   `json:"SG_MSG"`
  //SH  *SMessage   `json:"SH_MSG"`
}

type SMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  Type  string  `json:"msg_type"`
  Addr  string  `json:"address"`
  Data  string  `json:"data"`
}

func (m *SMessage) handle() {
  statsIncr( "td.area.msg" )
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
  //log.Println( "Area", m.Area, "tm", m.RepTM )
  //settings.Td.area( m.Area ).berth( to ).update( m.Time, m.Descr )
}

func tdStart() {
  _, err := settings.Amqp.channel.QueueDeclare( "td", true, false, false, false, nil )
  fatalOnError( err )

  fatalOnError( settings.Amqp.channel.QueueBind( "td", "feed.nrod.td", "amq.topic", false, nil ) )

  queue, err := settings.Amqp.channel.Consume( "td", "td go", false, false, false, false, nil )
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
        if tdmsg.CA != nil { tdmsg.CA.handle() }
        if tdmsg.CB != nil { tdmsg.CB.handle() }
        if tdmsg.CC != nil { tdmsg.CC.handle() }
        if tdmsg.CT != nil { tdmsg.CT.handle() }
      }

      msg.Ack( false )
    }
  }(  )
}

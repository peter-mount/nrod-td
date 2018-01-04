// Handles the actual TD feed
package main

import (
  "sync"
)

type TD struct {
  areas     map[string]*TDArea
  mutex    *sync.Mutex
}

func tdInit() {
  settings.Td.areas = make( map[string]*TDArea )
  settings.Td.mutex = &sync.Mutex{}
}

type TDArea struct {
  // The timestamp of the last operation
  timestamp int64
  // Map of berths
  berths    map[string]*TDBerth
}

func (a *TDArea) update( t int64 ) *TDArea {
  a.timestamp = t
  return a
}

func (t *TD) area( a string ) *TDArea {
  if val, ok := t.areas[ a ]; ok {
    return val
  }

  var v *TDArea = new( TDArea )
  v.berths = make( map[string]*TDBerth )
  t.areas[ a ] = v
  return t.areas[ a ]
}

type TDBerth struct {
  // The timestamp of the last operation
  Timestamp int64   `json:"timestamp"`
  // Descr on this berth
  Descr     string  `json:"descr"`
}

func (b *TDBerth) update( t int64, d string ) *TDBerth {
  b.Timestamp = t
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
  //CT  *CTMessage  `json:"CT_MSG"`
  //SF  *SMessage   `json:"SF_MSG"`
  //SG  *SMessage   `json:"SG_MSG"`
  //SH  *SMessage   `json:"SH_MSG"`
}

type SMessage struct {
  Time  int64   `json:"time"`
  Area  string  `json:"area_id"`
  Type  string  `json:"msg_type"`
  Addr  string  `json:"address"`
  Data  string  `json:"data"`
}

func (m *SMessage) handle() {
  statsIncr( "td.area.msg" )
}

type CAMessage struct {
  Time  int64   `json:"time"`
  Area  string  `json:"area_id"`
  From  string  `json:"from"`
  To    string  `json:"to"`
  Descr string  `json:"descr"`
}

func (m *CAMessage) handle() {
  statsIncr( "td.area.msg" )
  var a *TDArea = settings.Td.area( m.Area )
  a.berth( m.From ).update( m.Time, "" )
  a.berth( m.To ).update( m.Time, m.Descr )
}

type CBMessage struct {
  Time  int64   `json:"time"`
  Area  string  `json:"area_id"`
  From  string  `json:"from"`
  Descr string  `json:"descr"`
}

func (m *CBMessage) handle() {
  statsIncr( "td.area.msg" )
  settings.Td.area( m.Area ).berth( m.From ).update( m.Time, "" )
}

type CCMessage struct {
  Time  int64   `json:"time"`
  Area  string  `json:"area_id"`
  To    string  `json:"to"`
  Descr string  `json:"descr"`
}

func (m *CCMessage) handle() {
  statsIncr( "td.area.msg" )
  settings.Td.area( m.Area ).berth( m.To ).update( m.Time, m.Descr )
}

type CTMessage struct {
  Time  int64   `json:"time"`
  Area  string  `json:"area_id"`
  RepTM string  `json:"report_time"`
}

func (m *CTMessage) handle() {
  //settings.Td.area( m.Area ).berth( to ).update( m.Time, m.Descr )
}

func tdStart() {

}

// Thread safe statistics

package main

import (
  "encoding/json"
  "log"
  "net/http"
  "sync"
)

type Statistics struct {
  // If set then log stats to the log every duration
  Log         bool
  // If set then present /stats endpoint with json output
  Statistics  bool
  // The schedule to use to collect statistics, defaults to every minute
  Schedule    string
  // Predefined Statistics
  stats  map[string]*Statistic
  // ===== Internal
  mutex      *sync.Mutex
}

type Statistic struct {
  // the current value
  Value     int64       `json:"value"`
  // the number of updates
  Count     int64       `json:"count"`
  // The minimum value
  Min       int64       `json:"min"`
  // The maximum value
  Max       int64       `json:"max"`
  // The average value
  Ave       int64       `json:"average"`
  // The sum of all values
  Sum       int64       `json:"sum"`
}

func (s *Statistic) reset() {
  s.Value = 0
  s.Count = 0
  s.Min = int64(^uint64(0) >> 1)
  s.Max = -s.Min - 1
  s.Ave = 0
  s.Sum = 0
}

func (s *Statistic) clone() *Statistic {
  var r *Statistic = new(Statistic)
  r.Value = s.Value
  r.Count = s.Count
  r.Min = s.Min
  r.Max = s.Max
  r.Ave = s.Ave
  r.Sum = s.Sum
  return r
}

func (s *Statistic) update() {
  if( s.Value < s.Min ) {
    s.Min = s.Value
  }
  if( s.Value > s.Max ) {
    s.Max = s.Value
  }

  // protect against /0 - incase count is reset for some reason
  if( s.Count != 0 && s.Sum != 0 ) {
    s.Ave = s.Sum / s.Count
    } else {
      s.Ave = 0
    }
}

func (s *Statistic) set( v int64 ) {
  s.Value = v
  s.Sum += v
  s.Count ++
  s.update()
}

func (s *Statistic) incr( v int64 ) {
  s.Value += v
  s.Sum += v
  s.Count ++
  s.update()
}

func statsInit() {
  settings.Stats.stats = make( map[string]*Statistic )
  settings.Stats.mutex = &sync.Mutex{}

  if( settings.Stats.Schedule == "" ) {
    settings.Stats.Schedule = "0 * * * * *"
  }
  cronAdd( settings.Stats.Schedule, statsRecord )

  // Add /stats endpoint
  if( settings.Stats.Statistics ) {
    settings.Server.router.HandleFunc( "/stats", getStats ).Methods( "GET" )
  }

  debug( "Statistics initialised")
}

// Handler for /stats
func getStats(w http.ResponseWriter, r *http.Request) {
  var stats = make( map[string]*Statistic )

  settings.Stats.mutex.Lock()
  for key,value := range settings.Stats.stats {
    stats[key] = value.clone()
  }
  settings.Stats.mutex.Unlock()

  log.Println( len( stats ) )

  json.NewEncoder(w).Encode( stats )
}

// Record then reset all Statistics
func statsRecord() {
  settings.Stats.mutex.Lock()

  for key,value := range settings.Stats.stats {

    if( settings.Stats.Log ) {
      log.Printf(
        "%s Val %d Count %d Min %d Max %d Sum %d Ave %d\n",
        key,
        value.Value,
        value.Count,
        value.Min,
        value.Max,
        value.Sum,
        value.Ave )
    }

    value.reset()
  }

  settings.Stats.mutex.Unlock()
}

func statsIncr( s string ) {
  statsIncrVal( s, 1 )
}

func statsDecr( s string ) {
  statsIncrVal( s, -1 )
}

// return a statistic creating it as needed
func statsGet( s string ) *Statistic {
  if val, ok := settings.Stats.stats[s]; ok {
    return val
  }
  settings.Stats.stats[s] = new(Statistic)
  settings.Stats.stats[s].reset()
  return settings.Stats.stats[s]
}

func statsIncrVal( s string, v int64 ) {
  settings.Stats.mutex.Lock()
  statsGet( s ).incr( v )
  settings.Stats.mutex.Unlock()
}

func statsSet( s string, v int64 ) {
  settings.Stats.mutex.Lock()
  statsGet( s ).set( v )
  settings.Stats.mutex.Unlock()
}

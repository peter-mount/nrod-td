// Thread safe statistics

package main

import (
  "encoding/json"
  "log"
  "net/http"
  "sync"
  "time"
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
  // The timestamp of the last operation
  Timestamp int64         `json:"timestamp"`
  // the current value
  Value     int64         `json:"value"`
  // the number of updates
  Count     int64         `json:"count"`
  // The minimum value
  Min       int64         `json:"min"`
  // The maximum value
  Max       int64         `json:"max"`
  // The average value
  Ave       int64         `json:"average"`
  // The sum of all values
  Sum       int64         `json:"sum"`
  // Historic data, max 72 entries at 5 minute intervals
  History   []*Statistic  `json:"history,omitempty"`
  // The last 5 minutes data, used to build the history
  lastFive  []*Statistic  `json:"-"`
  latest    *Statistic    `json:"-"`
}

const (
  // 72 entries * STATS_HISTORY_PERIOD m = 6 hours when default schedule of 1 minute
  STATS_MAX_HISTORY = 72
  // Period of history in schedule units
  STATS_HISTORY_PERIOD = 5
)

func (s *Statistic) reset() {
  s.Timestamp = time.Now().Unix()
  s.Value = 0
  s.Count = 0
  s.Min = int64(^uint64(0) >> 1)
  s.Max = -s.Min - 1
  s.Ave = 0
  s.Sum = 0
}

func (s *Statistic) clone() *Statistic {
  var r *Statistic = new(Statistic)
  r.Timestamp = s.Timestamp
  r.Value = s.Value
  r.Count = s.Count
  r.Min = s.Min
  r.Max = s.Max
  r.Ave = s.Ave
  r.Sum = s.Sum
  return r
}

func statsCopyArray( s []*Statistic ) []*Statistic {
  var a []*Statistic
  if len( s ) > 0 {
    for _, v := range s {
      a = append( a, v.clone() )
    }
  }
  return a
}

func (s *Statistic) update() {
  s.Timestamp = time.Now().Unix()
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

func (s *Statistic) recordHistory() {
  // Add to last 5 entries
  s.latest = s.clone()
  s.lastFive = append( s.lastFive, s.latest )
  // If full then collate and push to history
  if len( s.lastFive ) >= STATS_HISTORY_PERIOD {
    // Form new statistoc of sum of all entries within it
    var hist = new(Statistic)
    for _, val := range s.lastFive {
      hist.Value += val.Value
      hist.Count += val.Count
      hist.Sum += val.Sum
    }
    hist.update()
    s.History = append( s.History, hist )
    s.lastFive = nil

    // Keep history down to size
    if len( s.History ) > STATS_MAX_HISTORY {
      s.History = s.History[1:]
    }
  }
}

func statsInit() {
  settings.Stats.stats = make( map[string]*Statistic )
  settings.Stats.mutex = &sync.Mutex{}

  if settings.Stats.Schedule == "" {
    settings.Stats.Schedule = "0 * * * * *"
  }
  cronAdd( settings.Stats.Schedule, statsRecord )

  // Add /stats endpoint
  if settings.Stats.Statistics {
    settings.Server.router.HandleFunc( "/stats", getStats ).Methods( "GET" )
  }

  debug( "Statistics initialised")
}

// Handler for /stats
func getStats(w http.ResponseWriter, r *http.Request) {
  var stats = make( map[string]*Statistic )

  settings.Stats.mutex.Lock()
  for key,value := range settings.Stats.stats {
    if value.latest != nil {
      stats[key] = value.latest.clone()
      stats[key].History = statsCopyArray( value.History )
    }
  }
  settings.Stats.mutex.Unlock()

  json.NewEncoder(w).Encode( stats )
}

// Record then reset all Statistics
func statsRecord() {
  settings.Stats.mutex.Lock()

  for key,value := range settings.Stats.stats {

    // Don't report stats with no submitted values, i.e. Min > Max
    if( value.Min <= value.Max ) {

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

    }

    value.recordHistory()
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
func statsGetOrCreate( s string ) *Statistic {
  if val, ok := settings.Stats.stats[s]; ok {
    return val
  }
  settings.Stats.stats[s] = new(Statistic)
  settings.Stats.stats[s].reset()
  return settings.Stats.stats[s]
}

func statsIncrVal( s string, v int64 ) {
  settings.Stats.mutex.Lock()
  statsGetOrCreate( s ).incr( v )
  settings.Stats.mutex.Unlock()
}

func statsSet( s string, v int64 ) {
  settings.Stats.mutex.Lock()
  statsGetOrCreate( s ).set( v )
  settings.Stats.mutex.Unlock()
}

func statsGet( s string ) *Statistic {
  settings.Stats.mutex.Lock()
  val, ok := settings.Stats.stats[s]
  settings.Stats.mutex.Unlock()
  if( ok ) {
    if val.latest != nil {
      return val.latest.clone()
    }
    return val.clone()
  }
  return nil
}

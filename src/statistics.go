// Thread safe statistics

package main

import (
  "log"
  "sync"
)

type Statistics struct {
  // If set then log stats to the log every duration
  Log         bool
  // The schedule to use to collect statistics, defaults to every minute
  Schedule    string
  // Predefined Statistics
  stats  map[string]*Statistic
  // ===== Internal
  mutex      *sync.Mutex
}

type Statistic struct {
  // the current value
  value     int64
  // the number of updates
  count     int64
  // The minimum value
  min       int64
  // The maximum value
  max       int64
  // The average value
  ave       int64
  // The sum of all values
  sum       int64
}

func (s *Statistic) reset() {
  s.value = 0
  s.count = 0
  s.min = int64(^uint64(0) >> 1)
  s.max = -s.min - 1
  s.ave = 0
  s.sum = 0
}

func (s *Statistic) set( v int64 ) {
  s.value = v
  s.sum += v
  s.count ++
  if( s.value < s.min ) {
    s.min = v
  }
  if( s.value > s.max ) {
    s.max = v
  }

  // protect against /0 - incase count is reset for some reason
  if( s.count != 0 && s.sum != 0 ) {
    s.ave = s.sum / s.count
  } else {
    s.ave = 0
  }
}

func (s *Statistic) incr( v int64 ) {
  s.value += v
  s.sum += v
  s.count ++
  if( s.value < s.min ) {
    s.min = v
  }
  if( s.value > s.max ) {
    s.max = v
  }

  // protect against /0 - incase count is reset for some reason
  if( s.count != 0 && s.sum != 0 ) {
    s.ave = s.sum / s.count
  } else {
    s.ave = 0
  }
}

func statsInit() {
  settings.Stats.stats = make( map[string]*Statistic )
  settings.Stats.mutex = &sync.Mutex{}

  if( settings.Stats.Schedule == "" ) {
    settings.Stats.Schedule = "0 * * * * *"
  }

  cronAdd( settings.Stats.Schedule, statsRecord )

  // Preinitialise stats table if needed

  debug( "Statistics initialised")
}

func statsRecord() {
  log.Println( "Record Tick" )

  settings.Stats.mutex.Lock()

  for key,value := range settings.Stats.stats {

    if( settings.Stats.Log ) {
      log.Printf(
        "%s Val %d Count %d Min %d Max %d Sum %d Ave %d\n",
        key,
        value.value,
        value.count,
        value.min,
        value.max,
        value.sum,
        value.ave )
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

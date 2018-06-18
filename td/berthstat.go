package td

import (
  "github.com/peter-mount/golib/statistics"
)

type berthstat struct {
  areas   int64
  berths  int64
  active  int64
}

// berthstat maintains the statistics td.areas & td.berths so that we have
// an approximate total count of how many areas and berths that are present
// TD feed fails.
func (s *TD) berthstat() {
  var stats berthstat
  stats.collect( s )

  statistics.Set( "td.areas", stats.areas )
  statistics.Set( "td.berths.total", stats.berths )
  statistics.Set( "td.berths.active", stats.active )
}

func (a *berthstat) collect(s *TD) {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()

  a.areas = int64(len( s.Td.areas ))

  for _, area := range s.Td.areas {
    for _, berth := range area.berths {
      a.berths++
      if berth.Descr != "" {
        a.active++
      }
    }
  }
}

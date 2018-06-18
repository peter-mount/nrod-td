package td

import (
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/golib/statistics"
  "sort"
)

// The output format for the /areas
type AreasOut struct {
  Areas     []string                  `json:"areas"`
  Timestamp   int64                   `json:"lastUpdate"`
  Reset       int64                   `json:"reset"`
  Total       int                     `json:"total"`
  Berths      int                     `json:"berths"`
  Latency    *statistics.Statistic    `json:"latency"`
}

// tdGetAreas returns the currently available TD Areas
func (s *TD) tdGetAreas( r *rest.Rest ) error {
  result := s.populateGetAreas()

  sort.Strings( result.Areas )

  result.Latency = statistics.Get( "td.all" )

  r.Status( 200 ).Value( result )
  return nil
}

func (s *TD) populateGetAreas() *AreasOut {
  result := &AreasOut{
    Timestamp: s.Td.timestamp,
    Reset: s.Td.reset,
  }

  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()

  var berths = 0

  for name, area := range s.Td.areas {
    result.Areas = append( result.Areas, name )
    berths += len( area.berths )
  }

  result.Total = len( result.Areas )
  result.Berths = berths

  return result
}

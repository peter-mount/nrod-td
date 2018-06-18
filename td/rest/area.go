package td

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
func ( s *TD ) tdGetAreas( w http.ResponseWriter, r *http.Request ) {
  var result = new( AreasOut )
  result.Timestamp = s.Td.timestamp
  result.Reset = s.Td.reset

  var berths = 0

  s.Td.mutex.Lock()
  for name, area := range s.Td.areas {
    result.Areas = append( result.Areas, name )
    berths += len( area.berths )
  }
  s.Td.mutex.Unlock()

  sort.Strings( result.Areas )

  result.Total = len( result.Areas )
  result.Berths = berths

  result.Latency = statistics.Get( "td.all" )

  s.Server.setJsonResponse( w, 0, result.Timestamp, 60 )
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

func ( s *TD ) tdGetArea( w http.ResponseWriter, r *http.Request ) {
  var params = mux.Vars( r )

  var result = new( AreaOut )
  result.Name = params[ "id" ]
  result.Berths = make( map[string]*TDBerth )
  result.Signals = make( map[string]string )

  s.Td.mutex.Lock()
  if area, ok := s.Td.areas[ result.Name ]; ok {
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
  s.Td.mutex.Unlock()

  result.Occupied = len( result.Berths )

  var sc = 200
  if result.Timestamp == 0 {
    sc = 404
  } else {
    result.Latency = statistics.Get( "td." + result.Name )
  }

  s.Server.setJsonResponse( w, sc, result.Timestamp, 10 )
  json.NewEncoder(w).Encode( result )
}

package td

import (
  "github.com/peter-mount/golib/rest"
)

type AreaOut struct {
  Name        string                `json:"name"`
  Berths      map[string]*TDBerth   `json:"berths"`
  Signals     map[string]string     `json:"signals"`
  Timestamp   int64                 `json:"lastUpdate"`
  HeartBeat   string                `json:"heartBeat"`
  Occupied    int                   `json:"occupied"`
  Total       int                   `json:"total"`
}

func ( s *TD ) tdGetArea( r *rest.Rest ) error {
  result := s.populateGetArea( r.Var( "id" ) )

  result.Occupied = len( result.Berths )

  if result.Timestamp == 0 {
    r.Status( 404 )
  } else {
    r.Status( 200 ).Value( result )
  }

  return nil
}

func (s *TD) populateGetArea( id string ) *AreaOut {
  result := &AreaOut{
    Name: id,
    Berths: make( map[string]*TDBerth ),
    Signals: make( map[string]string ),
  }

  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()

  area, ok := s.Td.areas[ result.Name ]
  if ok {
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

  return result
}

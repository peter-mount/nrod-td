package td

import (
  "log"
  "time"
)

// cleanup removes all berths that have not been updated for 3 hours in an
// effort to remove berths that have not had any removals from or when the
// TD feed fails.
func (s *TD) cleanup() {
  // deadline is 3 hours ago
  deadline := time.Now().Unix() - (3 * 3600 )

  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()

  count := 0
  active := 0
  total := 0

  for _, area := range s.Td.areas {
    for _, berth := range area.berths {
      total++
      if berth.Descr != "" && berth.Timestamp < deadline {
        berth.Descr = ""
        count++
      }
      if berth.Descr != "" {
        active++
      }
    }
  }

  log.Printf(
    "Areas %d Berths %d active %d removed %d",
    len( s.Td.areas ),
    total,
    active,
    count )
}

// reset resets the database by wiping all areas
func (s *TD) reset() {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()

  s.Td.reset = time.Now().Unix()
  for area, _ := range s.Td.areas {
    delete( s.Td.areas, area )
  }
}

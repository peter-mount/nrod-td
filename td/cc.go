package td

type CCMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  To    string  `json:"to"`
  Descr string  `json:"descr"`
}

func (m *CCMessage) handle( s *TD ) {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()
  s.Td.
    update( m.Time ).
    area( m.Area ).
    update( m.Time ).
    berth( m.To ).
    update( m.Time, m.Descr )
}

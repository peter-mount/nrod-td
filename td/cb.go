package td

type CBMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  From  string  `json:"from"`
  Descr string  `json:"descr"`
}

func (m *CBMessage) handle( s *TD ) {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()
  s.Td.
    update( m.Time ).
    area( m.Area ).
    update( m.Time ).
    berth( m.From ).
    update( m.Time, "" )
}

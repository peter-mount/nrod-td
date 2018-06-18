package td

type CTMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  RepTM string  `json:"report_time"`
}

func (m *CTMessage) handle( s *TD ) {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()
  s.Td.update( m.Time ).area( m.Area ).heartBeat = m.RepTM
}

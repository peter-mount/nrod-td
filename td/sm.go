package td

type SMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  Type  string  `json:"msg_type"`
  Addr  string  `json:"address"`
  Data  string  `json:"data"`
}

func (m *SMessage) handle( s *TD ) {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()
  s.Td.update( m.Time ).area( m.Area ).signals[ m.Addr ] = m.Data
}

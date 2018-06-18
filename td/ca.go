package td

type CAMessage struct {
  Time  string  `json:"time"`
  Area  string  `json:"area_id"`
  From  string  `json:"from"`
  To    string  `json:"to"`
  Descr string  `json:"descr"`
}

func (m *CAMessage) handle( s *TD ) {
  s.Td.mutex.Lock()
  defer s.Td.mutex.Unlock()
  var a *TDArea = s.Td.update( m.Time ).area( m.Area ).update( m.Time )
  a.berth( m.From ).update( m.Time, "" )
  a.berth( m.To ).update( m.Time, m.Descr )
}

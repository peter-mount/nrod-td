package td

import (
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "log"
  "path/filepath"
)

func (s *TD) loadConfig() error {
  filename, _ := filepath.Abs( *s.yamlFile )
  log.Println( "Loading config:", filename )

  yml, err := ioutil.ReadFile( filename )
  if err != nil {
    return err
  }

  return yaml.Unmarshal( yml, s )
}


package main

import (
  "gopkg.in/robfig/cron.v2"
//  "log"
)

type CRON struct {
  // The cron service
  service  *cron.Cron
}

func cronInit() {
  settings.Cron.service = cron.New()
}

func cronStart() {
  settings.Cron.service.Start()
}

func cronAdd( s string, f func() ) {
  settings.Cron.service.AddFunc( s, f )
}

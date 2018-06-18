package td

import (
  "flag"
  "fmt"
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/golib/rabbitmq"
  "github.com/peter-mount/golib/rest"
  "sync"
  "time"
)

type TD struct {
  yamlFile     *string

  Debug         bool                    // Debug logging
  //Stats         statistics.Statistics   // Statistics
  Amqp          rabbitmq.RabbitMQ       // RabbitMQ config
  Td            TDFeed                  // TDFeed

  restService  *rest.Server
  Graphite      Graphite
}

func (s *TD) Name() string {
  return "nrod-td"
}

func (s *TD) Init( k *kernel.Kernel ) error {
  s.yamlFile = flag.String( "c", "/config.yaml", "The config file to use" )

  service, err := k.AddService( &rest.Server{} )
  if err != nil {
    return err
  }
  s.restService = (service).(*rest.Server)

  return nil
}

func (s *TD) PostInit() error {

  // Load config
  err := s.loadConfig()
  if err != nil {
    return err
  }

  s.Graphite.rabbitMQ = &s.Amqp

  s.restService.Handle( "/areas", s.tdGetAreas ).Methods( "GET" )
  s.restService.Handle( "/area/{id}", s.tdGetArea ).Methods( "GET" )

  // Old endpoints for compatibility
  s.restService.Handle( "/area", s.tdGetAreas ).Methods( "GET" )
  s.restService.Handle( "/{id}", s.tdGetArea ).Methods( "GET" )

  return nil
}

func (s *TD) Start() error {

  err := s.Amqp.Connect()
  if err != nil {
    return err
  }

  err = s.Graphite.Start()
  if err != nil {
    return err
  }

  if s.Td.Queue == "" {
    return fmt.Errorf( "Queue name is required" )
  }

  if s.Td.RoutingKey == "" {
    return fmt.Errorf( "RoutingKey is required" )
  }

  if s.Td.Exchange == "" {
    s.Td.Exchange = "amq.topic"
  }

  s.Td.areas = make( map[string]*TDArea )
  s.Td.mutex = &sync.Mutex{}

  s.Td.reset = time.Now().Unix()

  return s.tdStart()
}

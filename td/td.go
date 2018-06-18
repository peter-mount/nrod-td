package td

import (
  "flag"
  "fmt"
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/golib/rabbitmq"
  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/golib/statistics"
  "sync"
  "time"
)

type TD struct {
  yamlFile     *string

  Debug         bool                    // Debug logging
  Stats         statistics.Statistics   // Statistics
  Amqp          rabbitmq.RabbitMQ       // RabbitMQ config
  Td            TDFeed                  // TDFeed

  restService  *rest.Server
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

  //s.tdInit()
  return nil
}

func (s *TD) Start() error {

  err := s.Amqp.Connect()
  if err != nil {
    return err
  }

  s.Stats.Configure()

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

  //s.Server.router.HandleFunc( "/area", tdGetAreas ).Methods( "GET" )
  //s.Server.router.HandleFunc( "/{id}", tdGetArea ).Methods( "GET" )

  return s.tdStart()
}

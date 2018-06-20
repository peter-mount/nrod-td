package td

import (
  "fmt"
  "github.com/peter-mount/golib/rabbitmq"
  "github.com/peter-mount/golib/statistics"
  "github.com/streadway/amqp"
  "strings"
)

type Graphite struct {
  rabbitMQ     *rabbitmq.RabbitMQ
  statistics    statistics.Statistics
  Enabled       bool
  Prefix        string
  Exchange      string
  Schedule      string
  channel      *amqp.Channel
}

func (g *Graphite) Name() string {
  return "Graphite"
}

func (g *Graphite) Start() error {
  // Custom statistics engine, capture every 10s so we submit to Graphite at
  // intervals it's expecting
  g.statistics.Log = false

  if g.Schedule != "" {
    g.statistics.Schedule = g.Schedule
  } else {
    g.statistics.Schedule = "0/10 * * * * *"
  }
  g.statistics.Configure()

  if g.Enabled && g.rabbitMQ != nil {

    // Default exchange is "graphite"
    if g.Exchange == "" {
      g.Exchange = "graphite"
    }

    err := g.rabbitMQ.Connect()
    if err != nil {
      return err
    }

    g.channel, err = g.rabbitMQ.NewChannel()
    if err != nil {
      return err
    }

    // We are a statistics Recorder
    g.statistics.Recorder = g
  }

  return nil
}

// PublishStatistic Handles publishing statistics to Graphite over RabbitMQ
func (g *Graphite) PublishStatistic( name string, s *statistics.Statistic ) {
  if strings.HasSuffix( name, "td.all" ) {
    // Value will be the latency
    g.publish( name + ".latency", s.Value, s.Timestamp )
    // Count the number of messages
    g.publish( name + ".count", s.Count, s.Timestamp )

    // Min/Max latency values - don't send if max<min - i.e. no data!
    if s.Max >= s.Min {
      g.publish( name + ".min", s.Min, s.Timestamp )
      g.publish( name + ".max", s.Max, s.Timestamp )
    }
  } else {
    g.publish( name, s.Value, s.Timestamp )
  }
}

func (g *Graphite) publish( name string, val int64, ts int64 ) {
  statName := name
  if g.Prefix != "" {
    statName = g.Prefix + "." + name
  }
  msg := fmt.Sprintf( "%s %d %d", statName, val, ts)

  g.channel.Publish(
    g.Exchange,
    statName,
    false,
    false,
    amqp.Publishing{
      Body: []byte(msg),
  })
}

package output

import (
	"fmt"
	"mon-agent/pkg/check"
	"mon-agent/pkg/tsdb"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func NewPublisher(client *tsdb.Graphite) *Publisher {
	return &Publisher{
		client: client,
	}
}

type Publisher struct {
	client *tsdb.Graphite
}

func (p *Publisher) Publish(resCh chan check.ExecResult) {
	for message := range resCh {

		_, pdata := message.Result()
		rawPoints := strings.Split(pdata, " ")

		var points string
		for _, point := range rawPoints {
			log.Debugf("Publishing perfdata: %s", point)
			ps := strings.Split(point, "=")
			points += fmt.Sprintf("%s.%s %s %d\n", "mon-agent", ps[0], ps[1], time.Now().Unix())
		}

		var err error
		err = p.client.Write(points)
		if err != nil {
			log.Debugf("Error writing datapoint to Graphite: %s", err.Error())
		}
	}
}

package check

import (
	"fmt"
	"mon-agent/pkg/tsdb"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Publish publishes the results of the checks to Graphite
func Publish(resCh chan ExecResult) {
	graphite := tsdb.NewGraphite("localhost", "2003", "mon-agent", "tcp", time.Duration(10*time.Second))
	for message := range resCh {

		var points string

		log.Debugf("Publishing check: %s", message.Name)

		_, pdata := message.Result()

		pre := strings.Split(pdata, " ")

		for _, p := range pre {
			log.Debugf("Publishing perfdata: %s", p)
			ps := strings.Split(p, "=")
			points += fmt.Sprintf("Opsview.FIFA.lta-s1.%s %s %d\n", ps[0], ps[1], time.Now().Unix())
		}

		err := graphite.Write(points)
		if err != nil {
			log.Errorf("Error writing datapoint to Graphite: %s", err.Error())
		}
	}
}

// Read reads the results of the checks
func Read(resCh chan ExecResult) {

	for message := range resCh {
		output, pdata := message.Result()
		if message.Error != nil {
			log.WithFields(log.Fields{
				"check":    message.Name,
				"exectime": message.ExecTime,
				"output":   output,
				"error":    message.Error,
			}).Info(output)
		} else {
			log.WithFields(log.Fields{
				"check":    message.Name,
				"exectime": message.ExecTime,
				"output":   output,
				"perfdata": pdata,
			}).Info(output)

		}
	}
}

// perfdata is the perfdata returned by the check
// we must clean it to make it easier to post to Graphite-line services
//
// TODO: use a strings.NewReplacer() to replace the perfdata characters
// instead of using harcoded options
func perfdata(p string) string {
	log.Debugf("Raw perfdata: %s", p)

	p = strings.Trim(p, " ")
	p = strings.Trim(p, "\n")

	p = strings.Replace(p, "/", "_", -1)
	return p
}

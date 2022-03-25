package check

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"mon-agent/pkg/tsdb"
)

// Read reads the results of the checks
func Read(resCh chan ExecResult) {
	graphite := tsdb.NewGraphite("localhost", "2003", "mon-agent", "tcp", time.Duration(10*time.Second))

	for message := range resCh {
		output, pdata := message.Result()

		if message.Error != nil {
			log.WithFields(log.Fields{
				"check":    message.Name,
				"exectime": message.ExecTime,
				"output":   output,
				"error":    message.Error,
			}).Error(output)
		} else {
			log.WithFields(log.Fields{
				"check":    message.Name,
				"exectime": message.ExecTime,
				"output":   output,
				"perfdata": pdata,
			}).Info(output)

			graphite.Write(pdata)
		}
		// split the perfdata
		if len(pdata) > 0 {
			log.Debugf("perfdata: %s", pdata)
		} else {
			log.Debugf("No perfdata returned by check: %s", message.Name)
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

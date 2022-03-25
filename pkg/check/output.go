package check

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// Read reads the results of the checks
func Read(resCh chan ExecResult) {
	for message := range resCh {
		if message.Error != nil {
			log.Infof("Error executing check: %s", message.Error)
			return
		}

		output, pdata := message.Result()

		log.Info(output)

		// split the perfdata
		if len(pdata) > 0 {
			log.Info("perfdata: ", pdata)
		} else {
			log.Warnf("No perfdata returned by check: %s", message.Name)
		}
	}
}

// perfdata is the perfdata returned by the check
// we must clean it to make it easier to post to Graphite-line services
func perfdata(p string) string {
	log.Debugf("perfdata: %s", p)

	p = strings.Replace(p, " ", "_", -1)
	p = strings.Replace(p, "/", "_", -1)
	return p
}

package check

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

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

package tsdb

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Point struct {
	Name  string
	Value float64
	Time  time.Time
}

func (p *Point) String() string {
	return ""
}

type Graphite struct {
	// Host is the hostname of the Graphite server
	Host string

	// Port is the port of the Graphite server
	Port string

	// Prefix is the prefix to be used for all metrics
	Prefix string

	// Protocol is the protocol to use to connect to the Graphite server
	Protocol string

	// Timeout is the timeout to use when connecting to the Graphite server
	Timeout time.Duration
}

// NewGraphite returns a new Graphite client
func NewGraphite(host string, port string, prefix string, protocol string, timeout time.Duration) *Graphite {
	return &Graphite{
		Host:     host,
		Port:     port,
		Prefix:   prefix,
		Protocol: protocol,
		Timeout:  timeout,
	}
}

// Write writes the given points to Graphite
func (g *Graphite) Write(point string) error {

	if len(point) == 0 {
		return nil
	}

	// connect to graphite server
	conn, err := net.DialTimeout(g.Protocol, net.JoinHostPort(g.Host, g.Port), g.Timeout)
	if err != nil {
		return err
	}

	for _, p := range strings.Split(point, " ") {
		log.Println("Sending point: ", p)

		// Send the points to the Graphite server
		p = strings.Split(p, "=")[1]
		dp := fmt.Sprintf("%s %s %d\n", "mon-agent.check.output", p, time.Now().Unix())
		_, err = conn.Write([]byte(dp))
		if err != nil {
			return err
		}
	}

	return conn.Close()
}

func (g *Graphite) format(point Point) error {
	return nil
}

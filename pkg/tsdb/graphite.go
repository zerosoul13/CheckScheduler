package tsdb

import (
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// NewGraphite returns a new Graphite client
func NewGraphite(host string, port string, protocol string, timeout time.Duration) *Graphite {
	return &Graphite{
		Host:     host,
		Port:     port,
		Protocol: protocol,
		Timeout:  timeout,
	}
}

type Point struct {
	Name      string
	Value     float64
	Timestamp int64
}

func (p Point) String() string {
	return fmt.Sprintf("%s %f %d\n", p.Name, p.Value, p.Timestamp)
}

func NewPoint(name string, value float64, timestamp int64) Point {

	log.Debugf("Creating point: %s %f %d", name, value, timestamp)
	return Point{
		Name:      name,
		Value:     0,
		Timestamp: timestamp,
	}
}

func NewPointFromString(check string, s string) (Point, error) {
	log.Debugf("Creating point from string: %s", s)

	// parse the string
	v := strings.Split(s, "")

	p := Point{
		Name:      check + "." + v[0],
		Value:     0,
		Timestamp: time.Now().Unix(),
	}

	return p, nil
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

// Write writes the given points to Graphite
func (g *Graphite) Write(points string) error {

	// connect to graphite server
	conn, err := net.DialTimeout(g.Protocol, net.JoinHostPort(g.Host, g.Port), g.Timeout)
	if err != nil {
		return err
	}

	// Send the points to the Graphite server
	log.Debugf("Publishing point: %s", points)
	_, err = conn.Write([]byte(points))
	if err != nil {
		return err
	}
	return conn.Close()
}

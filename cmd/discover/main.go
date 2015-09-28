package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/formatters/logstash"
	"github.com/agrarianlabs/router/discover"
	"github.com/creack/ehttp"
)

var (
	defaultPort      = 9090
	defaultDockerURL = "unix:///var/run/docker.sock"
)

// LogstashHook .
type LogstashHook struct {
	conn net.Conn
}

// NewLogstashHook .
func NewLogstashHook(network, addr string) (*LogstashHook, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return &LogstashHook{conn: conn}, nil
}

// Close terminates the socket.
func (h LogstashHook) Close() error {
	return h.conn.Close()
}

// Levels implements the logrus Hook interface.
func (h *LogstashHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Fire implements logrus Hook interface.
func (h *LogstashHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	if err != nil {
		return err
	}
	if _, err := h.conn.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}

var log = discover.NewLogger().
	SetLevel(logrus.DebugLevel).
	SetFormatter(&logstash.Formatter{Type: "discover"})

func main() {
	var (
		listenAddr string
		dockerURL  string
	)

	preHook := func(logstashIP string) {
		println("--->")
		if hook, err := NewLogstashHook("udp", logstashIP+":5000"); err != nil {
			log.Error(err)
		} else {
			log.AddHook("logstash", hook)
		}
	}

	os.Setenv("DISCOVERY_PATH", "/tmp/foo")
	go discover.WatchService(preHook, nil, "d", os.Getenv("DISCOVERY_PATH"), make(chan struct{}))

	if port := os.Getenv("PORT"); port == "" {
		listenAddr = fmt.Sprintf(":%d", defaultPort)
	} else {
		listenAddr = ":" + port
	}
	if url := os.Getenv("DOCKER_URL"); url == "" {
		dockerURL = defaultDockerURL
	} else {
		dockerURL = url
	}

	log.Printf("ready on %s", listenAddr)

	discovery, err := discover.NewDiscovery(dockerURL)
	if err != nil {
		log.Fatal(err)
	}
	logrus.Fatal(http.ListenAndServe(listenAddr, ehttp.HandlerFunc(discovery.LookupHandler)))
}

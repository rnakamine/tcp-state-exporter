package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/procfs"
	flag "github.com/spf13/pflag"
)

var (
	tcpConnections  *prometheus.GaugeVec
	tcpListingPorts *prometheus.GaugeVec
	dynamicLabels   prometheus.Labels
)

func init() {
	dynamicLabels = getDynamicLabels()

	dynamicLabelNames := []string{}
	for k := range dynamicLabels {
		dynamicLabelNames = append(dynamicLabelNames, k)
	}

	tcpConnectionsLabels := append([]string{"state", "remote_address", "remote_port"}, dynamicLabelNames...)
	tcpConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tcp_connections",
			Help: "Current number of TCP connections by state and remote address",
		},
		tcpConnectionsLabels,
	)

	tcpListingPortsLabels := append([]string{"local_address", "local_port"}, dynamicLabelNames...)
	tcpListingPorts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tcp_listening_ports",
			Help: "Current number of TCP listening ports by local address",
		},
		tcpListingPortsLabels,
	)
}

func getDynamicLabels() prometheus.Labels {
	prefix := "EXPORTER_LABEL_"
	labels := make(prometheus.Labels)

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if strings.HasPrefix(pair[0], prefix) {
			key := strings.TrimPrefix(pair[0], prefix)
			labels[key] = pair[1]
		}
	}

	return labels
}

type TcpStateCollector struct {
	fs procfs.FS
}

func (c TcpStateCollector) Describe(ch chan<- *prometheus.Desc) {
	tcpConnections.Describe(ch)
	tcpListingPorts.Describe(ch)
}

func (c TcpStateCollector) Collect(ch chan<- prometheus.Metric) {
	tcpConnections.Reset()
	tcpListingPorts.Reset()

	tcp, err := c.fs.NetTCP()
	if err != nil {
		log.Println(fmt.Errorf("error getting NetTCP stats: %w", err))
	}

	for _, t := range tcp {
		labels := make(prometheus.Labels)

		for k, v := range dynamicLabels {
			labels[k] = v
		}

		state := convertState(int(t.St))
		if state == "LISTEN" {
			labels["local_address"] = t.LocalAddr.String()
			labels["local_port"] = strconv.FormatUint(t.LocalPort, 10)
			tcpListingPorts.With(labels).Inc()
		} else {
			labels["state"] = state
			labels["remote_address"] = t.RemAddr.String()
			labels["remote_port"] = strconv.FormatUint(t.RemPort, 10)
			tcpConnections.With(labels).Inc()
		}
	}

	tcpConnections.Collect(ch)
	tcpListingPorts.Collect(ch)
}

func convertState(state int) string {
	switch state {
	case 1:
		return "ESTABLISHED"
	case 2:
		return "SYN_SENT"
	case 3:
		return "SYN_RECV"
	case 4:
		return "FIN_WAIT1"
	case 5:
		return "FIN_WAIT2"
	case 6:
		return "TIME_WAIT"
	case 7:
		return "CLOSE"
	case 8:
		return "CLOSE_WAIT"
	case 9:
		return "LAST_ACK"
	case 10:
		return "LISTEN"
	case 11:
		return "CLOSING"
	default:
		return "UNKNOWN"
	}
}

func main() {
	var port string
	var help bool

	flag.StringVarP(&port, "port", "p", "9112", "The port number for the metrics server to listen on")
	flag.BoolVarP(&help, "help", "h", false, "Display this help message")
	flag.Parse()

	if help {
		fmt.Println("Usage: your-program [-port port-number] [-help]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	fs, err := procfs.NewFS("/proc")
	if err != nil {
		log.Fatal(err)
	}

	c := TcpStateCollector{fs: fs}
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting HTTP server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

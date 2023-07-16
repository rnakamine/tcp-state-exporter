package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/procfs"
)

var (
	tcpConnectionsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tcp_connections_total",
			Help: "Total number of TCP connections by state and remote address",
		},
		[]string{"state", "remote_address", "remote_port"},
	)
	tcpListingPortsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tcp_listening_ports_total",
			Help: "Total number of TCP listening ports by local address",
		},
		[]string{"local_address", "local_port"},
	)
)

type TcpStateCollector struct {
	fs procfs.FS
}

func (c TcpStateCollector) Describe(ch chan<- *prometheus.Desc) {
	tcpConnectionsTotal.Describe(ch)
	tcpListingPortsTotal.Describe(ch)
}

func (c TcpStateCollector) Collect(ch chan<- prometheus.Metric) {
	tcpConnectionsTotal.Reset()
	tcpListingPortsTotal.Reset()

	tcp, err := c.fs.NetTCP()
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tcp {
		state := convertState(int(t.St))
		localPort := strconv.FormatUint(t.LocalPort, 10)
		remPort := strconv.FormatUint(t.RemPort, 10)

		if state == "LISTEN" {
			tcpListingPortsTotal.WithLabelValues(t.LocalAddr.String(), localPort).Inc()
		} else {
			tcpConnectionsTotal.WithLabelValues(state, t.RemAddr.String(), remPort).Inc()
		}
	}

	tcpConnectionsTotal.Collect(ch)
	tcpListingPortsTotal.Collect(ch)
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
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		log.Fatal(err)
	}

	c := TcpStateCollector{fs: fs}
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}

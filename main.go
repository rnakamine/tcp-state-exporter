package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/procfs"
)

type TcpCoonectionCollector struct {
	tcpConnectionsTotal  *prometheus.GaugeVec
	tcpListingPortsTotal *prometheus.GaugeVec
	fs                   procfs.FS
}

func NewTcpCoonectionCollector() *TcpCoonectionCollector {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		log.Fatal(err)
	}

	return &TcpCoonectionCollector{
		tcpConnectionsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tcp_connections_total",
				Help: "Total number of TCP connections by state and remote address",
			},
			[]string{"state", "remote_address", "remote_port"},
		),
		tcpListingPortsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tcp_listening_ports_total",
				Help: "Total number of TCP listening ports by local address",
			},
			[]string{"local_address", "local_port"},
		),
		fs: fs,
	}
}

func (c *TcpCoonectionCollector) Describe(ch chan<- *prometheus.Desc) {
	c.tcpConnectionsTotal.Describe(ch)
	c.tcpListingPortsTotal.Describe(ch)
}

func (c *TcpCoonectionCollector) Collect(ch chan<- prometheus.Metric) {
	c.tcpConnectionsTotal.Reset()
	c.tcpListingPortsTotal.Reset()

	tcp, err := c.fs.NetTCP()
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tcp {
		state := convertState(int(t.St))
		localPort := strconv.FormatUint(t.LocalPort, 10)
		remPort := strconv.FormatUint(t.RemPort, 10)

		if state == "LISTEN" {
			c.tcpListingPortsTotal.WithLabelValues(t.LocalAddr.String(), localPort).Inc()
		} else {
			c.tcpConnectionsTotal.WithLabelValues(state, t.RemAddr.String(), remPort).Inc()
		}
	}

	c.tcpListingPortsTotal.Collect(ch)
	c.tcpConnectionsTotal.Collect(ch)
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
	collecter := NewTcpCoonectionCollector()
	prometheus.MustRegister(collecter)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}

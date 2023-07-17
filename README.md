# TCP State Exporter

TCP State Exporter is a simple service written in Go that exposes metrics about the state of TCP connections on a machine. It leverages Prometheus's client library to expose these metrics at a `/metrics` endpoint, which can then be scraped by Prometheus.

## Usage

```sh
./tcpstateexporter --port <port-number>
```

Use the `-p` or `--port` flag to specify the port number the metrics server should listen on. The default is `9112`.

The `-h` or `--help` flag can be used to display a help message and exits.

## Example Output

When you navigate to `http://localhost:<port-number>/metrics`, you will see output similar to the following:

```
# HELP tcp_connections Current number of TCP connections by state and remote address
# TYPE tcp_connections gauge
tcp_connections{remote_address="192.0.2.1", remote_port="12345", state="ESTABLISHED"} 1
tcp_connections{remote_address="192.0.2.2", remote_port="23456", state="SYN_SENT"} 1
tcp_connections{remote_address="192.0.2.3", remote_port="34567", state="TIME_WAIT"} 1

# HELP tcp_listening_ports Current number of TCP listening ports by local address
# TYPE tcp_listening_ports gauge
tcp_listening_ports{local_address="0.0.0.0", local_port="22"} 1
tcp_listening_ports{local_address="0.0.0.0", local_port="80"} 1
```

HTTP Pinger
===========

A small utility application to make a GET request to a configurable HTTP URL,
measure the latency and push it to
[Carbon](https://github.com/graphite-project/carbon). The utility is useful to
recreate rare network issues or simply measure latencies.

Made by [Jens Rantil](https://jensrantil.github.io).

Usage
-----
The application tries to have sane defaults and could be started without any
additional parameters. Here's a list of all CLI parameters:

    $ ./http-pinger -help
    Usage of ./http-pinger:
      -carbon-host="localhost": Carbon host/IP.
      -carbon-interval=1m0s: Interval to write to Carbon.
      -carbon-namespace="http-pinger": Where the Carbon data should be stored in Graphite.
      -carbon-port=2003: Carbon port.
      -http-interval=10s: HTTP test interval.
      -http-timeout=30s: HTTP request socket timeout.
      -url="http://localhost/": The URL to issue the GET to.

Building
--------
Execute

    $ go get
    $ go build

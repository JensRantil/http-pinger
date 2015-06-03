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

Sample of metric submitted to Carbon/Graphite:

    http-pinger.min 0 1433333074
    http-pinger.max 2 1433333074
    http-pinger.median 0 1433333074
    http-pinger.5p 0 1433333074
    http-pinger.25p 0 1433333074
    http-pinger.95p 2 1433333074
    http-pinger.75p 0 1433333074
    http-pinger.timeouts 0 1433333074
    http-pinger.count 11 1433333074
    http-pinger.errors 11 1433333074

Building
--------
Execute

    $ go get
    $ go build

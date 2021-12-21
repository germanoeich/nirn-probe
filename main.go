package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	ErrorCounter prometheus.Counter
	BannedGauge prometheus.Gauge
	UpGauge prometheus.Gauge
)

func queryDiscord() (int, int, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v9/gateway", nil)
	if err != nil {
		return 0, 0, err
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return 0, 0, err
	}

	retryAfter := int64(0)
	retryAfterStr := resp.Header.Get("retry-after")
	if retryAfterStr != "" {
		retryAfter, _ = strconv.ParseInt(retryAfterStr, 10, 32)
	}

	return resp.StatusCode, int(retryAfter), nil
}

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName, _ = os.Hostname()
	}

	logrus.Info("Using node=" + nodeName)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8100"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	lvl, err := logrus.ParseLevel(logLevel)
	logrus.SetLevel(lvl)
	logrus.Info("Log level set to " + logLevel)

	// Disable go_exporter default metrics
	r := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = r
	prometheus.DefaultGatherer = r

	ErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nirn_probe_error",
		Help: "The total number of errors when processing requests",
		ConstLabels: map[string]string{
			"node": nodeName,
		},
	})

	BannedGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "nirn_probe_banned",
		Help:        "Gauge with banned state - 0 if unbanned, 1 if banned",
		ConstLabels: map[string]string{
			"node": nodeName,
		},
	})

	UpGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "nirn_probe_up",
		Help:        "Gauge with uptime state - always 1 if application is alive",
		ConstLabels: map[string]string{
			"node": nodeName,
		},
	})
	UpGauge.Set(1)

	check()
	go tick()

	http.Handle("/metrics", promhttp.Handler())
	logrus.Info("Starting metrics server on port " + port)
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
}

func check() {
	logrus.Debug("Querying Discord")
	status, retryAfter, err := queryDiscord()

	logrus.WithFields(logrus.Fields{
		"status": status,
		"retryAfter": retryAfter,
		"err": err,
	}).Debug("Discord replied")

	if err != nil {
		logrus.Error(err)
		ErrorCounter.Inc()
	}

	if status == 429 {
		BannedGauge.Set(1)
	} else {
		BannedGauge.Set(0)
	}
}

func tick() {
	t := time.NewTicker(1 * time.Minute)

	for range t.C {
		check()
	}
}

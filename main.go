package main

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/BaritoLog/barito-blackbox-exporter/config"
	"github.com/BaritoLog/barito-blackbox-exporter/exporter"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.SetLevel(log.DebugLevel)

	cfg := config.NewConfig()
	mR := o11y.NewMetricRecorder()

	appGroups := []map[string]string{
		{"code": "hoke", "secret": ""},
	}
	for _, appGroup := range appGroups {
		go createPushAgent(appGroup["code"], appGroup["secret"], cfg, mR).Run()
	}

	http.Handle("/metrics", promhttp.HandlerFor(
		mR.GetRegistry(),
		promhttp.HandlerOpts{EnableOpenMetrics: true},
	))
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func createPushAgent(code, secret string, cfg *config.Config, mR o11y.MetricRecorder) *exporter.PushAgent {
	return exporter.NewPushAgent(code, secret, context.Background(), cfg, mR)
}

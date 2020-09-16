package main

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/BaritoLog/barito-blackbox-exporter/appgroup"
	"github.com/BaritoLog/barito-blackbox-exporter/config"
	"github.com/BaritoLog/barito-blackbox-exporter/exporter"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.SetLevel(log.DebugLevel)

	cfg := config.NewConfig()
	mR := o11y.NewMetricRecorder()

	appGroups := getClusterAndSecret()

	for _, mapAppGroup := range appGroups {
		aG := appgroup.NewAppGroup(mapAppGroup["code"], mapAppGroup["secret"], cfg)
		go createPushAgent(mapAppGroup["code"], mapAppGroup["secret"], cfg, mR).Run()
		go createESProbeAgent(aG, cfg, mR).Run()
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

func createESProbeAgent(appGroup appgroup.AppGroup, cfg *config.Config, mR o11y.MetricRecorder) *exporter.ESProbeAgent {
	return exporter.NewESProbeAgent(appGroup, context.Background(), cfg, mR)
}

func getClusterAndSecret() []map[string]string {
	result := []map[string]string{}

	file, err := os.Open("./secret_sample")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), " ")
		result = append(result, map[string]string{"code": s[0], "secret": s[1]})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Infof("Found %d appgroup", len(result))
	return result
}

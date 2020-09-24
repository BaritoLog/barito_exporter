package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
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

	mapAppGroups := getClusterAndSecret()
	appGroups := []appgroup.AppGroup{}

	for _, v := range mapAppGroups {
		appGroups = append(appGroups, appgroup.NewAppGroup(v["code"], v["secret"], cfg))
	}

	for _, aG := range appGroups {
		go createPushAgent(aG, cfg, mR).Run()
		go createESProbeAgent(aG, cfg, mR).Run()
		go createKibanaProbeAgent(aG, cfg, mR).Run()
	}

	// todo: disable for now, because after deleting the topic, consumer must be restarted
	//go deleteProberKafkaTopic(appGroups, cfg)

	http.Handle("/metrics", promhttp.HandlerFor(
		mR.GetRegistry(),
		promhttp.HandlerOpts{EnableOpenMetrics: true},
	))
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func createPushAgent(appGroup appgroup.AppGroup, cfg *config.Config, mR o11y.MetricRecorder) *exporter.PushAgent {
	return exporter.NewPushAgent(appGroup.GetClusterName(), appGroup.GetSecret(), context.Background(), cfg, mR)
}

func createESProbeAgent(appGroup appgroup.AppGroup, cfg *config.Config, mR o11y.MetricRecorder) *exporter.ESProbeAgent {
	return exporter.NewESProbeAgent(appGroup, context.Background(), cfg, mR)
}

func createKibanaProbeAgent(appGroup appgroup.AppGroup, cfg *config.Config, mR o11y.MetricRecorder) *exporter.KibanaProbeAgent {
	return exporter.NewKibanaProbeAgent(appGroup, context.Background(), cfg, mR)
}

func getClusterAndSecret() []map[string]string {
	result := []map[string]string{}

	file, err := os.Open("./secrets_sample")
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

func deleteProberKafkaTopic(appGroups []appgroup.AppGroup, cfg *config.Config) {
	for {
		for _, aG := range appGroups {
			aG.RefreshMetadata()
			listKafka, err := aG.GetListKafka()
			if err != nil {
				log.Errorf("Failed to GetListKafka on app_group: %q, err: %v", aG.GetClusterName(), err)
				continue
			}

			saramaCfg := sarama.NewConfig()
			saramaCfg.Version = sarama.V2_5_0_0
			clusterAdmin, err := sarama.NewClusterAdmin(listKafka, saramaCfg)
			if err != nil {
				log.Errorf("Failed to create sarama client, app_group: %q, err: %v", aG.GetClusterName(), err)
				continue
			}

			topicName := fmt.Sprintf("%s-%s_pb", cfg.ProduceAppPrefix, aG.GetClusterName())
			err = clusterAdmin.DeleteTopic(topicName)
			if err != nil {
				log.Errorf("Failed to delete topic, app_group: %q, err: %v", aG.GetClusterName(), err)
				continue
			}
		}
		time.Sleep(cfg.DeleteTopicInterval)
	}
}

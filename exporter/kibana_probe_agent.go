package exporter

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/appgroup"
	"github.com/BaritoLog/barito-blackbox-exporter/config"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	log "github.com/sirupsen/logrus"
)

type KibanaProbeAgent struct {
	appGroup       appgroup.AppGroup
	probePath      string
	interval       time.Duration
	requestTimeout time.Duration
	metricRecorder o11y.MetricRecorder
	ctx            context.Context
}

func NewKibanaProbeAgent(appGroup appgroup.AppGroup, ctx context.Context, cfg *config.Config, mR o11y.MetricRecorder) *KibanaProbeAgent {
	path := fmt.Sprintf("/%s/api/index_management/indices", appGroup.GetClusterName())
	return &KibanaProbeAgent{
		appGroup:       appGroup,
		probePath:      path,
		interval:       cfg.KibanaProbeInterval,
		requestTimeout: cfg.KibanaProbeTimeout,
		metricRecorder: mR,
		ctx:            ctx,
	}
}

func (e *KibanaProbeAgent) Run() {
	for {
		select {
		case <-e.ctx.Done():
			log.Println("Exit")
			return
		default:
			err := e.tick()
			if err != nil {
				log.Errorf("Failed to probe kibana, appGroup: %q, error: %v", e.appGroup.GetClusterName(), err)
			}
			time.Sleep(e.interval)
		}
	}
}

func (e *KibanaProbeAgent) tick() error {
	err := e.appGroup.RefreshMetadata()
	if err != nil {
		e.metricRecorder.IncreaseProbeKibanaFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_KIBANA_FAILED_FETCH_METADATA)
		return err
	}

	kibanaURL, err := e.appGroup.GetKibanaHost()
	if err != nil {
		e.metricRecorder.IncreaseProbeKibanaFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_KIBANA_FAILED_GET_KIBANA_FROM_CONSUL)
		return err
	}
	if len(kibanaURL) == 0 {
		e.metricRecorder.IncreaseProbeKibanaFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_KIBANA_NO_KIBANA_FOUND)
		return err
	}

	url := kibanaURL + e.probePath
	_, err = e.doRequest(url)
	if err != nil {
		log.Debugf("Failed to hit Kibana, appgroup: %q, es: %q", e.appGroup.GetClusterName(), url)
		e.metricRecorder.IncreaseProbeKibanaFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_KIBANA_REQUEST_FAILED)
		return err
	}

	e.metricRecorder.IncreaseProbeKibanaSuccess(e.appGroup.GetClusterName())
	return nil
}

func (e *KibanaProbeAgent) doRequest(url string) ([]byte, error) {
	log.Debugf("Do Kibana requests, appGroup: %q, URL: %q", e.appGroup.GetClusterName(), url)

	var c = &http.Client{
		Timeout: e.requestTimeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte(""), errors.New("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return []byte(""), err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Got response status %d", resp.StatusCode)
		log.Debugf("Kibana requests got status: %d, appGroup: %q, URL: %q", resp.StatusCode, e.appGroup.GetClusterName(), url)
		return []byte(""), err
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

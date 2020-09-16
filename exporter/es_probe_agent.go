package exporter

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/appgroup"
	"github.com/BaritoLog/barito-blackbox-exporter/config"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	"github.com/Jeffail/gabs/v2"
	log "github.com/sirupsen/logrus"
)

type ESProbeAgent struct {
	appGroup       appgroup.AppGroup
	appPrefix      string
	esTimeField    string
	interval       time.Duration
	requestTimeout time.Duration
	metricRecorder o11y.MetricRecorder
	ctx            context.Context
}

func NewESProbeAgent(appGroup appgroup.AppGroup, ctx context.Context, cfg *config.Config, mR o11y.MetricRecorder) *ESProbeAgent {
	return &ESProbeAgent{
		appGroup:       appGroup,
		appPrefix:      cfg.ProduceAppPrefix,
		esTimeField:    cfg.ProduceTimeField,
		interval:       cfg.ESProbeInterval,
		requestTimeout: cfg.ESProbeTimeout,
		metricRecorder: mR,
		ctx:            ctx,
	}
}

func (e *ESProbeAgent) Run() {
	for {
		select {
		case <-e.ctx.Done():
			log.Println("Exit")
			return
		default:
			err := e.tick()
			if err != nil {
				log.Errorf("Failed to probe ES, appGroup: %q, error: %v", e.appGroup.GetClusterName(), err)
			}
			time.Sleep(e.interval)
		}
	}
}

func (e *ESProbeAgent) tick() error {
	err := e.appGroup.RefreshMetadata()
	if err != nil {
		e.metricRecorder.IncreaseProbeElasticSearchFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_ELASTICSEARCH_FAILED_FETCH_METADATA)
		return err
	}

	// get ES Url
	esUrls, err := e.appGroup.GetListES()
	if err != nil {
		e.metricRecorder.IncreaseProbeElasticSearchFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_ELASTICSEARCH_FAILED_GET_LIST_FROM_CONSUL)
		return err
	}

	if len(esUrls) == 0 {
		e.metricRecorder.IncreaseProbeElasticSearchFailed(e.appGroup.GetClusterName(),
			o11y.REASON_PROBE_ELASTICSEARCH_NO_ELASTICSEARCH_FOUND)
		return err
	}

	var dataTime int64
	for _, esUrl := range esUrls {
		body, err := e.doRequest(esUrl)
		if err != nil {
			log.Debugf("Failed to hit ES, appgroup: %q, es: %q", e.appGroup.GetClusterName(), esUrl)
			e.metricRecorder.IncreaseProbeElasticSearchFailed(e.appGroup.GetClusterName(),
				o11y.REASON_PROBE_ELASTICSEARCH_REQUEST_FAILED)
			continue
		}
		dataTime, err = e.parseESBody(body)
		if err != nil {
			log.Debugf("Failed to parse ES response, got error: %v", err)
			e.metricRecorder.IncreaseProbeElasticSearchFailed(e.appGroup.GetClusterName(),
				o11y.REASON_PROBE_ELASTICSEARCH_GET_DATA_FAILED)
			continue
		}
		break
	}

	if dataTime != 0 {
		delay := math.Floor(float64(((time.Now().UnixNano() / 1000000) - dataTime) / 1000))
		e.metricRecorder.IncreaseProbeElasticSearchSuccess(e.appGroup.GetClusterName())
		e.metricRecorder.SetProbeElasticsearchDelay(e.appGroup.GetClusterName(), delay)
	}
	return nil
}

func (e *ESProbeAgent) parseESBody(body []byte) (int64, error) {
	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		log.Debugf("Failed to parse json, got error: %v", err)
		return 0, err
	}
	value, ok := jsonParsed.Search("hits", "hits", "0", "_source", e.esTimeField).Data().(float64)
	if !ok {
		return 0, errors.New("Can't find value")
	}
	return int64(value), nil
}

func (e *ESProbeAgent) doRequest(esUrl string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s-%s*/_search?sort=%s:desc&size=1", esUrl, e.appPrefix, e.appGroup.GetClusterName(), e.esTimeField)
	log.Debugf("Do ES requests, appGroup: %q, URL: %q", e.appGroup.GetClusterName(), url)

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
		log.Debugf("ES requests got status: %d, appGroup: %q, URL: %q", resp.StatusCode, e.appGroup.GetClusterName(), url)
		return []byte(""), err
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

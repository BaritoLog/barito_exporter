package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/config"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	log "github.com/sirupsen/logrus"
)

type PushAgent struct {
	appGroup       string
	secretKey      string
	appPrefix      string
	produceURL     string
	interval       time.Duration
	requestTimeout time.Duration
	ctx            context.Context
	metricRecorder o11y.MetricRecorder
}

func NewPushAgent(code, secret string, ctx context.Context, cfg *config.Config, mR o11y.MetricRecorder) *PushAgent {
	return &PushAgent{
		appGroup:       code,
		secretKey:      secret,
		appPrefix:      cfg.ProduceAppPrefix,
		produceURL:     cfg.ProduceURL,
		interval:       cfg.ProduceInterval,
		requestTimeout: cfg.ProduceTimeout,
		ctx:            ctx,
		metricRecorder: mR,
	}
}

func (p *PushAgent) Run() {
	for {
		select {
		case <-p.ctx.Done():
			log.Println("Exit")
			return
		default:
			err := p.doRequest()
			if err == nil {
				log.Debugf("Requests success, appGroup: %q, appPrefix: %q, URL: %q", p.appGroup, p.appPrefix, p.produceURL)
				p.metricRecorder.IncreasePushLogSuccess(p.appGroup)
			} else {
				log.Debugf("Requests failed, appGroup: %q, appPrefix: %q, URL: %q", p.appGroup, p.appPrefix, p.produceURL)
				p.metricRecorder.IncreasePushLogFailed(p.appGroup)
			}
			time.Sleep(p.interval)
		}
	}
}

func (p *PushAgent) doRequest() error {
	log.Debugf("Do requests, appGroup: %q, appPrefix: %q, URL: %q", p.appGroup, p.appPrefix, p.produceURL)
	var c = &http.Client{
		Timeout: p.requestTimeout,
	}

	body := fmt.Sprintf(`{"items": [{"barito_trace_time": "%d"}] }`, time.Now().UnixNano()/1000000)
	req, err := http.NewRequest("POST", p.produceURL, strings.NewReader(body))
	if err != nil {
		return errors.New("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Group-Secret", p.secretKey)
	req.Header.Set("X-App-Name", p.appPrefix+"-"+p.appGroup)
	resp, err := c.Do(req)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Got response status %d", resp.StatusCode)
		log.Debugf("Requests got status: %d, appGroup: %q, appPrefix: %q, URL: %q", resp.StatusCode, p.appGroup, p.appPrefix, p.produceURL)
	}

	return err
}

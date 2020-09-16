package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/mock"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	"github.com/golang/mock/gomock"
)

func TestESProbeAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var pathCalled string
	var queryString url.Values
	esSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		pathCalled = r.URL.Path
		queryString = r.URL.Query()
		body := fmt.Sprintf(`{"hits": {"hits": [{"_source": { "barito_trace_time": %d}}]}}`, (time.Now().UnixNano()/1000000)-1001)
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(body))
	}))

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(2)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(2)
	ag.EXPECT().GetListES().Return([]string{esSrv.URL}, nil).MinTimes(2)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeElasticSearchSuccess("lama").MinTimes(1)
	// expect delay 1 second
	mr.EXPECT().SetProbeElasticsearchDelay("lama", float64(1)).MinTimes(1)

	agent := ESProbeAgent{
		appGroup:       ag,
		appPrefix:      "barito-log-probe",
		esTimeField:    "barito_trace_time",
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}

	agent.Run()

	expectedESPath := fmt.Sprintf("/%s-%s*/_search", agent.appPrefix, agent.appGroup.GetClusterName())
	if pathCalled != expectedESPath {
		t.Errorf("Should call ES at path: %q, got: %q", expectedESPath, pathCalled)
	}

	expectedQueryString := url.Values{
		"sort": []string{agent.esTimeField + ":desc"},
		"size": []string{"1"},
	}
	if !reflect.DeepEqual(queryString, expectedQueryString) {
		t.Errorf("Should call ES at with query string: %+v, got: %+v", expectedQueryString, queryString)
	}
}

func TestESProbeAgent_failed_noES(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(1)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(1)
	ag.EXPECT().GetListES().MinTimes(1).Return([]string{}, nil)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeElasticSearchFailed("lama", o11y.REASON_PROBE_ELASTICSEARCH_NO_ELASTICSEARCH_FOUND).MinTimes(1)

	agent := ESProbeAgent{
		appGroup:       ag,
		appPrefix:      "barito-log-probe",
		esTimeField:    "barito_trace_time",
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}

	agent.Run()
}
func TestESProbeAgent_failed_errorGetListES(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(1)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(1)
	ag.EXPECT().GetListES().MinTimes(1).Return([]string{}, errors.New("err"))

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeElasticSearchFailed("lama", o11y.REASON_PROBE_ELASTICSEARCH_FAILED_GET_LIST_FROM_CONSUL).MinTimes(1)

	agent := ESProbeAgent{
		appGroup:       ag,
		appPrefix:      "barito-log-probe",
		esTimeField:    "barito_trace_time",
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}

	agent.Run()
}
func TestESProbeAgent_failed_esTimeout(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	esSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(1)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(1)
	ag.EXPECT().GetListES().MinTimes(1).Return([]string{esSrv.URL}, nil)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeElasticSearchFailed("lama", o11y.REASON_PROBE_ELASTICSEARCH_REQUEST_FAILED).MinTimes(1)

	agent := ESProbeAgent{
		appGroup:       ag,
		appPrefix:      "barito-log-probe",
		esTimeField:    "barito_trace_time",
		interval:       1 * time.Second,
		requestTimeout: 1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}
	agent.Run()
}
func TestESProbeAgent_failed_invalidData(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	esSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf(`{"hahaha": {"hits": [{"_source": { "barito_trace_time": %d}}]}}`, (time.Now().UnixNano()/1000000)-1001)
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(body))
	}))

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(1)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(1)
	ag.EXPECT().GetListES().MinTimes(1).Return([]string{esSrv.URL}, nil)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeElasticSearchFailed("lama", o11y.REASON_PROBE_ELASTICSEARCH_GET_DATA_FAILED).MinTimes(1)

	agent := ESProbeAgent{
		appGroup:       ag,
		appPrefix:      "barito-log-probe",
		esTimeField:    "barito_trace_time",
		interval:       1 * time.Second,
		requestTimeout: 1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}
	agent.Run()
}

func TestParseBody(t *testing.T) {
	payload := `
	{
		"took": 4,
		"timed_out": false,
		"_shards": {
		  "total": 3,
		  "successful": 3,
		  "skipped": 0,
		  "failed": 0
		},
		"hits": {
		  "total": 4,
		  "max_score": null,
		  "hits": [
			{
			  "_index": "barito-prober-hoke-2020.09.11",
			  "_type": "_doc",
			  "_id": "Cswnf3QB64TNSvn13KoH",
			  "_score": null,
			  "_source": {
				"@timestamp": "2020-09-11T21:52:32Z",
				"barito_trace_time": 99
			  },
			  "sort": [
				99
			  ]
			}
		  ]
		}
	  }
	`

	agent := ESProbeAgent{
		esTimeField: "barito_trace_time",
	}
	result, err := agent.parseESBody([]byte(payload))
	if err != nil {
		t.Fatalf("Parse valid body should not error, got: %v", err)
	}

	if 99 != result {
		t.Fatalf("Parse valid body should return: %d, got: %d", 99, result)
	}
}

package exporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/mock"
	"github.com/BaritoLog/barito-blackbox-exporter/o11y"
	"github.com/golang/mock/gomock"
)

func TestKibanaProbeAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var pathCalled string
	esSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		pathCalled = r.URL.Path
	}))

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(2)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(2)
	ag.EXPECT().GetKibanaHost().Return(esSrv.URL, nil).MinTimes(2)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeKibanaSuccess("lama").MinTimes(1)

	agent := KibanaProbeAgent{
		appGroup:       ag,
		probePath:      "/lama/api/index_management/indices",
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}

	agent.Run()

	expectedESPath := agent.probePath
	if pathCalled != expectedESPath {
		t.Errorf("Should call kibana at path: %q, got: %q", expectedESPath, pathCalled)
	}
}

func TestKibanaProbeAgent_kibanaError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	esSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
	}))

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(2)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(2)
	ag.EXPECT().GetKibanaHost().Return(esSrv.URL, nil).MinTimes(2)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeKibanaFailed("lama", o11y.REASON_PROBE_KIBANA_REQUEST_FAILED).MinTimes(1)

	agent := KibanaProbeAgent{
		appGroup:       ag,
		probePath:      "/lama/api/index_management/indices",
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}
	agent.Run()
}

func TestKibanaProbeAgent_kibanaTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	esSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))

	ag := mock.NewMockAppGroup(ctrl)
	ag.EXPECT().RefreshMetadata().MinTimes(1)
	ag.EXPECT().GetClusterName().Return("lama").MinTimes(1)
	ag.EXPECT().GetKibanaHost().Return(esSrv.URL, nil).MinTimes(1)

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreaseProbeKibanaFailed("lama", o11y.REASON_PROBE_KIBANA_REQUEST_FAILED).MinTimes(1)

	agent := KibanaProbeAgent{
		appGroup:       ag,
		probePath:      "/lama/api/index_management/indices",
		interval:       1 * time.Second,
		requestTimeout: 1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}
	agent.Run()
}

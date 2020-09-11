package exporter

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/mock"
	"github.com/golang/mock/gomock"
)

type LogBody struct {
	Items []map[string]string `json: items`
}

func TestPushAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedHeaders := map[string]string{
		"Content-Type":       "application/json",
		"X-App-Group-Secret": "ABC123",
		"X-App-Name":         "barito-log-probe-lama",
	}

	timesCalled := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timesCalled++
		for k, v := range expectedHeaders {
			if r.Header.Get(k) != v {
				t.Errorf("Request should have header %q with value: %q, got: %q ", k, v, r.Header.Get(k))
			}
		}
		var payload LogBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Request body should be a valid json, got: %q", string(body))
		}

		json.Unmarshal(body, &payload)
		if len(payload.Items) == 0 {
			t.Fatal("Request body should have `item`")
		}
		if _, ok := payload.Items[0]["barito_trace_time"]; !ok {
			t.Errorf("Should have `barito_trace_time` field on body request, got: %+v", payload.Items)
		}
		w.WriteHeader(200)
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreasePushLogSuccess("lama").Times(2)

	agent := PushAgent{
		appGroup:       "lama",
		secretKey:      "ABC123",
		appPrefix:      "barito-log-probe",
		produceURL:     srv.URL,
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}
	defer srv.Close()

	agent.Run()

	expectedTimesCalled := 2
	if timesCalled != expectedTimesCalled {
		t.Errorf("agent.Run() on 1s interval & 2s should make %d request to produce_url, got: %d", expectedTimesCalled, timesCalled)
	}
}

func TestPushAgent_non200ShouldMarkedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timesCalled := 0
	status := []int{502, 404}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status[timesCalled])
		timesCalled++
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreasePushLogFailed("lama").Times(2)

	agent := PushAgent{
		appGroup:       "lama",
		secretKey:      "ABC123",
		appPrefix:      "barito-log-probe",
		produceURL:     srv.URL,
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
	}
	defer srv.Close()

	agent.Run()
}

func TestPushAgent_TimeoutShouldMarkedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	mr := mock.NewMockMetricRecorder(ctrl)
	mr.EXPECT().IncreasePushLogFailed("lama").Times(2)

	agent := PushAgent{
		appGroup:       "lama",
		secretKey:      "ABC123",
		appPrefix:      "barito-log-probe",
		produceURL:     srv.URL,
		interval:       1 * time.Second,
		metricRecorder: mr,
		ctx:            ctx,
		requestTimeout: 5 * time.Second,
	}
	defer srv.Close()

	agent.Run()
}

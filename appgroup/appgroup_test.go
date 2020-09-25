package appgroup

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/BaritoLog/barito-blackbox-exporter/config"
)

func TestRefreshMetadata(t *testing.T) {
	var pathCalled string
	var queryCalled url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathCalled = r.URL.Path
		queryCalled = r.URL.Query()

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := `
		{
			"name": "SomeAppgroup",
			"consul_hosts": [ "one", "two", "three" ],
			"meta": {
			  "service_names": { "elasticsearch": "elasticsearch" } }
		}`
		w.Write([]byte(resp))
	}))

	aG := appGroup{
		clusterName:       "lama",
		baritoMarketHost:  srv.URL,
		baritoMarketToken: "ABC12345",
	}
	expectedAppGroup := appGroup{
		clusterName:        "lama",
		name:               "SomeAppgroup",
		consulHosts:        []string{"one", "two", "three"},
		consulServiceNames: map[string]string{"elasticsearch": "elasticsearch"},
		baritoMarketHost:   srv.URL,
		baritoMarketToken:  "ABC12345",
	}

	err := aG.RefreshMetadata()
	if err != nil {
		t.Errorf("Should not return error, got: %v", err)
	}

	expectedQuery := url.Values{
		"cluster_name": []string{aG.clusterName},
		"access_token": []string{aG.baritoMarketToken},
	}

	expectedPathCalled := "/api/v2/profile_by_cluster_name"
	if pathCalled != expectedPathCalled {
		t.Errorf("Should called barito market at path:\n%q\ngot:\n%q", expectedPathCalled, pathCalled)
	}

	if !reflect.DeepEqual(queryCalled, expectedQuery) {
		t.Errorf("Should called barito market at with query:\n%v\ngot:\n%v", expectedQuery, queryCalled)
	}

	if !reflect.DeepEqual(aG, expectedAppGroup) {
		t.Errorf("Failed to parse metadata, want:\n%+v, got:\n%+v", expectedAppGroup, aG)
	}
}

func TestGetListES(t *testing.T) {
	var pathCalled string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathCalled = r.URL.Path

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := `[
		{ "Service": { "Address": "172.0.0.1", "Port": 9200 } },
		{ "Service": { "Address": "172.0.0.2", "Port": 9200 } }
		]`
		w.Write([]byte(resp))
	}))

	aG := appGroup{
		clusterName:        "lama",
		consulHosts:        []string{srv.URL},
		consulServiceNames: map[string]string{"elasticsearch": "elasticsearch"},
	}

	listES, err := aG.GetListES()
	if err != nil {
		t.Fatalf("Should not return error, got: %v", err)
	}

	expectedListES := []string{"http://172.0.0.1:9200", "http://172.0.0.2:9200"}
	if !reflect.DeepEqual(listES, expectedListES) {
		t.Errorf("Invalid List ES, got:\n%v,\nwant:\n%v\n", listES, expectedListES)
	}

	expectedPathCalled := fmt.Sprintf("/v1/health/service/%s", aG.consulServiceNames["elasticsearch"])
	if pathCalled != expectedPathCalled {
		t.Errorf("Invalid consul path called, got:\n%q,\nwant:\n%q\n", pathCalled, expectedPathCalled)
	}
}

func TestGetKibanaHost(t *testing.T) {
	var pathCalled string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathCalled = r.URL.Path

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := `[
		{ "Service": { "Address": "172.0.0.1", "Port": 5601 } }
		]`
		w.Write([]byte(resp))
	}))

	aG := appGroup{
		clusterName:        "lama",
		consulHosts:        []string{srv.URL},
		consulServiceNames: map[string]string{"kibana": "kibana"},
	}

	kibanaHost, err := aG.GetKibanaHost()
	if err != nil {
		t.Fatalf("Should not return error, got: %v", err)
	}

	expectedKibanaHost := "http://172.0.0.1:5601"
	if !reflect.DeepEqual(kibanaHost, expectedKibanaHost) {
		t.Errorf("Invalid Kibana Host, got:\n%v,\nwant:\n%v\n", kibanaHost, expectedKibanaHost)
	}

	expectedPathCalled := fmt.Sprintf("/v1/health/service/%s", aG.consulServiceNames["kibana"])
	if pathCalled != expectedPathCalled {
		t.Errorf("Invalid consul path called, got:\n%q,\nwant:\n%q\n", pathCalled, expectedPathCalled)
	}
}

func TestGetListKafka(t *testing.T) {
	var pathCalled string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathCalled = r.URL.Path

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := `[
		{ "Service": { "Address": "172.0.0.1", "Port": 9092 } },
		{ "Service": { "Address": "172.0.0.2", "Port": 9092 } }
		]`
		w.Write([]byte(resp))
	}))

	aG := appGroup{
		clusterName:        "lama",
		consulHosts:        []string{srv.URL},
		consulServiceNames: map[string]string{"kafka": "kafka"},
	}

	kafkaHosts, err := aG.GetListKafka()
	if err != nil {
		t.Fatalf("Should not return error, got: %v", err)
	}

	expectedKafkaHosts := []string{"172.0.0.1:9092", "172.0.0.2:9092"}
	if !reflect.DeepEqual(kafkaHosts, expectedKafkaHosts) {
		t.Errorf("Invalid Kafka Host, got:\n%v,\nwant:\n%v\n", kafkaHosts, expectedKafkaHosts)
	}

	expectedPathCalled := fmt.Sprintf("/v1/health/service/%s", aG.consulServiceNames["kafka"])
	if pathCalled != expectedPathCalled {
		t.Errorf("Invalid consul path called, got:\n%q,\nwant:\n%q\n", pathCalled, expectedPathCalled)
	}
}

func TestGetListAppGroups(t *testing.T) {
	requests := []map[string]string{}
	expectedRequests := []map[string]string{
		{"path": "/api/v2/profile_index", "page": "1"},
		{"path": "/api/v2/profile_index", "page": "2"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, map[string]string{
			"path": r.URL.Path,
			"page": r.URL.Query().Get("page"),
		})
		w.WriteHeader(http.StatusOK)

		bodyList := []string{}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page == 1 {
			for i := 1; i <= 10; i++ {
				bodyList = append(bodyList, fmt.Sprintf(`{"cluster_name": "%d", "app_group_secret": "abc"}`, i))
			}
		}

		if page == 2 {
			for i := 1; i <= 2; i++ {
				bodyList = append(bodyList, fmt.Sprintf(`{"cluster_name": "%d", "app_group_secret": "abc"}`, i))
			}
		}
		body := fmt.Sprintf(`[%s]`, strings.Join(bodyList, ","))
		w.Write([]byte(body))
	}))

	cfg := config.Config{
		BaritoMarketHost:             srv.URL,
		BaritoMarketToken:            "ABC007",
		BaritoMarketProfileIndexPath: "/api/v2/profile_index",
	}

	appGroups, err := GetListAppGroups(cfg)
	if err != nil {
		t.Fatalf("Should not return error, got: %v", err)
	}

	if !reflect.DeepEqual(expectedRequests, requests) {
		t.Fatalf("Should got called with these request:\n%v\ngot:\n%v\n", expectedRequests, requests)
	}

	if len(appGroups) != 12 {
		t.Fatalf("Should return 12 appgroups, got: %d", len(appGroups))
	}
}

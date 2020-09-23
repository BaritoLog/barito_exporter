package appgroup

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/BaritoLog/barito-blackbox-exporter/config"
	"github.com/Jeffail/gabs/v2"
	log "github.com/sirupsen/logrus"
)

type AppGroup interface {
	RefreshMetadata() error
	GetName() string
	GetClusterName() string
	GetSecret() string
	GetListES() ([]string, error)
	GetKibanaHost() (string, error)
}

type appGroup struct {
	name               string
	clusterName        string
	consulHosts        []string
	consulServiceNames map[string]string
	baritoMarketHost   string
	baritoMarketToken  string
	secret             string
}

func NewAppGroup(clusterName, secret string, cfg *config.Config) *appGroup {
	return &appGroup{
		clusterName:       clusterName,
		secret:            secret,
		baritoMarketHost:  cfg.BaritoMarketHost,
		baritoMarketToken: cfg.BaritoMarketToken,
	}
}

func (a *appGroup) GetName() string {
	return a.name
}

func (a *appGroup) GetClusterName() string {
	return a.clusterName
}

func (a *appGroup) GetSecret() string {
	return a.secret
}

func (a *appGroup) RefreshMetadata() error {
	rawJson, err := fetchAppgroupMetadata(a.clusterName, a.baritoMarketHost, a.baritoMarketToken)
	if err != nil {
		return err
	}
	g, err := gabs.ParseJSON(rawJson)
	if err != nil {
		return err
	}
	// get name
	name, ok := g.Path("name").Data().(string)
	if ok {
		a.name = name
	}

	// get consul_hosts
	consulHosts := []string{}
	for _, v := range g.Path("consul_hosts").Children() {
		h, ok := v.Data().(string)
		if ok {
			consulHosts = append(consulHosts, h)
		}
	}
	if len(consulHosts) > 0 {
		a.consulHosts = consulHosts
	}

	// get consul_service_names
	consulServiceNames := map[string]string{}
	for k, v := range g.Path("meta.service_names").ChildrenMap() {
		vString, _ := v.Data().(string)
		consulServiceNames[k] = vString

	}
	if len(consulServiceNames) > 0 {
		a.consulServiceNames = consulServiceNames
	}
	return nil
}

func (a *appGroup) GetListES() ([]string, error) {
	if len(a.consulHosts) == 0 {
		log.Errorf("Can't fetch ES, no consul to contacted to")
		return nil, errors.New("Can't fetch ES, no consul to contacted to")
	}

	serviceName, ok := a.consulServiceNames["elasticsearch"]
	if !ok {
		log.Errorf("Can't find elasticsearch service name")
		return nil, errors.New("Can't find elasticsearch service name")
	}
	for _, consul := range a.consulHosts {
		listES, err := fetchConsulServices(consul, serviceName)
		if err != nil {
			log.Errorf("Failed to fetch elasticsearch, error: %v", err)
			continue
		}
		return listES, nil
	}
	return nil, errors.New("No ES found")
}

func (a *appGroup) GetKibanaHost() (string, error) {
	if len(a.consulHosts) == 0 {
		log.Errorf("Can't fetch kibana, no consul to contacted to")
		return "", errors.New("Can't fetch kibana, no consul to contacted to")
	}

	serviceName, ok := a.consulServiceNames["kibana"]
	if !ok {
		log.Errorf("Can't find kibana service name")
		return "", errors.New("Can't find kibana service name")
	}
	for _, consul := range a.consulHosts {
		kibanaHost, err := fetchConsulServices(consul, serviceName)
		if err != nil || len(kibanaHost) == 0 {
			log.Errorf("Failed to fetch kibana, error: %v", err)
			continue
		}
		return kibanaHost[0], nil
	}
	return "", errors.New("No Kibana found")
}

func fetchConsulServices(consulHost, serviceName string) ([]string, error) {
	var c = &http.Client{
		Timeout: 5 * time.Second,
	}

	if !strings.HasPrefix(consulHost, "http") {
		consulHost = "http://" + consulHost
	}

	url := fmt.Sprintf("%s/v1/health/service/%s", consulHost, serviceName)
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return nil, errors.New("failed to create request")
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Got response status %d when fetch service %q from consul %q", resp.StatusCode, serviceName, consulHost)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	g, err := gabs.ParseJSON(body)
	if err != nil {
		return nil, err
	}

	var hosts []string
	for _, s := range g.Children() {
		host, hostOK := s.Path("Service.Address").Data().(string)
		port, portOK := s.Path("Service.Port").Data().(float64)
		if hostOK && portOK {
			hosts = append(hosts, fmt.Sprintf("http://%s:%d", host, int(port)))
		}
	}
	return hosts, nil
}

func fetchAppgroupMetadata(clusterName, baritoMarketHost, accessToken string) ([]byte, error) {
	var c = &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v2/profile_by_cluster_name?cluster_name=%s&access_token=%s", baritoMarketHost, clusterName, accessToken)
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return []byte(""), errors.New("failed to create request")
	}

	resp, err := c.Do(req)
	if err != nil {
		return []byte(""), err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Got response status %d", resp.StatusCode)
		log.Debugf("Request fetch metadata got status: %d, appGroup: %q", resp.StatusCode, clusterName)
		return []byte(""), err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return body, nil
}

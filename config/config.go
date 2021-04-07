package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	BaritoMarketHost             string
	BaritoMarketToken            string
	baritoMarketHost             string
	ProduceURL                   string
	ProduceAppPrefix             string
	BaritoMarketProfileIndexPath string
	ProduceInterval              time.Duration
	ProduceTimeout               time.Duration
	ProduceTimeField             string
	ESProbeInterval              time.Duration
	ESProbeTimeout               time.Duration
	KibanaProbeInterval          time.Duration
	KibanaProbeTimeout           time.Duration
	DeleteTopicInterval          time.Duration
}

func NewConfig() *Config {
	return &Config{
		BaritoMarketHost:             envOrDefaultString("BARITO_MARKET_HOST", "https://barito.golabs.io"),
		BaritoMarketToken:            envOrDefaultString("BARITO_MARKET_TOKEN", ""),
		BaritoMarketProfileIndexPath: envOrDefaultString("BARITO_MARKET_PROFILE_INDEX_PATH", "/api/v2/profile_index"),
		ProduceURL:                   envOrDefaultString("PRODUCE_URL", "https://barito-router.golabs.io/produce_batch"),
		ProduceAppPrefix:             envOrDefaultString("PRODUCE_APP_PREFIX", "barito-prober"),
		ProduceInterval:              time.Duration(envOrDefaultInt("PRODUCE_INTERVAL_SECOND", 30)) * time.Second,
		ProduceTimeout:               time.Duration(envOrDefaultInt("PRODUCE_TIMEOUT", 10)) * time.Second,
		ProduceTimeField:             envOrDefaultString("PRODUCE_TIME_FIELD", "barito_trace_time"),
		ESProbeInterval:              time.Duration(envOrDefaultInt("ES_PROBE_INTERVAL", 30)) * time.Second,
		ESProbeTimeout:               time.Duration(envOrDefaultInt("ES_PROBE_TIMEOUT", 10)) * time.Second,
		KibanaProbeInterval:          time.Duration(envOrDefaultInt("KIBANA_PROBE_INTERVAL", 60)) * time.Second,
		KibanaProbeTimeout:           time.Duration(envOrDefaultInt("KIBANA_PROBE_TIMEOUT", 30)) * time.Second,
		DeleteTopicInterval:          time.Duration(envOrDefaultInt("DELETE_TOPIC_INTERVAL", 3600)) * time.Second,
	}
}

func envOrDefaultString(envName, defaultValue string) string {
	if v := os.Getenv(envName); v != "" {
		return v
	}
	return defaultValue
}

func envOrDefaultInt(envName string, defaultValue int) int {
	if v := os.Getenv(envName); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return defaultValue
		}
		return i
	}
	return defaultValue
}

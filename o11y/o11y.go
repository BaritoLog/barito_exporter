package o11y

import "github.com/prometheus/client_golang/prometheus"

const (
	REASON_PROBE_ELASTICSEARCH_FAILED_GET_LIST_FROM_CONSUL = "failed_get_list_from_consul"
	REASON_PROBE_ELASTICSEARCH_NO_ELASTICSEARCH_FOUND      = "no_elasticsearch_found"
	REASON_PROBE_ELASTICSEARCH_REQUEST_FAILED              = "request_failed"
	REASON_PROBE_ELASTICSEARCH_GET_DATA_FAILED             = "get_data_failed"
	REASON_PROBE_ELASTICSEARCH_FAILED_FETCH_METADATA       = "failed_fetch_metadata"
	REASON_PROBE_KIBANA_REQUEST_FAILED                     = "request_failed"
	REASON_PROBE_KIBANA_FAILED_GET_KIBANA_FROM_CONSUL      = "failed_get_kibana_from_consul"
	REASON_PROBE_KIBANA_FAILED_FETCH_METADATA              = "failed_fetch_metadata"
	REASON_PROBE_KIBANA_NO_KIBANA_FOUND                    = "no_kibana_found"
)

type MetricRecorder interface {
	IncreasePushLogSuccess(appGroup string)
	IncreasePushLogFailed(appGroup string)
	IncreaseProbeElasticSearchSuccess(appGroup string)
	IncreaseProbeElasticSearchFailed(appGroup, reason string)
	IncreaseProbeKibanaSuccess(appGroup string)
	IncreaseProbeKibanaFailed(appGroup, reason string)
	SetProbeElasticsearchDelay(appGroup string, delaySecond float64)
}

type metricRecorder struct {
	registry                        *prometheus.Registry
	metricPushLogSuccess            *prometheus.CounterVec
	metricPushLogFailed             *prometheus.CounterVec
	metricProbeElasticSearchSuccess *prometheus.CounterVec
	metricProbeElasticSearchFailed  *prometheus.CounterVec
	metricProbeElasticDelaySecond   *prometheus.GaugeVec
	metricProbeKibanaSuccess        *prometheus.CounterVec
	metricProbeKibanaFailed         *prometheus.CounterVec
}

func NewMetricRecorder() *metricRecorder {
	r := prometheus.NewRegistry()

	metricPushLogSuccess := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "barito_push_log_success",
			Help: "Number push log success",
		}, []string{"app_group"},
	)
	metricPushLogFailed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "barito_push_log_failed",
			Help: "Number push log failed",
		}, []string{"app_group"},
	)
	metricProbeElasticSearchSuccess := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "barito_probe_elasticsearch_success",
			Help: "Number probe elasticsearch success",
		}, []string{"app_group"},
	)
	metricProbeElasticSearchFailed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "barito_probe_elasticsearch_failed",
			Help: "Number probe elasticsearch failed",
		}, []string{"app_group", "reason"},
	)
	metricProbeElasticDelaySecond := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "barito_probe_elasticsearch_delay_second",
			Help: "Number of second the delay between current time and last log time",
		}, []string{"app_group"},
	)
	metricProbeKibanaSuccess := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "barito_probe_kibana_success",
			Help: "Number probe kibana success",
		}, []string{"app_group"},
	)
	metricProbeKibanaFailed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "barito_probe_kibana_failed",
			Help: "Number probe kibana failed",
		}, []string{"app_group", "reason"},
	)

	r.MustRegister(metricPushLogSuccess)
	r.MustRegister(metricPushLogFailed)
	r.MustRegister(metricProbeElasticSearchSuccess)
	r.MustRegister(metricProbeElasticSearchFailed)
	r.MustRegister(metricProbeElasticDelaySecond)
	r.MustRegister(metricProbeKibanaSuccess)
	r.MustRegister(metricProbeKibanaFailed)

	return &metricRecorder{
		registry:                        r,
		metricPushLogSuccess:            metricPushLogSuccess,
		metricPushLogFailed:             metricPushLogFailed,
		metricProbeElasticSearchSuccess: metricProbeElasticSearchSuccess,
		metricProbeElasticSearchFailed:  metricProbeElasticSearchFailed,
		metricProbeElasticDelaySecond:   metricProbeElasticDelaySecond,
		metricProbeKibanaSuccess:        metricProbeKibanaSuccess,
		metricProbeKibanaFailed:         metricProbeKibanaFailed,
	}
}

func (mR *metricRecorder) IncreasePushLogSuccess(appGroup string) {
	mR.metricPushLogSuccess.WithLabelValues(appGroup).Inc()
	mR.metricPushLogFailed.WithLabelValues(appGroup).Add(0)
}

func (mR *metricRecorder) IncreasePushLogFailed(appGroup string) {
	mR.metricPushLogFailed.WithLabelValues(appGroup).Inc()
	mR.metricPushLogSuccess.WithLabelValues(appGroup).Add(0)
}

func (mR *metricRecorder) IncreaseProbeElasticSearchSuccess(appGroup string) {
	mR.metricProbeElasticSearchSuccess.WithLabelValues(appGroup).Inc()
	mR.metricProbeElasticSearchFailed.WithLabelValues(appGroup, "").Add(0)
}

func (mR *metricRecorder) IncreaseProbeElasticSearchFailed(appGroup, reason string) {
	mR.metricProbeElasticSearchFailed.WithLabelValues(appGroup, reason).Inc()
	mR.metricProbeElasticSearchSuccess.WithLabelValues(appGroup).Add(0)
}

func (mR *metricRecorder) SetProbeElasticsearchDelay(appGroup string, delaySecond float64) {
	mR.metricProbeElasticDelaySecond.WithLabelValues(appGroup).Set(delaySecond)
}

func (mR *metricRecorder) IncreaseProbeKibanaSuccess(appGroup string) {
	mR.metricProbeKibanaSuccess.WithLabelValues(appGroup).Inc()
	mR.metricProbeKibanaFailed.WithLabelValues(appGroup, "").Add(0)
}

func (mR *metricRecorder) IncreaseProbeKibanaFailed(appGroup, reason string) {
	mR.metricProbeKibanaFailed.WithLabelValues(appGroup, reason).Inc()
	mR.metricProbeKibanaSuccess.WithLabelValues(appGroup).Add(0)
}

func (mR *metricRecorder) GetRegistry() *prometheus.Registry {
	return mR.registry
}

package o11y

import "github.com/prometheus/client_golang/prometheus"

type MetricRecorder interface {
	IncreasePushLogSuccess(appGroup string)
	IncreasePushLogFailed(appGroup string)
}

type metricRecorder struct {
	registry             *prometheus.Registry
	metricPushLogSuccess *prometheus.CounterVec
	metricPushLogFailed  *prometheus.CounterVec
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

	r.MustRegister(metricPushLogSuccess)
	r.MustRegister(metricPushLogFailed)

	return &metricRecorder{
		registry:             r,
		metricPushLogSuccess: metricPushLogSuccess,
		metricPushLogFailed:  metricPushLogFailed,
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

func (mR *metricRecorder) GetRegistry() *prometheus.Registry {
	return mR.registry
}

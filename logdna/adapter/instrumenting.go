package adapter

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	statusLabel = "status"
	actionLabel = "action"

	// actions
	streamActionLabel      = "stream"
	readQueueActionLabel   = "read_queue"
	flushBufferActionLabel = "flush_buffer"
)

var (
	emptyPrometheusLabels = prometheus.Labels{}

	actionLabelsValues = []string{
		streamActionLabel,
		readQueueActionLabel,
		flushBufferActionLabel,
	}
	// logDNAStatusCodes are the possible response codes from logDNA server when shipping logs
	logDNAStatusCodes = []string{"200", "400", "403", "500", "501", "502", "503"}
)

type instrumentingAdapter struct {
	// logdna adapter metrics
	// general metrics
	buildInfo *prometheus.GaugeVec
	// performance metrics
	requestCount       *prometheus.CounterVec
	requestLatency     *prometheus.HistogramVec
	logDNARequestCount *prometheus.CounterVec

	// error metrics
	streamMarshalDataErrorCount     *prometheus.CounterVec
	flushBufferEncodeDataErrorCount *prometheus.CounterVec
	flushBufferLogDNACodeErrorCount *prometheus.CounterVec

	logger *log.Logger
}

// newInstrumentingAdapter returns an instance of an instrumenting Adapter.
func newInstrumentingAdapter() *instrumentingAdapter {
	logger := log.New(os.Stdout, "Instrumenting ", log.LstdFlags)

	// ## Define metrics
	// general metrics
	buildInfo := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "build_info",
		Help:      "Number of logspout logdna adapter builds.",
	},
		[]string{"go_version", "adapter_version"},
	)

	// performance metrics
	requestCount := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "request_total",
		Help:      "Number of stream requests received.",
	}, []string{actionLabel})
	requestLatency := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "request_latency_seconds",
		Help:      "Total duration of stream requests in seconds.",
	}, []string{actionLabel})
	logDNARequestCount := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "logdna_request_total",
		Help:      "Number of requests to LogDNA service while flushing message.",
	}, []string{actionLabel, statusLabel})

	// error metrics
	streamMarshalDataErrorCount := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "stream_marshal_data_error_total",
		Help:      "Number of errors when marshalling the received stream data.",
	}, []string{})
	flushBufferEncodeDataErrorCount := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "flush_buffer_encode_data_error_total",
		Help:      "Number of errors when encoding the data while flushing message.",
	}, []string{})
	flushBufferLogDNACodeErrorCount := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "logspout_logdna",
		Subsystem: "adapter",
		Name:      "flush_buffer_logdna_code_error_total",
		Help:      "Number of errors when posting the message to LogDNA because of a code error when performing the request while flushing message.",
	}, []string{})

	// ## Initialize metrics: why? https://www.robustperception.io/existential-issues-with-metrics
	// performance metrics
	for _, action := range actionLabelsValues {
		if _, err := requestCount.GetMetricWith(prometheus.Labels{
			actionLabel: action,
		}); err != nil {
			logger.Println(fmt.Errorf("cannot initialize request counter metric: %s", err))
		}
	}
	for _, action := range actionLabelsValues {
		if _, err := requestLatency.GetMetricWith(prometheus.Labels{
			actionLabel: action,
		}); err != nil {
			logger.Println(fmt.Errorf("cannot initialize request latency metric: %s", err))
		}
	}

	for _, status := range logDNAStatusCodes {
		for _, action := range actionLabelsValues {
			if _, err := logDNARequestCount.GetMetricWith(prometheus.Labels{
				actionLabel: action,
				statusLabel: status,
			}); err != nil {
				logger.Println(fmt.Errorf("cannot initialize flush buffer logdna client error counter metric: %s", err))
			}
		}
	}

	return &instrumentingAdapter{
		buildInfo: buildInfo,

		requestCount:       requestCount,
		requestLatency:     requestLatency,
		logDNARequestCount: logDNARequestCount,

		streamMarshalDataErrorCount:     streamMarshalDataErrorCount,
		flushBufferEncodeDataErrorCount: flushBufferEncodeDataErrorCount,
		flushBufferLogDNACodeErrorCount: flushBufferLogDNACodeErrorCount,

		logger: logger,
	}
}

func (inst *instrumentingAdapter) fireBuildInfo(adapterVersion string) {
	inst.buildInfo.WithLabelValues(runtime.Version(), adapterVersion).Set(1)
}

func (inst *instrumentingAdapter) fireAddRequestCount(action string) {
	inst.requestCount.With(prometheus.Labels{
		actionLabel: action,
	}).Add(1)
}

func (inst *instrumentingAdapter) fireAddRequestLatency(begin time.Time, action string) {
	inst.requestLatency.With(prometheus.Labels{
		actionLabel: action,
	}).Observe(time.Since(begin).Seconds())
}

func (inst *instrumentingAdapter) fireAddStreamMarshalDataError() {
	inst.streamMarshalDataErrorCount.With(emptyPrometheusLabels).Add(1)
}

func (inst *instrumentingAdapter) fireAddFlushBufferEncodeDataError() {
	inst.flushBufferEncodeDataErrorCount.With(emptyPrometheusLabels).Add(1)
}

func (inst *instrumentingAdapter) fireAddFlushBufferLogDNACodeError() {
	inst.flushBufferLogDNACodeErrorCount.With(emptyPrometheusLabels).Add(1)
}

func (inst *instrumentingAdapter) fireAddLogDNAClientRequest(action string, statusCode string) {
	inst.logDNARequestCount.With(prometheus.Labels{
		statusLabel: statusCode,
		actionLabel: action,
	}).Add(1)
}

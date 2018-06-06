package startup_metrics

import (
	"github.com/eSailors/go-datadog"
	"github.com/flachnetz/startup"
	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

var log = logrus.WithField("prefix", "main")

type MetricsPrefix string

type MetricsOptions struct {
	Datadog struct {
		ApiKey   string        `long:"datadog-apikey" description:"Datadog app key to enable datadog metrics reporting."`
		Tags     string        `long:"datadog-tags" description:"Extra datadog tags to add to every metric. Comma or space separated list of key:value pairs."`
		Interval time.Duration `long:"datadog-report-interval" default:"60s" description:"Data collection and reporting interval."`
	}

	Inputs struct {
		// Prefix to apply to all metrics. This must not be empty.
		MetricsPrefix string `validate:"required"`

		// Disable capture of runtime metrics for some reasons
		NoRuntimeMetrics bool
	}

	once sync.Once
}

func (opts *MetricsOptions) Initialize() {
	opts.once.Do(func() {
		registry := metrics.DefaultRegistry

		if prefix := strings.TrimSuffix(opts.Inputs.MetricsPrefix, "."); prefix != "" {
			log.Debugf("Prefixing all metrics with '%s'", prefix)
			registry = metrics.NewPrefixedChildRegistry(registry, prefix+".")
			metrics.DefaultRegistry = registry

		} else {
			startup.Panicf("Metrics prefix must be set")
			return
		}

		if !opts.Inputs.NoRuntimeMetrics {
			captureRuntimeMetrics(registry)
		}

		if opts.Datadog.ApiKey != "" {
			err := opts.setupDatadogMetricsReporter(registry)
			startup.PanicOnError(err, "Cannot start datadog metrics reporter")
		}
	})
}

func captureRuntimeMetrics(registry metrics.Registry) {
	log.Debug("Start capturing of golang runtime metrics")

	// start capturing of metrics
	metrics.RegisterRuntimeMemStats(registry)
	go metrics.CaptureRuntimeMemStats(registry, 5*time.Second)
}

func (opts *MetricsOptions) setupDatadogMetricsReporter(registry metrics.Registry) error {
	node, err := os.Hostname()
	if err != nil {
		return errors.WithMessage(err, "get hostname of machine")
	}

	client := datadog.New("", opts.Datadog.ApiKey)

	tags := strings.FieldsFunc(opts.Datadog.Tags, isCommaOrSpace)
	tags = append(tags, "node:"+node)

	log.Infof("Starting datadog metrics reporting with tags: %s",
		strings.Join(tags, ", "))

	reporter := datadog.Reporter(client, registry, tags)
	go reporter.Start(opts.Datadog.Interval)

	return nil
}

func isCommaOrSpace(r rune) bool {
	return r == ',' || unicode.IsSpace(r)
}

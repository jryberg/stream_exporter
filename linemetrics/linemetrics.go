// Package linemetrics turns configured regular expressions into Prometheus
// metrics that are updated for each matching line.
package linemetrics

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type LineMetric interface {
	MatchLine(s string)
	Name() string
}

type BaseLineMetric struct {
	name    string
	pattern regexp.Regexp
	labels  []string
}

func (m BaseLineMetric) Name() string {
	return m.name
}

type config struct {
	Metrics []MetricsConfig
}

func ReadPatternConfig(path string) ([]MetricsConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Strict, so that a misspelled key is an error rather than a silently empty field
	var config config
	err = yaml.UnmarshalStrict(content, &config)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(config.Metrics))
	for i, m := range config.Metrics {
		if err := m.validate(); err != nil {
			return nil, fmt.Errorf("metric %d (%q): %w", i+1, m.Name, err)
		}
		if seen[m.Name] {
			return nil, fmt.Errorf("metric %d (%q): name is already used by another metric", i+1, m.Name)
		}
		seen[m.Name] = true
	}

	return config.Metrics, nil
}

func NewLineMetric(config MetricsConfig) (LineMetric, prometheus.Collector) {
	pattern := regexp.MustCompile(config.Pattern)
	labels := pattern.SubexpNames()[1:] // First element is entire expression

	var lineMetric LineMetric
	base := BaseLineMetric{
		name:    config.Name,
		pattern: *pattern,
		labels:  labels,
	}
	var collector prometheus.Collector
	switch config.Kind {
	case counter:
		lineMetric, collector = NewCounterLineMetric(base)
	case gauge:
		lineMetric, collector = NewGaugeLineMetric(base)
	case histogram:
		lineMetric, collector = NewHistogramLineMetric(base, config.HistogramConfig)
	case summary:
		lineMetric, collector = NewSummaryLineMetric(base, config.SummaryConfig)
	}

	return lineMetric, collector
}

func getValueCaptureIndex(labels []string) (int, error) {
	foundValue := false
	valueIdx := 0
	for idx, l := range labels {
		if l == "value" {
			foundValue = true
			valueIdx = idx
			break
		}
	}
	if !foundValue {
		return valueIdx, errors.New("no named capture group for 'value'")
	}

	return valueIdx, nil
}

func getLabelsAndValue(matches []string, valueIdx int) ([]string, float64, error) {
	captures := matches[1:]
	valueStr := captures[valueIdx]
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Warnf("Unable to convert %s to float\n", valueStr)
		return nil, 0, err
	}
	labels := append(captures[0:valueIdx], captures[valueIdx+1:]...)
	return labels, value, nil
}

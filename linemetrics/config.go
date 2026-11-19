package linemetrics

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"
)

type metricKind string

const (
	counter   metricKind = "counter"
	gauge     metricKind = "gauge"
	histogram metricKind = "histogram"
	summary   metricKind = "summary"
)

type MetricsConfig struct {
	Name    string
	Kind    metricKind
	Pattern string

	HistogramConfig `yaml:",inline"`
	SummaryConfig   `yaml:",inline"`
}

// validate catches configuration mistakes that would otherwise panic at startup or silently match every line
func (c MetricsConfig) validate() error {
	if c.Name == "" {
		return errors.New("name is not set")
	}
	switch c.Kind {
	case counter, gauge, histogram, summary:
	default:
		return fmt.Errorf("unknown kind %q, valid kinds are counter, gauge, histogram and summary", c.Kind)
	}
	if c.Pattern == "" {
		return errors.New("pattern is not set")
	}
	pattern, err := regexp.Compile(c.Pattern)
	if err != nil {
		return err
	}
	labels := pattern.SubexpNames()[1:] // First element is entire expression
	if slices.Contains(labels, "") {
		return errors.New("pattern has an unnamed capture group, use (?P<name>...) for a label or (?:...) for grouping")
	}
	if c.Kind != counter {
		if _, err := getValueCaptureIndex(labels); err != nil {
			return err
		}
	}
	return nil
}

type HistogramConfig struct {
	Buckets []float64
}

type SummaryConfig struct {
	Objectives map[float64]float64
	MaxAge     time.Duration `yaml:"max_age"`
	AgeBuckets uint32        `yaml:"age_buckets"`
	BufCap     uint32        `yaml:"buf_cap"`
}

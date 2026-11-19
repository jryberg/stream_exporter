package linemetrics

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestUnmarshalSummaryConfig(t *testing.T) {
	configSlice, err := ReadPatternConfig("testdata/summary.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(configSlice) != 1 {
		t.Fatal("Unexpected length of test data")
	}
	config := configSlice[0].SummaryConfig

	expectedMaxAge, _ := time.ParseDuration("15m")
	expectedConfig := SummaryConfig{
		Objectives: map[float64]float64{
			0.1: 0.1,
			0.2: 0.2,
			0.3: 0.3,
			0.4: 0.4,
		},
		MaxAge:     expectedMaxAge,
		AgeBuckets: 3,
		BufCap:     500,
	}

	if !reflect.DeepEqual(expectedConfig, config) {
		t.Fatalf("Mismatch! Expected\n\t%v\nGot\n\t%v\n", expectedConfig, config)
	}
}

func TestUnmarshalHistogramConfig(t *testing.T) {
	configSlice, err := ReadPatternConfig("testdata/histogram.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(configSlice) != 1 {
		t.Fatal("Unexpected length of test data")
	}
	config := configSlice[0].HistogramConfig

	expectedConfig := HistogramConfig{
		Buckets: []float64{0.1, 0.2, 0.5},
	}

	if !reflect.DeepEqual(expectedConfig, config) {
		t.Fatalf("Mismatch! Expected\n\t%v\nGot\n\t%v\n", expectedConfig, config)
	}
}

func TestReadPatternConfigValidation(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid counter with label",
			yaml: "metrics:\n  - name: a\n    kind: counter\n    pattern: \"(?P<section>\\\\w+) (?:x|y)\"\n",
		},
		{
			name:    "misspelled key",
			yaml:    "metrics:\n  - name: a\n    kind: counter\n    patern: x\n",
			wantErr: "field patern not found",
		},
		{
			name:    "missing name",
			yaml:    "metrics:\n  - kind: counter\n    pattern: x\n",
			wantErr: "name is not set",
		},
		{
			name:    "unknown kind",
			yaml:    "metrics:\n  - name: a\n    kind: countr\n    pattern: x\n",
			wantErr: `unknown kind "countr"`,
		},
		{
			name:    "missing pattern",
			yaml:    "metrics:\n  - name: a\n    kind: counter\n",
			wantErr: "pattern is not set",
		},
		{
			name:    "invalid regexp",
			yaml:    "metrics:\n  - name: a\n    kind: counter\n    pattern: \"(x\"\n",
			wantErr: "missing closing )",
		},
		{
			name:    "unnamed capture group",
			yaml:    "metrics:\n  - name: a\n    kind: counter\n    pattern: \"(x|y)\"\n",
			wantErr: "unnamed capture group",
		},
		{
			name:    "gauge without value",
			yaml:    "metrics:\n  - name: a\n    kind: gauge\n    pattern: x\n",
			wantErr: "no named capture group for 'value'",
		},
		{
			name:    "duplicate name",
			yaml:    "metrics:\n  - name: a\n    kind: counter\n    pattern: x\n  - name: a\n    kind: counter\n    pattern: y\n",
			wantErr: `metric 2 ("a"): name is already used`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(c.yaml), 0o600); err != nil {
				t.Fatal(err)
			}

			_, err := ReadPatternConfig(path)
			if c.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("Expected error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

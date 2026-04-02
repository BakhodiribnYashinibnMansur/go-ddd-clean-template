package config_test

import (
	"os"
	"testing"

	"gct/config"

	"gopkg.in/yaml.v3"
)

func TestMetricsLatencyDefaults(t *testing.T) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse config.yaml: %v", err)
	}

	m := cfg.Metrics
	if !m.LatencyEnabled {
		t.Error("expected LatencyEnabled=true")
	}
	if m.LatencyP95Threshold != "200ms" {
		t.Errorf("expected LatencyP95Threshold=200ms, got %s", m.LatencyP95Threshold)
	}
	if m.LatencyP99Threshold != "500ms" {
		t.Errorf("expected LatencyP99Threshold=500ms, got %s", m.LatencyP99Threshold)
	}
	if m.LatencyWindowSec != 60 {
		t.Errorf("expected LatencyWindowSec=60, got %d", m.LatencyWindowSec)
	}
	if m.LatencyLogIntervalSec != 10 {
		t.Errorf("expected LatencyLogIntervalSec=10, got %d", m.LatencyLogIntervalSec)
	}
}

package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCollectors(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]interface{}
		wantCollectors int
		wantErrors     int
	}{
		{
			name:           "success",
			config:         nil,
			wantCollectors: 4,
			wantErrors:     0,
		},
		{
			name: "1 failure",
			config: map[string]interface{}{
				"crm-mon-path": "foo",
			},
			wantCollectors: 3,
			wantErrors:     1,
		},
		{
			name: "2 failures",
			config: map[string]interface{}{
				"crm-mon-path":              "foo",
				"corosync-cfgtoolpath-path": "foo",
			},
			wantCollectors: 2,
			wantErrors:     2,
		},
		{
			name: "3 failures",
			config: map[string]interface{}{
				"crm-mon-path":              "foo",
				"corosync-cfgtoolpath-path": "foo",
				"sbd-path":                  "foo",
			},
			wantCollectors: 1,
			wantErrors:     3,
		},
		{
			name: "4 failures",
			config: map[string]interface{}{
				"crm-mon-path":              "foo",
				"corosync-cfgtoolpath-path": "foo",
				"sbd-path":                  "foo",
				"drbdsetup-path":            "foo",
			},
			wantCollectors: 0,
			wantErrors:     4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheus.DefaultRegisterer = prometheus.NewRegistry()
			prometheus.DefaultGatherer = prometheus.NewRegistry()
			config := viper.New()
			config.SetConfigFile("test/test_config.yaml")
			_ = config.ReadInConfig()
			_ = config.MergeConfigMap(tt.config)
			collectors, errors := registerCollectors(config)
			assert.Len(t, collectors, tt.wantCollectors)
			assert.Len(t, errors, tt.wantErrors)
		})
	}
}

package metricer

import "time"

type ApiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message,omitempty"`
	} `json:"error"`
}

type Health struct {
	Status string `json:"status"`
	Metric string `json:"metric,omitempty"`
	Msg    string `json:"message,omitempty"`
}

type Metrics struct {
	Instance string                 `json:"instance"`
	Uptime   time.Duration          `json:"uptime"`
	Metrics  map[string]interface{} `json:"metrics"`
}

type RawMetrics map[string]interface{}

type LogLevels map[string]string

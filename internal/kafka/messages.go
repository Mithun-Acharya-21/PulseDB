package kafka

type CheckJob struct {
	MonitorID string `json:"monitor_id"`

	MonitorName string `json:"monitor_name"`

	URL string `json:"url"`
}

type CheckResult struct {
	MonitorID string `json:"monitor_id"`

	StatusCode int `json:"status_code"`

	ResponseMs int `json:"response_ms"`

	Error *string `json:"error,omitempty"`
}

package execlog

import (
	"encoding/json"
	"time"
)

type Entry struct {
	ExecutionID     string          `json:"execution_id"`
	StartedAt       time.Time       `json:"started_at"`
	Code            string          `json:"code"`
	CodeSHA256      string          `json:"code_sha256"`
	DurationMS      int64           `json:"duration_ms"`
	ResultCode      string          `json:"result_code"`
	Result          json.RawMessage `json:"result"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	ServiceRevision string          `json:"service_revision,omitempty"`
	ClientIP        string          `json:"client_ip,omitempty"`
	UserAgent       string          `json:"user_agent,omitempty"`
}

package execlog

import (
	"context"
	"fmt"

	bq "google.golang.org/api/bigquery/v2"
)

type BigQueryLogger struct {
	ProjectID string
	DatasetID string
	TableID   string
	Client    *bq.Service
}

func (l *BigQueryLogger) Log(ctx context.Context, entry *Entry) error {
	data := map[string]bq.JsonValue{
		"execution_id": entry.ExecutionID,
		"started_at":   entry.StartedAt.UTC().Format("2006-01-02 15:04:05.999999"),
		"code":         entry.Code,
		"code_sha256":  entry.CodeSHA256,
		"duration_ms":  entry.DurationMS,
		"result_code":  entry.ResultCode,
	}

	if entry.Result != nil {
		data["result"] = string(entry.Result)
	}

	for key, value := range map[string]string{
		"error_message":    entry.ErrorMessage,
		"service_revision": entry.ServiceRevision,
		"client_ip":        entry.ClientIP,
		"user_agent":       entry.UserAgent,
	} {
		if value == "" {
			continue
		}

		data[key] = value
	}

	resp, err := l.Client.Tabledata.InsertAll(l.ProjectID, l.DatasetID, l.TableID, &bq.TableDataInsertAllRequest{
		Rows: []*bq.TableDataInsertAllRequestRows{{
			InsertId: entry.ExecutionID,
			Json:     data,
		}},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("insert execution log: %w", err)
	}

	if len(resp.InsertErrors) == 0 {
		// no errors
		return nil
	}

	for _, rowError := range resp.InsertErrors {
		for _, detail := range rowError.Errors {
			return fmt.Errorf("insert execution log row %d: %s: %s", rowError.Index, detail.Reason, detail.Message)
		}
	}

	return fmt.Errorf("insert execution log: %d row errors", len(resp.InsertErrors))
}

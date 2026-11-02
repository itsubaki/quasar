CREATE TABLE IF NOT EXISTS `quasar.execution_log` (
  execution_id     STRING NOT NULL,
  started_at       TIMESTAMP NOT NULL,
  code             STRING NOT NULL,
  code_sha256      STRING NOT NULL,
  duration_ms      INT64 NOT NULL,
  result_code      STRING NOT NULL,
  result           JSON,
  error_message    STRING,
  service_revision STRING,
  client_ip        STRING,
  user_agent       STRING
)
PARTITION BY DATE(started_at)
CLUSTER BY code_sha256
OPTIONS (
  require_partition_filter = TRUE,
  description = 'OpenQASM simulation execution logs with results'
);

-- queries/incident.sql
SELECT 
  d.container_name, 
  d.log_line, 
  d.timestamp, 
  g.title        AS last_pr, 
  g.merged_at    AS deploy_time, 
  sl.text        AS slack_discussion, 
  az.metric_name AS azure_alert, 
  az.value       AS metric_value
FROM docker.logs d
JOIN github.pull_requests g 
  ON g.merged_at <= d.timestamp
JOIN slack.messages sl 
  ON sl.channel = '#incidents'
JOIN azure.metrics az 
  ON az.triggered_at >= g.merged_at
WHERE d.log_level = 'ERROR' 
  AND az.severity = 'critical'
ORDER BY d.timestamp DESC
LIMIT 20;
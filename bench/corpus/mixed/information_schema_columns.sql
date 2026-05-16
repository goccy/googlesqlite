CREATE TABLE `proj.ds.users` (
  user_id INT64,
  email STRING,
  signup_at TIMESTAMP,
  preferences ARRAY<STRING>
);
CREATE TABLE `proj.ds.events` (
  event_id INT64,
  user_id INT64,
  kind STRING,
  payload STRING,
  ts TIMESTAMP
);
CREATE TABLE `proj.ds.sessions` (
  session_id INT64,
  user_id INT64,
  started_at TIMESTAMP,
  duration_seconds INT64
);
-- @query
SELECT table_name, column_name, data_type, ordinal_position
  FROM ds.INFORMATION_SCHEMA.COLUMNS
 ORDER BY table_name, ordinal_position

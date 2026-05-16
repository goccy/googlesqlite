CREATE TABLE events (id INT64, ts TIMESTAMP, payload JSON);
INSERT INTO events (id, ts, payload)
SELECT n,
       TIMESTAMP_ADD(TIMESTAMP '2024-01-01 00:00:00 UTC', INTERVAL MOD(n, 86400) SECOND),
       PARSE_JSON(CONCAT(
         '{"event":"', CASE MOD(n, 4) WHEN 0 THEN 'view' WHEN 1 THEN 'click' WHEN 2 THEN 'purchase' ELSE 'logout' END,
         '","user":{"id":', CAST(n AS STRING),
         ',"region":"', CASE MOD(n, 3) WHEN 0 THEN 'us' WHEN 1 THEN 'eu' ELSE 'apac' END,
         '"},"props":{"value":', CAST(MOD(n, 1000) AS STRING), '}}'
       ))
FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
-- @query
SELECT JSON_VALUE(payload, '$.user.region') AS region,
       JSON_VALUE(payload, '$.event')       AS event,
       COUNT(*)                             AS n,
       SUM(SAFE_CAST(JSON_VALUE(payload, '$.props.value') AS INT64)) AS total_value
  FROM events
 WHERE JSON_VALUE(payload, '$.event') IN ('click', 'purchase')
 GROUP BY region, event
 ORDER BY region, event

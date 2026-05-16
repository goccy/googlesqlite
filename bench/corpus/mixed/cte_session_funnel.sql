CREATE TABLE events (user_id INT64, session_id INT64, kind STRING, ts TIMESTAMP);
INSERT INTO events (user_id, session_id, kind, ts)
SELECT MOD(n, 200),
       MOD(n, 50),
       CASE MOD(n, 4) WHEN 0 THEN 'view' WHEN 1 THEN 'add_to_cart' WHEN 2 THEN 'checkout' ELSE 'purchase' END,
       TIMESTAMP_ADD(TIMESTAMP '2024-04-01 00:00:00 UTC', INTERVAL MOD(n, 86400) SECOND)
FROM UNNEST(GENERATE_ARRAY(1, 4000)) AS n;
-- @query
WITH per_session AS (
  SELECT session_id,
         COUNTIF(kind = 'view')        AS views,
         COUNTIF(kind = 'add_to_cart') AS adds,
         COUNTIF(kind = 'checkout')    AS checkouts,
         COUNTIF(kind = 'purchase')    AS purchases
    FROM events
   GROUP BY session_id
)
SELECT
  COUNT(*)                                                    AS sessions,
  COUNTIF(views > 0)                                          AS viewed,
  COUNTIF(adds > 0)                                           AS added,
  COUNTIF(checkouts > 0)                                      AS checked_out,
  COUNTIF(purchases > 0)                                      AS purchased,
  SAFE_DIVIDE(COUNTIF(purchases > 0), COUNTIF(views > 0))     AS conversion_rate
FROM per_session

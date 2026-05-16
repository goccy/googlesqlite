CREATE TABLE events (bucket INT64, value INT64);
INSERT INTO events (bucket, value)
  SELECT MOD(n, 10), n FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
-- @query
SELECT bucket, SUM(value), COUNT(*) FROM events GROUP BY bucket ORDER BY bucket

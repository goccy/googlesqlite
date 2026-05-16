CREATE TABLE events (bucket INT64, ts INT64);
INSERT INTO events (bucket, ts)
  SELECT MOD(n, 10), n FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
-- @query
SELECT bucket, ts,
       ROW_NUMBER() OVER (PARTITION BY bucket ORDER BY ts) AS rn
FROM events
ORDER BY bucket, ts

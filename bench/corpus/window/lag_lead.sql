CREATE TABLE prices (sym INT64, day INT64, px FLOAT64);
INSERT INTO prices (sym, day, px)
  SELECT MOD(n, 5), n, CAST(n AS FLOAT64) * 1.5
  FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
-- @query
SELECT sym, day, px,
       LAG(px) OVER (PARTITION BY sym ORDER BY day) AS prev_px,
       LEAD(px) OVER (PARTITION BY sym ORDER BY day) AS next_px
FROM prices
ORDER BY sym, day

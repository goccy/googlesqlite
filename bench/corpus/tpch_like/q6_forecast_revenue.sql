CREATE TABLE lineitem (
  l_extendedprice FLOAT64,
  l_discount      FLOAT64,
  l_quantity      FLOAT64,
  l_shipdate      DATE
);
INSERT INTO lineitem (l_extendedprice, l_discount, l_quantity, l_shipdate)
SELECT
  CAST((MOD(n, 1000) + 1) * 100 AS FLOAT64),
  CAST(MOD(n, 11) AS FLOAT64) / 100.0,
  CAST(MOD(n, 50) + 1 AS FLOAT64),
  DATE_ADD(DATE '1994-01-01', INTERVAL MOD(n, 730) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 5000)) AS n;
-- @query
SELECT SUM(l_extendedprice * l_discount) AS revenue
  FROM lineitem
 WHERE l_shipdate >= DATE '1994-01-01'
   AND l_shipdate <  DATE '1995-01-01'
   AND l_discount BETWEEN 0.05 AND 0.07
   AND l_quantity <  24

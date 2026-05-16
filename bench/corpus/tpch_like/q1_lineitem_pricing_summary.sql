CREATE TABLE lineitem (
  l_orderkey      INT64,
  l_quantity      FLOAT64,
  l_extendedprice FLOAT64,
  l_discount      FLOAT64,
  l_tax           FLOAT64,
  l_returnflag    STRING,
  l_linestatus    STRING,
  l_shipdate      DATE
);
INSERT INTO lineitem (l_orderkey, l_quantity, l_extendedprice, l_discount, l_tax, l_returnflag, l_linestatus, l_shipdate)
SELECT
  n,
  CAST(MOD(n, 50) + 1 AS FLOAT64),
  CAST((MOD(n, 1000) + 1) * 100 AS FLOAT64),
  CAST(MOD(n, 11) AS FLOAT64) / 100.0,
  CAST(MOD(n, 9)  AS FLOAT64) / 100.0,
  CASE MOD(n, 3) WHEN 0 THEN 'A' WHEN 1 THEN 'N' ELSE 'R' END,
  CASE MOD(n, 2) WHEN 0 THEN 'F' ELSE 'O' END,
  DATE_ADD(DATE '1995-01-01', INTERVAL MOD(n, 365) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 5000)) AS n;
-- @query
SELECT l_returnflag,
       l_linestatus,
       SUM(l_quantity)                                   AS sum_qty,
       SUM(l_extendedprice)                              AS sum_base_price,
       SUM(l_extendedprice * (1 - l_discount))           AS sum_disc_price,
       SUM(l_extendedprice * (1 - l_discount) * (1 + l_tax)) AS sum_charge,
       AVG(l_quantity)                                   AS avg_qty,
       AVG(l_extendedprice)                              AS avg_price,
       AVG(l_discount)                                   AS avg_disc,
       COUNT(*)                                          AS count_order
  FROM lineitem
 WHERE l_shipdate <= DATE '1995-12-01'
 GROUP BY l_returnflag, l_linestatus
 ORDER BY l_returnflag, l_linestatus

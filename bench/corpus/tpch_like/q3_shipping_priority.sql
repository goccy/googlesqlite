CREATE TABLE customer (c_custkey INT64, c_mktsegment STRING);
CREATE TABLE orders   (o_orderkey INT64, o_custkey INT64, o_orderdate DATE, o_shippriority INT64);
CREATE TABLE lineitem (l_orderkey INT64, l_extendedprice FLOAT64, l_discount FLOAT64, l_shipdate DATE);

INSERT INTO customer (c_custkey, c_mktsegment)
SELECT n, CASE MOD(n, 5) WHEN 0 THEN 'BUILDING' WHEN 1 THEN 'AUTOMOBILE' WHEN 2 THEN 'MACHINERY' WHEN 3 THEN 'FURNITURE' ELSE 'HOUSEHOLD' END
FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;

INSERT INTO orders (o_orderkey, o_custkey, o_orderdate, o_shippriority)
SELECT n, MOD(n, 1000) + 1,
       DATE_ADD(DATE '1995-01-01', INTERVAL MOD(n, 200) DAY),
       MOD(n, 3)
FROM UNNEST(GENERATE_ARRAY(1, 2000)) AS n;

INSERT INTO lineitem (l_orderkey, l_extendedprice, l_discount, l_shipdate)
SELECT MOD(n, 2000) + 1,
       CAST(MOD(n, 1000) * 10 AS FLOAT64),
       CAST(MOD(n, 11) AS FLOAT64) / 100.0,
       DATE_ADD(DATE '1995-03-15', INTERVAL MOD(n, 60) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 5000)) AS n;
-- @query
SELECT l.l_orderkey,
       SUM(l.l_extendedprice * (1 - l.l_discount)) AS revenue,
       o.o_orderdate,
       o.o_shippriority
  FROM customer c
  JOIN orders   o ON c.c_custkey = o.o_custkey
  JOIN lineitem l ON l.l_orderkey = o.o_orderkey
 WHERE c.c_mktsegment = 'BUILDING'
   AND o.o_orderdate < DATE '1995-03-15'
   AND l.l_shipdate  > DATE '1995-03-15'
 GROUP BY l.l_orderkey, o.o_orderdate, o.o_shippriority
 ORDER BY revenue DESC, o.o_orderdate
 LIMIT 10

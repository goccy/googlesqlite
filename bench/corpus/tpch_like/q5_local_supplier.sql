CREATE TABLE region   (r_regionkey INT64, r_name STRING);
CREATE TABLE nation   (n_nationkey INT64, n_regionkey INT64, n_name STRING);
CREATE TABLE supplier (s_suppkey INT64, s_nationkey INT64);
CREATE TABLE customer (c_custkey INT64, c_nationkey INT64);
CREATE TABLE orders   (o_orderkey INT64, o_custkey INT64, o_orderdate DATE);
CREATE TABLE lineitem (l_orderkey INT64, l_suppkey INT64, l_extendedprice FLOAT64, l_discount FLOAT64);

INSERT INTO region (r_regionkey, r_name) VALUES
  (1, 'AFRICA'), (2, 'AMERICA'), (3, 'ASIA'), (4, 'EUROPE'), (5, 'MIDDLE EAST');
INSERT INTO nation (n_nationkey, n_regionkey, n_name)
SELECT n, MOD(n, 5) + 1, CONCAT('NATION_', CAST(n AS STRING))
FROM UNNEST(GENERATE_ARRAY(1, 25)) AS n;
INSERT INTO supplier (s_suppkey, s_nationkey)
SELECT n, MOD(n, 25) + 1 FROM UNNEST(GENERATE_ARRAY(1, 200)) AS n;
INSERT INTO customer (c_custkey, c_nationkey)
SELECT n, MOD(n, 25) + 1 FROM UNNEST(GENERATE_ARRAY(1, 500)) AS n;
INSERT INTO orders (o_orderkey, o_custkey, o_orderdate)
SELECT n, MOD(n, 500) + 1, DATE_ADD(DATE '1994-01-01', INTERVAL MOD(n, 365) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
INSERT INTO lineitem (l_orderkey, l_suppkey, l_extendedprice, l_discount)
SELECT MOD(n, 1000) + 1, MOD(n, 200) + 1,
       CAST(MOD(n, 5000) * 10 AS FLOAT64),
       CAST(MOD(n, 11) AS FLOAT64) / 100.0
FROM UNNEST(GENERATE_ARRAY(1, 3000)) AS n;
-- @query
SELECT n.n_name,
       SUM(l.l_extendedprice * (1 - l.l_discount)) AS revenue
  FROM customer c
  JOIN orders   o ON c.c_custkey   = o.o_custkey
  JOIN lineitem l ON l.l_orderkey  = o.o_orderkey
  JOIN supplier s ON l.l_suppkey   = s.s_suppkey AND s.s_nationkey = c.c_nationkey
  JOIN nation   n ON s.s_nationkey = n.n_nationkey
  JOIN region   r ON n.n_regionkey = r.r_regionkey
 WHERE r.r_name = 'ASIA'
   AND o.o_orderdate >= DATE '1994-01-01'
   AND o.o_orderdate <  DATE '1995-01-01'
 GROUP BY n.n_name
 ORDER BY revenue DESC

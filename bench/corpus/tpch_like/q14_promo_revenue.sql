CREATE TABLE part (p_partkey INT64, p_type STRING);
CREATE TABLE lineitem (l_partkey INT64, l_extendedprice FLOAT64, l_discount FLOAT64, l_shipdate DATE);

INSERT INTO part (p_partkey, p_type)
SELECT n,
       CASE MOD(n, 5)
         WHEN 0 THEN 'PROMO BURNISHED COPPER'
         WHEN 1 THEN 'STANDARD ANODIZED STEEL'
         WHEN 2 THEN 'PROMO PLATED BRASS'
         WHEN 3 THEN 'ECONOMY POLISHED NICKEL'
         ELSE        'LARGE BRUSHED TIN'
       END
FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
INSERT INTO lineitem (l_partkey, l_extendedprice, l_discount, l_shipdate)
SELECT MOD(n, 1000) + 1,
       CAST((MOD(n, 1000) + 1) * 100 AS FLOAT64),
       CAST(MOD(n, 11) AS FLOAT64) / 100.0,
       DATE_ADD(DATE '1995-01-01', INTERVAL MOD(n, 365) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 3000)) AS n;
-- @query
SELECT 100.00 * SUM(CASE
                      WHEN STARTS_WITH(p.p_type, 'PROMO')
                      THEN l.l_extendedprice * (1 - l.l_discount)
                      ELSE 0
                    END)
              / SUM(l.l_extendedprice * (1 - l.l_discount)) AS promo_revenue
  FROM lineitem l
  JOIN part     p ON p.p_partkey = l.l_partkey
 WHERE l.l_shipdate >= DATE '1995-09-01'
   AND l.l_shipdate <  DATE '1995-10-01'

CREATE TABLE users (id INT64, name STRING, region STRING, signed_up DATE);
INSERT INTO users (id, name, region, signed_up)
SELECT n,
       CONCAT('user_', CAST(n AS STRING)),
       CASE MOD(n, 4) WHEN 0 THEN 'us' WHEN 1 THEN 'eu' WHEN 2 THEN 'apac' ELSE 'sa' END,
       DATE_ADD(DATE '2023-01-01', INTERVAL MOD(n, 365) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 5000)) AS n;
-- @query
SELECT id, name, region, signed_up
  FROM users
 WHERE region = 'eu'
   AND signed_up >= DATE '2023-06-01'
 ORDER BY signed_up
 LIMIT 100

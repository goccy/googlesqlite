CREATE TABLE users    (id INT64, name STRING, region STRING, joined TIMESTAMP);
CREATE TABLE products (product_id INT64, sku STRING, price NUMERIC);
CREATE TABLE orders   (order_id INT64, user_id INT64, product_id INT64, qty INT64, placed_at TIMESTAMP);

INSERT INTO users (id, name, region, joined)
SELECT n,
       CONCAT('user_', CAST(n AS STRING)),
       CASE MOD(n, 4) WHEN 0 THEN 'us' WHEN 1 THEN 'eu' WHEN 2 THEN 'apac' ELSE 'sa' END,
       TIMESTAMP_ADD(TIMESTAMP '2023-01-01 00:00:00 UTC', INTERVAL MOD(n, 365) DAY)
FROM UNNEST(GENERATE_ARRAY(1, 500)) AS n;

INSERT INTO products (product_id, sku, price)
SELECT n,
       CONCAT('sku-', CAST(n AS STRING)),
       CAST(MOD(n * 17, 1000) + 1 AS NUMERIC) / 10
FROM UNNEST(GENERATE_ARRAY(1, 100)) AS n;

INSERT INTO orders (order_id, user_id, product_id, qty, placed_at)
SELECT n,
       MOD(n, 500) + 1,
       MOD(n, 100) + 1,
       MOD(n, 5) + 1,
       TIMESTAMP_ADD(TIMESTAMP '2024-01-01 00:00:00 UTC', INTERVAL MOD(n, 86400) SECOND)
FROM UNNEST(GENERATE_ARRAY(1, 5000)) AS n;
-- @query
SELECT u.region,
       COUNT(DISTINCT u.id)                          AS active_users,
       SUM(p.price * o.qty)                          AS revenue,
       ARRAY_AGG(STRUCT(p.sku AS sku, o.qty AS qty)
                 ORDER BY p.price * o.qty DESC LIMIT 3) AS top_lines
  FROM orders   AS o
  JOIN users    AS u ON u.id = o.user_id
  JOIN products AS p ON p.product_id = o.product_id
 WHERE o.placed_at >= TIMESTAMP '2024-02-01 00:00:00 UTC'
 GROUP BY u.region
HAVING SUM(p.price * o.qty) > 0
 ORDER BY revenue DESC

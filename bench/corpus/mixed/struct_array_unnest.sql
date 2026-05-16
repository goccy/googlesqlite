CREATE TABLE events (
  id INT64,
  payload STRUCT<user STRUCT<id INT64, name STRING>, items ARRAY<STRUCT<sku STRING, qty INT64>>>
);
INSERT INTO events (id, payload)
SELECT n,
       STRUCT(
         STRUCT(n AS id, CONCAT('user_', CAST(n AS STRING)) AS name) AS user,
         [STRUCT(CONCAT('sku-', CAST(MOD(n, 5) AS STRING)) AS sku, MOD(n, 7) + 1 AS qty),
          STRUCT(CONCAT('sku-', CAST(MOD(n + 1, 5) AS STRING)) AS sku, MOD(n + 1, 7) + 1 AS qty)] AS items
       )
FROM UNNEST(GENERATE_ARRAY(1, 500)) AS n;
-- @query
SELECT e.id,
       e.payload.user.name AS user_name,
       SUM(item.qty)       AS total_qty
  FROM events e, UNNEST(e.payload.items) AS item
 GROUP BY e.id, user_name
 ORDER BY e.id
 LIMIT 50

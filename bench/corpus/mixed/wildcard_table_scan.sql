CREATE TABLE `proj.ds.events_20240101` (id INT64, name STRING);
CREATE TABLE `proj.ds.events_20240102` (id INT64, name STRING);
CREATE TABLE `proj.ds.events_20240103` (id INT64, name STRING);

INSERT INTO `proj.ds.events_20240101` (id, name)
SELECT n, CONCAT('a_', CAST(n AS STRING)) FROM UNNEST(GENERATE_ARRAY(1, 500)) AS n;
INSERT INTO `proj.ds.events_20240102` (id, name)
SELECT n + 500, CONCAT('b_', CAST(n AS STRING)) FROM UNNEST(GENERATE_ARRAY(1, 500)) AS n;
INSERT INTO `proj.ds.events_20240103` (id, name)
SELECT n + 1000, CONCAT('c_', CAST(n AS STRING)) FROM UNNEST(GENERATE_ARRAY(1, 500)) AS n;
-- @query
SELECT _TABLE_SUFFIX AS day, COUNT(*) AS row_count
  FROM `proj.ds.events_*`
 GROUP BY day
 ORDER BY day

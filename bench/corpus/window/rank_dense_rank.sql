CREATE TABLE scores (cls INT64, person INT64, score INT64);
INSERT INTO scores (cls, person, score)
  SELECT MOD(n, 4), n, MOD(n * 17, 101)
  FROM UNNEST(GENERATE_ARRAY(1, 1000)) AS n;
-- @query
SELECT cls, person, score,
       RANK() OVER (PARTITION BY cls ORDER BY score DESC) AS r,
       DENSE_RANK() OVER (PARTITION BY cls ORDER BY score DESC) AS dr
FROM scores
ORDER BY cls, score DESC, person

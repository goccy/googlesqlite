CREATE TABLE txn (acct INT64, ts INT64, amt INT64);
INSERT INTO txn (acct, ts, amt)
  SELECT MOD(n, 8), n, n
  FROM UNNEST(GENERATE_ARRAY(1, 2000)) AS n;
-- @query
SELECT acct, ts, amt,
       SUM(amt) OVER (PARTITION BY acct ORDER BY ts ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running_total
FROM txn
ORDER BY acct, ts

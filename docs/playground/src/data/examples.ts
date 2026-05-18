// Example queries shown in the editor's "Examples" menu. They are a
// quick tour of GoogleSQL for first-time visitors.

export interface Example {
  label: string
  sql: string
}

export const EXAMPLES: Example[] = [
  {
    label: 'Create a table',
    sql: `CREATE TABLE users (id INT64, name STRING, score FLOAT64);

INSERT INTO users (id, name, score) VALUES
  (1, 'Alice', 92.5),
  (2, 'Bob', 78.0),
  (3, 'Carol', 88.5);

SELECT * FROM users ORDER BY score DESC;`,
  },
  {
    label: 'Arrays & UNNEST',
    sql: `SELECT n, n * n AS squared
FROM UNNEST([1, 2, 3, 4, 5, 6, 7, 8]) AS n
WHERE MOD(n, 2) = 0;`,
  },
  {
    label: 'STRUCT & ARRAY',
    sql: `SELECT
  STRUCT('Tokyo' AS city, 13960000 AS population) AS place,
  ARRAY_LENGTH(['a', 'b', 'c']) AS letters;`,
  },
  {
    label: 'Window function with QUALIFY',
    sql: `WITH sales AS (
  SELECT 'A' AS product, 100 AS amount UNION ALL
  SELECT 'B', 250 UNION ALL
  SELECT 'C', 175 UNION ALL
  SELECT 'D', 300
)
SELECT product, amount,
       RANK() OVER (ORDER BY amount DESC) AS sales_rank
FROM sales
QUALIFY sales_rank <= 3;`,
  },
  {
    label: 'Dates',
    sql: `SELECT
  d,
  FORMAT_DATE('%Y-%m-%d', d) AS formatted,
  EXTRACT(DAYOFWEEK FROM d) AS weekday
FROM UNNEST(GENERATE_DATE_ARRAY('2024-01-01', '2024-01-07')) AS d;`,
  },
  {
    label: 'JSON',
    sql: `SELECT
  JSON_VALUE(j, '$.name') AS name,
  JSON_QUERY(j, '$.tags') AS tags
FROM UNNEST([
  JSON '{"name": "Alice", "tags": ["a", "b"]}',
  JSON '{"name": "Bob", "tags": ["c"]}'
]) AS j;`,
  },
]

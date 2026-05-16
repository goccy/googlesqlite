---
name: AGG
dialect: bigquery
category: functions/aggregate
status: implemented
notes: |
  Fully implemented end-to-end:
  - Analyzer recognises AGG(MEASURE<T1>) via FEATURE_ENABLE_MEASURES.
  - CREATE PROPERTY GRAPH builds a go-googlesql SimplePropertyGraph
    (internal/property_graph.go); measure-typed property declarations
    wrap their inner type via TypeFactory.MakeMeasureType so
    `l.<measure_prop>` types as MEASURE<T>.
  - Graph query syntax (GRAPH ... MATCH ... RETURN) is lowered to
    SQL by internal/graph_scan_node.go. The upstream measure
    rewriter is disabled in our analyzer options because it only
    accepts AGG over a direct ColumnRef; graph queries instead
    produce AGG(GraphGetElementProperty), which the formatter
    handles via tryFormatMeasureAGG by emitting
    googlesqlite_agg(<inner_expr>, <locking_key>, '<kind>').
  - Runtime AGG aggregator lives at internal/functions/aggregate/agg.go
    and performs locking-key dedup + SUM / AVG / MIN / MAX / COUNT /
    COUNT_DISTINCT / ANY_VALUE finalisation.
source_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#agg
upstream_url: https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#agg
last_synced: 2026-05-04
testdata: testdata/specs/bigquery/functions/aggregate/agg.yaml
---

# AGG

## Summary
Aggregates a `MEASURE`-typed expression by invoking the calculation that
the measure encapsulates exactly once per its locking key, so that
group-level results never overcount values that are repeated by joins
on a property graph. Used to apply business-metric measures defined on
graph node and edge tables.

## Signatures
- `AGG(measure_expression)`

## Arguments
- `measure_expression`: a single value of `MEASURE` type. A measure
  bundles an aggregate calculation together with a key that defines its
  granularity.

## Return type
The data type returned by the aggregate calculation that the measure
expression wraps (for example, `INT64` when the measure is built from
`SUM(population)` over an `INT64` column).

## Behavior
- The argument must be of `MEASURE` type; ordinary scalar columns are
  not accepted.
- The measure is evaluated once per distinct value of the key declared
  when the measure was defined, so duplicated rows produced by graph
  expansion do not contribute multiple times.
- `AGG` is the only way to read a property whose definition is a
  `MEASURE`; selecting that property directly without `AGG` is not
  permitted, and such columns are typically excluded from `SELECT *`
  results.
- `AGG` is intended for measures exposed by property graphs accessed
  through the `GRAPH_EXPAND` table-valued function.
- This feature is in Preview and is subject to the Pre-GA Offerings
  Terms; behaviour and availability may change before general
  availability.

## Examples
```sql
-- Define a property graph whose Locations node table exposes a
-- measure called total_population, defined as SUM(population) keyed
-- by id.
CREATE OR REPLACE TABLE mydataset.Stores (
  name STRING PRIMARY KEY NOT ENFORCED,
  location_id INT64 REFERENCES mydataset.Locations(id) NOT ENFORCED
) AS (
  SELECT 'Store 1' AS name, 101 AS location_id
  UNION ALL
  SELECT 'Store 2' AS name, 101 AS location_id
);

CREATE OR REPLACE TABLE mydataset.Locations (
  id INT64 PRIMARY KEY NOT ENFORCED,
  name STRING,
  population INT64
) AS (
  SELECT 101 AS id, 'Anytown' AS name, 1000 AS population
);

CREATE OR REPLACE PROPERTY GRAPH mydataset.StoreGraph
  NODE TABLES (
    mydataset.Stores AS S,
    mydataset.Locations AS L
    PROPERTIES(id, name, population, MEASURE(SUM(population)) AS total_population)
  )
  EDGE TABLES (
    mydataset.Stores AS SL
    SOURCE KEY (location_id) REFERENCES L (id)
    DESTINATION KEY (name) REFERENCES S (name)
  );

-- GRAPH_EXPAND flattens the graph; measure-typed columns such as
-- L_total_population can't be selected directly and are dropped here.
SELECT * EXCEPT(L_total_population)
FROM GRAPH_EXPAND('mydataset.StoreGraph');
-- expected rows: (101, 'Store 2', 101, 'Anytown', 1000),
--                (101, 'Store 1', 101, 'Anytown', 1000)

-- AGG counts each location's population once per distinct location_id,
-- whereas SUM over the joined L_population overcounts duplicated rows.
SELECT
  S_location_id,
  AGG(L_total_population) AS true_total_population,
  SUM(L_population)       AS overcounted_population
FROM GRAPH_EXPAND('mydataset.StoreGraph')
GROUP BY S_location_id;
-- expected: (101, 1000, 2000)
```

## Edge cases
- Calling `AGG` on a non-`MEASURE` expression is invalid.
- Selecting a measure-typed column without wrapping it in `AGG` is not
  allowed; tools that expand `*` typically must `EXCEPT` such columns.
- Because measures must be defined on a property graph, `AGG` is only
  meaningful in queries that read from `GRAPH_EXPAND` (or otherwise
  surface the underlying `MEASURE` type).
- Pre-GA: support is provided on an "as is" basis and the function is
  not yet covered by a general-availability service-level commitment.

## Reference (upstream)

See the upstream BigQuery documentation for the authoritative text:
<https://cloud.google.com/bigquery/docs/reference/standard-sql/aggregate_functions#agg>.

# String literal coercion to DATE / TIMESTAMP

## Status

Accepted — matches the GoogleSQL specification.

## Context

While reviewing the analyzer's behavior on operator overloads, three
queries were flagged as potentially-erroneous candidates:

```sql
SELECT '2022-09-09' - INTERVAL 1 DAY;
SELECT '2022-09-09' > TIMESTAMP '2022-09-09 00:00:00+00';
SELECT CAST('2022-09-09' AS DATE) - INTERVAL 1 DAY;
```

Both `googlesqlite` (via `go-googlesql` v0.2.1) and the predecessor
the upstream wasm bridge accept all three:

| Query                                                            | Result                  |
|------------------------------------------------------------------|-------------------------|
| `'2022-09-09' - INTERVAL 1 DAY`                                  | `2022-09-08T00:00:00`   |
| `'2022-09-09' > TIMESTAMP '2022-09-09 00:00:00+00'`              | `true`                  |
| `CAST('2022-09-09' AS DATE) - INTERVAL 1 DAY`                    | `2022-09-08T00:00:00`   |

The two implementations agree, so there is no implementation
divergence. The original concern was whether the analyzer ought to
reject the first two as type errors.

## Decision

This behavior is correct per the GoogleSQL specification. From
`docs/third_party/googlesql-docs/conversion_rules.md` (the upstream
`Literal coercion` section):

> Literal coercion is needed when the actual literal type is different
> from the type expected by the function in question. For example, if
> function `func()` takes a `DATE` argument, then the expression
> `func("2014-09-27")` is valid because the string literal
> `"2014-09-27"` is coerced to `DATE`.

The arithmetic and comparison operators in question expect
`DATE` / `TIMESTAMP` operands, so the string literals coerce per the
documented rule. The analyzer's behavior matches the spec.

No upstream issue is filed against `go-googlesql`. No strict-mode
toggle is added to `googlesqlite`.

## Notes

- The first and third queries return a value formatted as
  `2022-09-08T00:00:00` rather than a bare `2022-09-08`. This is the
  internal `DATE` rendering used by both implementations and is
  consistent across the predecessor and the current driver — it is not
  a regression introduced by the `go-googlesql` v0.1.0 → v0.2.1 bump.
- Coercion only applies to **literals**. Per the same section, query
  parameters and column references with `STRING` type are not silently
  coerced to `DATE` / `TIMESTAMP` for arithmetic; an explicit `CAST`
  is required.

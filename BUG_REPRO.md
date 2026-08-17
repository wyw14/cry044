# Bug reproduction

## Bug

Statistics for a requested `from`/`to` window do not consistently apply that
window to return reasons and review scores. Historical batches outside the
window can be included, while current in-window returns can be omitted. The
in-memory and PostgreSQL implementations can consequently report different
results.

## Trigger

1. Create an older returned batch and score it outside the target window.
2. Create a newer return and score inside the target window.
3. Request statistics for the newer window and run the focused regression test:

```text
go test ./internal/repository -run TestStatisticsWindowExcludesStaleReturnsAndScores -count=20
```

## Error / observed result

The response contains stale return reasons or scores from the older batch (and
may miss the in-window return). Identical `from`/`to` requests can therefore
produce inconsistent in-memory versus PostgreSQL statistics.

# Bug reproduction

- Bug: scores outside a criterion's configured scale enter aggregation.
- Trigger: submit a score below `Scale.Min` or above `Scale.Max`.
- Error: aggregation accepts the invalid score and produces a result.

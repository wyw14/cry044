# Bug reproduction

## Bug

When two different JSON template drafts are imported consecutively without an
`Idempotency-Key`, the application reuses one fixed repository request key.
The second, distinct draft is therefore treated as a duplicate and the first
template is returned again instead of creating an independent template.

## Trigger

1. Start from the repository at the bugfix branch baseline.
2. Submit two template-import requests with different JSON payloads and omit
   the `Idempotency-Key` header from both requests.
3. Observe the result of the second import (or run the focused regression test):

```text
go test ./internal/application -run TestImportWithoutHeaderDoesNotCollapseDistinctPayloads -count=20
```

## Error / observed result

The second import is reported as successful but returns the first template's
identity/content. Distinct drafts are collapsed into one stored template,
causing the newly imported draft to be silently lost.

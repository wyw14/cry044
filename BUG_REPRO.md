# Bug reproduction

## Bug

Template comparison only considers an item's title and weight. Changes to
decision semantics—scoring scale, labels, veto/required flags, applicability
conditions, or variables—are omitted from the diff, so a materially different
approval rule can be reported as unchanged.

## Trigger

1. Create two versions of the same review template and keep the item ID
   unchanged.
2. Change only one decision-semantic field (for example the scoring scale,
   veto flag, required/optional setting, conditions, or variables).
3. Compare the two versions with the focused regression command:

```text
go test ./internal/domain -run TestTemplateComparisonDetectsDecisionSemanticChanges -count=20
```

## Error / observed result

The comparison returns an empty or incomplete change list (often “no change”)
even though the decision outcome can change. Reviewers can therefore miss a
meaningful approval-rule modification in publication review and history.

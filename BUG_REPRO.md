# Bug reproduction

- Bug: a template snapshot aliases mutable maps and slices from its source.
- Trigger: create a snapshot, then edit the source variables or criterion conditions.
- Error: the historical snapshot changes after the source template is edited.

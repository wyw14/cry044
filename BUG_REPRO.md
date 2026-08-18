# Bug reproduction

- Bug: one reviewer can submit the same material more than once in a batch.
- Trigger: submit the same batch, material, and reviewer twice; both calls succeed.
- Error: the second submission is accepted instead of returning the review conflict.

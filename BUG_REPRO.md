# Bug reproduction

- Bug: publishing with an old template revision is silently normalized.
- Trigger: publish a review-stage template with an expected revision of zero.
- Error: the stale request publishes instead of returning a revision conflict.

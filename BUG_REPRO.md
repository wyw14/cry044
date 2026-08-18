# Bug reproduction

- Bug: local exchange accepts a path-traversal filename.
- Trigger: save a JSON file named `../outside.json`.
- Error: content is written outside the configured local file root instead of being rejected.

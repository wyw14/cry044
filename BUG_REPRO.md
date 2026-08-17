# 并发最终裁决丢失更新

同一批次的两个协调员使用相同 `revision` 并发提交材料最终结论时，两个请求都会成功，后一份结论会覆盖前一份。

复现：

```bash
go test ./tests -run TestConcurrentFinalDecisionUsesOneBatchRevision -count=20
```

错误信息稳定为 `expected exactly one concurrent decision to commit, got 2`。

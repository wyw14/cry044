# 打印评议单丢失席位和分歧信息

完成批次生成可打印评议单时，聚合得分仍在，但评审人数被重置为 0，分歧条目也从导出数据中消失。

复现：

```bash
go test ./internal/application -run TestPrintableSheetKeepsPanelCountsAndDisagreements -count=20
```

失败输出包含 `printable data lost panel detail`，并显示 `ReviewerCount:0`、空 `Disagreements`。

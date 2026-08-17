# 审核标准模板复用与评议平台

这是一个完全离线运行的 Go + Vue 3 全栈示例，用来治理审核标准的克隆、参数化、发布、批次快照、多评审聚合、分歧议事和历史回溯。批次启动时封存模板快照；后续模板版本变化不会污染进行中的评议。否决项优先于总分，存在否决或显著分歧时必须由协调人形成带依据的最终结论。

## 目录职责

- `cmd/server`：依赖装配、HTTP 生命周期与优雅停机。
- `internal/domain`：模板版本、批次状态机、聚合、法定人数、分歧和最终结论规则。
- `internal/application`：模板克隆/发布、评议提交、议事单、导入导出和统计用例。
- `internal/repository`：并发安全内存实现和 pgx/PostgreSQL 实现。
- `internal/platform/localfiles`：受控本地文件保存、扩展名/大小/目录穿越校验和 SHA-256。
- `internal/transport/http`：`/api/v1` 路由、角色门禁、输入校验和统一错误包络。
- `migrations`：可重复执行的表结构与 `ON CONFLICT` 演示数据。
- `web`：标准谱系、模板对比、评议席位、分歧议事、批次历史、发布审批与受控交换页。

## 本地启动

要求 Go 1.24+、Node.js 20+、PostgreSQL 16+。复制 `.env.example` 后执行：

```powershell
$env:DATABASE_URL='postgres://review:review@localhost:5432/review?sslmode=disable'
./scripts/migrate.ps1
go run ./cmd/server
cd web
npm install
npm run dev
```

也可以执行 `docker compose up --build`，API 位于 `http://localhost:8080`。`/healthz` 只表示进程存活，`/readyz` 会在一秒超时内探测 PostgreSQL。

## 配置与本地文件

`HTTP_ADDR` 默认 `:8080`，`DATABASE_URL` 默认指向本地数据库，`UPLOAD_DIR` 默认 `./var/exchange`。导入限制为 2 MiB JSON；导出只写入受控目录，拒绝危险扩展名和目录穿越。密码、令牌、附件正文和评审证据正文不会写日志。

## 演示数据与状态规则

迁移会幂等创建“生产质量评议 V1”发布模板、一个已启动批次和两席示例评审。模板状态为 `draft → review → published → deprecated`；已发布/弃用版本不可原地改写，被批次引用的历史版本不可删除。批次状态为 `draft → submitted → returned → re_review → completed`，`submitted/returned/re_review` 均可进入 `voided`。退回必须填写原因。

最终结论还遵守三条规则：未满足最少两席不可裁决；评分分差超过阈值时必须记录议事依据；任一否决项触发后不能用高总分改判通过。

## API 示例

```bash
curl -H "X-Actor-ID: chair-1" -H "X-Actor-Role: coordinator" \
  http://localhost:8080/api/v1/review-batches/batch-1/deliberation

curl -X POST -H "Content-Type: application/json" \
  -H "X-Actor-ID: chair-1" -H "X-Actor-Role: coordinator" \
  -d '{"expected_revision":2,"passed":false,"conclusion":"退回","rationale":"维持安全否决"}' \
  http://localhost:8080/api/v1/review-batches/batch-1/materials/material-1/final-decisions
```

错误响应始终包含稳定 `code`、`message`、`field_errors` 与 `request_id`。完整接口见 `api/openapi/openapi.yaml`。

## 验证

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
cd web && npm test
cd web && npm run build
```

基线实际验证结果：上述 Go 构建、测试、race、vet，以及前端测试和生产构建均通过。仓库不提交 `.env`、依赖缓存、`web/dist`、`var` 或运行期数据库文件。

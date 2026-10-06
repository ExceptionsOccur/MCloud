# 任务队列（ROADMAP）

> **本文件是任务状态的唯一事实来源（source of truth）。** [AGENTS.md](../AGENTS.md) 中的「状态快照」只是它的缓存，两者不一致时**以本文件为准**，并修复 AGENTS.md。
> 执行协议（如何领任务、如何定位代码、如何回写）见 [AGENTS.md · 会话协议](../AGENTS.md#会话协议)。

## 字段约定（AI 必读）

| 字段 | 取值规则 |
|------|----------|
| ID | `T-001` 形式，全局唯一，**永不复用**（done 后也不再启用） |
| 状态 | 仅允许 `todo` / `in_progress` / `blocked` / `done` 四个值，状态变化只在本文件发生 |
| 负责 | 进入 `in_progress` 时必须填执行者标识（AI 会话名或人名）+ 认领日期 |
| 优先级 | `P0` 阻塞他人 > `P1` 当前迭代 > `P2` 可延后；同级按 ID 顺序 |
| 验收标准 | 客观可判定（命令通过 / 文件存在 / 行为可复现），禁止主观描述 |
| 定位 | 涉及文件清单，来源 `docs/CODE_INDEX.md`「功能→代码映射」；**领任务后必须核对/补齐**，行号仅为快照（以 grep 复核为准），补不出先声明、不盲写 |
| 回写 | 完成后必须更新的文档（通常由 AGENTS.md 责任矩阵推导，此处列出例外） |

**认领规则**：一次只领 **1 条** `P0/P1` 任务；领用 = 把该条移到 `in_progress` 并填写负责字段（在写任何代码之前先改本文件）。

---

## in_progress

- [ ] **T-018** 后端 lint 清零 + golangci-lint v2 配置迁移 ｜ P1 ｜ 负责: AI/opencode ｜ 认领: 2026-10-06
  - 背景：本机 golangci-lint 为 **v2.14.0**，而 `backend/.golangci.yml` 仍是 v1 格式 → `run` 直接 `exit 3`（配置未被读取）；用 v2 重跑项目既有规则集实得 **20 个问题**
  - 内容：① `golangci-lint migrate` 把 `.golangci.yml` 升到 v2（gosimple 并入 staticcheck、gofmt/goimports 移入 `formatters`）② 修完 20 个问题且**不改对外 API 行为**
  - 定位：`backend/.golangci.yml`；errcheck —— `config/config.go:26`、`database/postgres.go:20`（`godotenv.Load`）、`controllers/websocket.go:76,80`（`json.Marshal`/`WriteMessage`）、`services/host_service.go:705-708`（4×`strconv.Atoi`）、`utils/csv.go:82,84`（`writer.Write`）；gofmt —— `config/config.go`、`models/user.go`、`models/host.go`、`models/cloud_resource.go`、`services/host_service.go:197`；goimports —— `services/stats_service.go:6`；ineffassign —— `controllers/response.go:24`；staticcheck QF —— `services/business_stats.go:70`、`services/host_service.go:128,710`
  - 验收标准：`cd backend && go build ./...` 与 `PATH=$PATH:$(go env GOPATH)/bin golangci-lint run ./...` 均 exit 0；CSV 导入导出与 WebSocket 探测行为不变；`docs/DEVELOPMENT.md` Linter 章节的规则集与 v2 配置一致
  - 回写：PROJECT_STATUS 变更记录 + 技术债表（若有）、DEVELOPMENT Linter 章节、AGENTS 状态快照 `last_updated`
  - 分支：`chore/lint-v2`

## todo（按优先级）

- [ ] **T-002** 后端 `services/host_service.go` 拆分 ｜ P1 ｜ 负责: —
  - 依赖：T-018（先清零 lint 再拆分，避免格式/errcheck 修改与拆分 diff 冲突）
  - 内容：拆为 `host_service.go`（CRUD/筛选）+ `host_batch_service.go`（文本/结构化批量）+ `host_csv_service.go`（行映射、导入导出）
  - 定位：`backend/services/host_service.go`（18 个函数，22KB）；回写 `docs/CODE_INDEX.md` 服务索引
  - 验收标准：`cd backend && go build ./...` 与 `golangci-lint run ./...` 通过；**对外 API 行为零变化**；`docs/CODE_INDEX.md` 服务索引同步
  - 分支：`refactor/split-host-service`
- [ ] **T-003** `controllers/csv.go` 业务逻辑下沉 ｜ P1 ｜ 负责: —
  - 内容：行数校验、表头判断、去重等逻辑移入 service，controller 只剩绑定+调用+响应（红线 1）
  - 定位：`backend/controllers/csv.go`（L34 文件后缀、L61 行循环、L65 行数校验、L78 重复判断）→ 下沉至 `services/host_service.go` 的 CSV 部分
  - 验收标准：controller 中无业务校验代码；导入导出行为不变；`go build` + lint 通过
  - 注意：与 T-002 共享 `host_service.go`，建议排在 T-002 之后或同分支执行
  - 分支：`refactor/csv-controller-thin`
- [ ] **T-004** 前端 composables 抽取 ｜ P1 ｜ 负责: —
  - 内容：新建 `frontend/src/composables/`，抽 `useProbeWebSocket` / `useChartOption` / `usePagedTable`
  - 定位：抽取源 —— `views/IpStatistics.vue`（WebSocket+探测）、`views/ResourceStatistics.vue` / `views/BusinessStatistics.vue`（ECharts 配置）、`views/HostManagement.vue` + `stores/host.js`（分页）
  - 验收标准：`cd frontend && npm run lint && npm run build` 通过；抽离后各页面行为不变
- [ ] **T-005** `IpStatistics.vue`（746 行）拆分 ｜ P1 ｜ 负责: —
  - 依赖：T-004
  - 内容：拆出 `components/ip/`（网格局部、探测面板、网段管理），单文件 ≤ 400 行
  - 定位：`frontend/src/views/IpStatistics.vue`（template 1-96 / script 98-432 / 全文 746 行）→ 新建 `frontend/src/components/ip/`
  - 验收标准：无单个 `.vue` 超过 500 行；`npm run lint && npm run build` 通过；功能无回归
- [ ] **T-006** 测试骨架 ｜ P1 ｜ 负责: —
  - 内容：后端 `services` 表驱动单测（先覆盖主机 CRUD、登录）+ 前端 Vitest 冒烟
  - 定位：新建 `backend/services/*_test.go`；前端 `frontend/package.json`（加 vitest 依赖与 script）+ `src/**/__tests__/`
  - 验收标准：`cd backend && go test ./...` 有用例非零且通过；`npm run test` 可运行
- [ ] **T-007** CI 流水线落地 ｜ P1 ｜ 负责: —
  - 依赖：无（可与 T-006 并行）
  - 内容：`.github/workflows/ci.yml` 执行 `go build` / `golangci-lint` / `npm run lint` / `npm run build`
  - 定位：新建 `.github/workflows/ci.yml`（`.github/` 目前不存在）；回写 `docs/DEVELOPMENT.md` CI 章节
  - 验收标准：workflow 文件存在且语法有效；DEVELOPMENT.md 删除「CI 尚未落地」警告
  - 回写：`docs/DEVELOPMENT.md` CI 章节、PROJECT_STATUS 技术债表
- [ ] **T-008** 数据导出报表（Excel/CSV） ｜ P2 ｜ 负责: —
  - 内容：报表导出接口与前端入口
  - 定位：新建 `backend/services/report_service.go` + `controllers/report.go`，注册于 `routes/routes.go`；前端入口 `components/SearchToolbar.vue` / `views/HostManagement.vue`；CSV 工具复用 `utils/csv.go`（BOM 红线）
  - 验收标准：新 API 记入 `docs/API.md`；导出文件带 BOM 可被 Excel 正确打开
- [ ] **T-009** 密码哈希升级 bcrypt ｜ P2 ｜ 负责: —
  - 内容：慢哈希替换 SHA-256，登录支持 `salt$hash` 旧格式平滑迁移
  - 定位：`backend/utils/password.go`（`HashPassword`/`VerifyPassword`）、`services/auth_service.go`、`database/postgres.go`（`seedAdmin`）+ 迁移 SQL
  - 验收标准：旧密码用户登录成功后自动升级为新格式；迁移 SQL 写入 `backend/migrations/`
  - 回写：ARCHITECTURE 数据模型、BUSINESS_LOGIC 认证流程、README 技术栈
- [ ] **T-010** `routes/routes.go` 按域拆分 ｜ P2 ｜ 负责: —
  - 定位：`backend/routes/routes.go`（单文件 100+ 行，JWT 逐组挂载）→ 按域拆 `routes/*.go`，JWT 挂载收口
  - 验收标准：路由按域拆为多文件；`docs/CODE_INDEX.md` 请求链路同步
- [ ] **T-011** IP 探测去宿主机 `ping` 依赖 ｜ P2 ｜ 负责: —
  - 内容：评估纯 Go ICMP/TCP 方案，消除平台差异
  - 定位：`backend/services/stats_service.go` L100（`exec.CommandContext(ctx, "ping", ...)`）、`pingICMP`/`probeTCP`；调用链 `controllers/stats.go` → `Probe`、`controllers/websocket.go`
  - 验收标准：探测颜色状态机行为与 `docs/BUSINESS_LOGIC.md` 描述一致
  - 回写：BUSINESS_LOGIC「探测方式」
- [ ] **T-012** 文档一致性巡检 ｜ P2 ｜ 负责: — ｜ 备注: 适合小型会话
  - 定位：`docs/*.md`、`README.md`、`AGENTS.md`（协议条文部分需人类授权，见红线 7）
  - 验收标准：全部 md 相对链接/锚点可解析；目录树与实际文件一致；发现的错误全部修复
- [ ] **T-013** 零信任台账 ｜ P2 ｜ 负责: — ｜ 备注: 字段清单已由人类确认
  - 内容：新增零信任接入申请台账，字段：**申请单位、账户名、申请人联系方式、申请主机、申请端口、申请时间、密码（可选）、备注**；后端 models + 迁移 SQL + CRUD API + 前端管理页
  - 定位：后端新建 `models/zero_trust.go`、`services/zero_trust_service.go`、`controllers/zero_trust.go`，注册 `routes/routes.go`，迁移 SQL 入 `migrations/`；前端新建 `views/ZeroTrustLedger.vue`、`api/zero_trust.js`、`router/index.js` 加路由（入口模式参照 `CloudResourceDialog` 或顶部 Tab 二选一，实现时定）
  - 验收标准：8 个字段齐全可增删改查；`go build` + `golangci-lint` + 前端 `npm run lint` 通过；表结构记入 `docs/ARCHITECTURE.md`、接口记入 `docs/API.md`
  - 进入条件（实现前须人类确认）：① 密码字段的存储方式（敏感数据，建议加密/哈希，不落明文）② 「申请主机」是否为 `hosts.id` 外键
  - 回写：ARCHITECTURE.md、API.md、CODE_INDEX.md、PROJECT_STATUS 完成度
  - 分支：`feat/zero-trust-ledger`
- [ ] **T-014** 域名台账 ｜ P2 ｜ 负责: — ｜ 备注: **字段需求待人类细化**
  - 内容：新增域名台账（字段清单未定；参考项：域名、解析公网 IP、服务商、到期时间、备注）
  - 定位：与 T-013 同构 —— 后端 `models/domain.go`、`services/domain_service.go`、`controllers/domain.go` + `routes/routes.go`；前端 `views/DomainLedger.vue`、`api/domain.js`、`router/index.js`
  - 验收标准：字段清单经人类确认后方可进入 `in_progress`；其余同 T-013
  - 回写：ARCHITECTURE.md、API.md、CODE_INDEX.md、PROJECT_STATUS 完成度
  - 分支：`feat/domain-ledger`
- [ ] **T-015** 公网 IP 字段改造 + 双台账关联 ｜ P2 ｜ 依赖: T-013、T-014 ｜ 负责: —
  - 内容：按人类指定方案，`hosts.public_ip`（`varchar(45)`）改为**布尔值「是否做了映射」**；主机记录展示与零信任台账、域名台账的关联状态
  - 定位（行号为 2026-10-06 快照，以 grep `public_ip` 复核）：后端 `models/host.go:11`、`services/host_service.go`（L201/243/301/358/472/607/720/753 共 8 处）、`utils/csv.go:15`（表头「公网IP」）、迁移 SQL；前端 `components/HostTable.vue:28`、`components/HostFormDialog.vue:100/209/355`
  - 验收标准：迁移 SQL 写入 `backend/migrations/`（含改名/改类型）；API 响应、主机表单、CSV 导入导出、筛选同步更新；`docs/ARCHITECTURE.md` hosts 表与 `docs/API.md` 同步；前后端验证命令全绿
  - 进入条件（实现前须人类确认）：① 新字段名（如 `ip_mapped`）② 布尔值语义（"已映射"是否区分映射到哪个台账）③ 旧 `public_ip` 数据的迁移/丢弃策略
  - 回写：ARCHITECTURE.md、API.md、BUSINESS_LOGIC.md（如涉及统计口径）、PROJECT_STATUS 变更记录
  - 分支：`feat/host-ip-mapping-flag`
- [ ] **T-016** Goose 迁移执行器实装 ｜ P2 ｜ 负责: — ｜ 备注: 2026-10-06 人类确认**暂不实装，仅入队**；现状为归档-only 口径
  - 内容：引入 `github.com/pressly/goose/v3`，启动 `Migrate()` 时执行 `migrations/*.sql`（`goose.Up`），`AutoMigrate` 降级为兜底或移除；存量 8 个 SQL 与现有库结构一致性验证
  - 定位：`backend/go.mod`（当前无 goose 依赖）、`backend/database/postgres.go` → `Migrate()`（L39-54，现仅 AutoMigrate+seedAdmin）、`backend/migrations/`（8 个 `-- +goose` SQL，**从未执行过**）、如用 CLI 则涉 `backend/run.sh` / `Dockerfile`
  - 进入条件（实现前须人类确认）：① 存量 dev/prod 库结构与 SQL 文件是否一致（不一致需先补差量 SQL）② AutoMigrate 的去留 ③ seed_admin.sql 与 `seedAdmin()` 的职责划分
  - 验收标准：新库从零启动仅靠 SQL 建表成功；重复启动幂等；`go build` + lint 通过
  - 回写（口径反转，必做）：DEVELOPMENT.md 迁移规范口径表、AGENTS.md 红线 9、PROJECT_STATUS 注意事项 8、ARCHITECTURE 目录树注释——改回"运行时执行 SQL"
  - 分支：`feat/goose-migrator`

## blocked

（无。填写格式：`- [ ] **T-xxx** … ｜ blocked 原因：具体条件` —— 泛泛的"困难"不算 blocked）

## done

| ID | 任务 | 完成时间 | 提交 | 备注 |
|----|------|----------|------|------|
| T-001 | 协作文档体系改造（AGENTS 改为 AI 入口、ROADMAP 改为任务队列、删除 ONBOARDING、分支策略对齐 main+feature） | 2026-10-06 | `0237d40` | 验收：无 ONBOARDING 链接残留、AGENTS 含状态快照/会话协议/验证命令、md 链接锚点全通 |
| — | 2026-09 及之前的交付项 | — | 见 `git log` | 详细记录见 [PROJECT_STATUS.md](./PROJECT_STATUS.md#近期变更记录) |

> 归档规则：done 条目保留 3 个月后可移入上方汇总行；**ID 不复用**。

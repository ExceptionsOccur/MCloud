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

- [ ] **T-017** 文档事实修正与协议补强 ｜ P1 ｜ 负责: AI/opencode ｜ 认领: 2026-10-06
  - 背景：2026-10-06 全量评估发现的文档漂移与协议缺口，校验器 `scripts/check_docs.sh`（T-021）已能机器化定位其中的计数/目录树/快照类问题
  - 内容：
    - **A 事实修正（7 项）**：① ARCHITECTURE 数据模型表数改 6 ② 目录树删 `backend/.env.example`（不存在）、补 `Dockerfile.dev`/`docker/.env.example`/`.golangci.yml`/`.eslintrc.cjs`/`package-lock.json` 并标树快照日期 ③ T-002 定位「18 个函数，22KB」实测复核（T-018 后为 18 函数，以实测为准）④ 3 处 SQL 计数 8→7（PROJECT_STATUS、ROADMAP×2）⑤ T-015 定位 `host_service.go` `public_ip` 处数与行号 grep 复核（原记 8 处，补 L100）⑥ README 技术栈「ORM/迁移」口径修正（SQL 文件为归档，运行时 AutoMigrate 兜底）⑦ T-012 验收计数一致性（T-021 已接线，复核即可）
    - **B 门禁与前置（3 项，协议条文，人类已授权）**：① 门禁命令改 `npm run lint:check` 并注明 `lint` 带 `--fix` 会改文件 ② AGENTS 验证命令章节加「工具前置」（Go 1.22+/Node 18+/`export PATH=$PATH:$(go env GOPATH)/bin`/golangci-lint v2 已迁移/docker 环境）③ DEVELOPMENT 同步
    - **C 快照可靠性（4 项，协议条文，人类已授权）**：① 快照新增 `next_task` 字段（= todo 首条 `T-002`）② 会话协议 A 步骤 1 加「比对快照与 ROADMAP（last_updated/todo/next_task），不一致先修快照」③ 回写协议第 4 步刷新对象明确含 `next_task`/`last_updated`/todo 队列 ④ 快照「功能完成度」行缩为一行摘要 + 指向 `PROJECT_STATUS.md#功能完成度`
    - **E 业务文档补缺（2 项）**：① `BUSINESS_LOGIC.md` 新增「人员管理」（重复判定/40901/手输自动建人员）② `API.md` 新增「WebSocket 探测帧协议」
    - **F 入队（2 项，仅登记不实现）**：`T-019` 抽 Layout/AppNav、`T-020` `config.go` `ServerPort` 硬编码违反红线 3
  - 定位：`docs/ARCHITECTURE.md`、`README.md:60`、`docs/ROADMAP.md`、`docs/PROJECT_STATUS.md`、`AGENTS.md`（快照+协议A/回写4+验证命令）、`docs/DEVELOPMENT.md`、`docs/BUSINESS_LOGIC.md`、`docs/API.md`；源码复核 `backend/services/host_service.go`（public_ip）、`backend/config/config.go` ServerPort、`frontend/src/views/*.vue`（nav-tab 重复）、`backend/services/person_service.go`（E1）、探测 WS handler（E2）
  - 验收标准：`bash scripts/check_docs.sh` **0 errors 0 warnings**；E1/E2 章节与代码事实一致；T-019/T-020 入队且 AGENTS 快照同步；协议条文改动仅限上述 B/C 列明项；`bash -n scripts/check_docs.sh` 通过
  - 回写：ROADMAP（T-017 done、T-019/T-020 入队）、PROJECT_STATUS 变更记录、AGENTS 快照（含 `next_task`）、责任矩阵指定文档
  - 分支：`docs/doc-drift-fix`

（认领规则：从 `todo` 取 1 条 P0/P1 移入此处）

## todo（按优先级）

- [ ] **T-002** 后端 `services/host_service.go` 拆分 ｜ P1 ｜ 负责: —
  - 依赖：T-018（先清零 lint 再拆分，避免格式/errcheck 修改与拆分 diff 冲突）
  - 内容：拆为 `host_service.go`（CRUD/筛选）+ `host_batch_service.go`（文本/结构化批量）+ `host_csv_service.go`（行映射、导入导出）
  - 定位：`backend/services/host_service.go`（18 个函数，22KB）；回写 `docs/CODE_INDEX.md` 服务索引
  - 验收标准：`cd backend && go build ./...` 与 `golangci-lint run ./...` 通过；**对外 API 行为零变化**；`docs/CODE_INDEX.md` 服务索引同步
  - 分支：`refactor/split-host-service`
- [ ] **T-003** `controllers/csv.go` 业务逻辑下沉 ｜ P1 ｜ 负责: —
  - 依赖：T-018（已完成 lint 清零，避免格式修改与下沉 diff 冲突）
  - 内容：行数校验、表头判断、去重等逻辑移入 service，controller 只剩绑定+调用+响应（红线 1）
  - 定位：`backend/controllers/csv.go`（L34 文件后缀、L61 行循环、L65 行数校验、L78 重复判断）→ 下沉至 `services/host_service.go` 的 CSV 部分
  - 验收标准：controller 中无业务校验代码；导入导出行为不变；`go build` + lint 通过
  - 注意：与 T-002 共享 `host_service.go`，建议排在 T-002 之后或同分支执行
  - 分支：`refactor/csv-controller-thin`
- [ ] **T-004** 前端 composables 抽取 ｜ P1 ｜ 负责: —
  - 内容：新建 `frontend/src/composables/`，抽 `useProbeWebSocket` / `useChartOption` / `usePagedTable`
  - 定位：抽取源 —— `views/IpStatistics.vue`（WebSocket+探测）、`views/ResourceStatistics.vue` / `views/BusinessStatistics.vue`（ECharts 配置）、`views/HostManagement.vue` + `stores/host.js`（分页）
  - 验收标准：`cd frontend && npm run lint:check && npm run build` 通过；抽离后各页面行为不变
- [ ] **T-005** `IpStatistics.vue`（746 行）拆分 ｜ P1 ｜ 负责: —
  - 依赖：T-004
  - 内容：拆出 `components/ip/`（网格局部、探测面板、网段管理），单文件 ≤ 400 行
  - 定位：`frontend/src/views/IpStatistics.vue`（template 1-96 / script 98-432 / 全文 746 行）→ 新建 `frontend/src/components/ip/`
  - 验收标准：无单个 `.vue` 超过 500 行；`npm run lint:check && npm run build` 通过；功能无回归
- [ ] **T-006** 测试骨架 ｜ P1 ｜ 负责: —
  - 内容：后端 `services` 表驱动单测（先覆盖主机 CRUD、登录）+ 前端 Vitest 冒烟
  - 定位：新建 `backend/services/*_test.go`；前端 `frontend/package.json`（加 vitest 依赖与 script）+ `src/**/__tests__/`
  - 验收标准：`cd backend && go test ./...` 有用例非零且通过；`npm run test` 可运行
- [ ] **T-007** CI 流水线落地 ｜ P1 ｜ 负责: —
  - 依赖：无（可与 T-006 并行）
  - 内容：`.github/workflows/ci.yml` 执行 `go build` / `golangci-lint` / `npm run lint:check`（CI 禁用带 `--fix` 的 `lint`）/ `npm run build` / `bash scripts/check_docs.sh`
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
  - 定位：`docs/*.md`、`README.md`、`AGENTS.md`（协议条文部分需人类授权，见红线 7）；校验器 `scripts/check_docs.sh`（T-021 已交付）
  - 验收标准：`bash scripts/check_docs.sh` 退出码 0（links/tree/snapshot/counts/shas 全绿）；warn 一并处理；发现的错误全部修复
- [ ] **T-013** 零信任台账 ｜ P2 ｜ 负责: — ｜ 备注: 字段清单已由人类确认
  - 内容：新增零信任接入申请台账，字段：**申请单位、账户名、申请人联系方式、申请主机、申请端口、申请时间、密码（可选）、备注**；后端 models + 迁移 SQL + CRUD API + 前端管理页
  - 定位：后端新建 `models/zero_trust.go`、`services/zero_trust_service.go`、`controllers/zero_trust.go`，注册 `routes/routes.go`，迁移 SQL 入 `migrations/`；前端新建 `views/ZeroTrustLedger.vue`、`api/zero_trust.js`、`router/index.js` 加路由（入口模式参照 `CloudResourceDialog` 或顶部 Tab 二选一，实现时定）
  - 验收标准：8 个字段齐全可增删改查；`go build` + `golangci-lint` + 前端 `npm run lint:check` 通过；表结构记入 `docs/ARCHITECTURE.md`、接口记入 `docs/API.md`
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
  - 定位（行号为 2026-10-06 T-018 后快照，以 grep `PublicIP\|public_ip` 复核）：后端 `models/host.go:11`、`services/host_service.go`（L100/202/244/302/359-360/473/608/727/760 共 9 处）、`utils/csv.go:15`（表头「公网IP」）、迁移 SQL；前端 `components/HostTable.vue:28`、`components/HostFormDialog.vue:100/209/355`
  - 验收标准：迁移 SQL 写入 `backend/migrations/`（含改名/改类型）；API 响应、主机表单、CSV 导入导出、筛选同步更新；`docs/ARCHITECTURE.md` hosts 表与 `docs/API.md` 同步；前后端验证命令全绿
  - 进入条件（实现前须人类确认）：① 新字段名（如 `ip_mapped`）② 布尔值语义（"已映射"是否区分映射到哪个台账）③ 旧 `public_ip` 数据的迁移/丢弃策略
  - 回写：ARCHITECTURE.md、API.md、BUSINESS_LOGIC.md（如涉及统计口径）、PROJECT_STATUS 变更记录
  - 分支：`feat/host-ip-mapping-flag`
- [ ] **T-016** Goose 迁移执行器实装 ｜ P2 ｜ 负责: — ｜ 备注: 2026-10-06 人类确认**暂不实装，仅入队**；现状为归档-only 口径
  - 内容：引入 `github.com/pressly/goose/v3`，启动 `Migrate()` 时执行 `migrations/*.sql`（`goose.Up`），`AutoMigrate` 降级为兜底或移除；存量 7 个 SQL 与现有库结构一致性验证
  - 定位：`backend/go.mod`（当前无 goose 依赖）、`backend/database/postgres.go` → `Migrate()`（L39-54，现仅 AutoMigrate+seedAdmin）、`backend/migrations/`（7 个 `-- +goose` SQL，**从未执行过**）、如用 CLI 则涉 `backend/run.sh` / `Dockerfile`
  - 进入条件（实现前须人类确认）：① 存量 dev/prod 库结构与 SQL 文件是否一致（不一致需先补差量 SQL）② AutoMigrate 的去留 ③ seed_admin.sql 与 `seedAdmin()` 的职责划分
  - 验收标准：新库从零启动仅靠 SQL 建表成功；重复启动幂等；`go build` + lint 通过
  - 回写（口径反转，必做）：DEVELOPMENT.md 迁移规范口径表、AGENTS.md 红线 9、PROJECT_STATUS 注意事项 8、ARCHITECTURE 目录树注释——改回"运行时执行 SQL"
  - 分支：`feat/goose-migrator`
- [ ] **T-019** 前端抽 Layout/AppNav ｜ P2 ｜ 负责: — ｜ 备注: 定位已实测（2026-10-06）
  - 内容：5 个页面 Tab 各自内嵌同一段 `nav-tab` 导航结构（各 7 处 `nav-tab` 引用），抽为共享 `components/AppNav.vue` 或 Layout（`App.vue` 根布局），消除重复粘贴
  - 定位：`frontend/src/views/HostManagement.vue`、`ResourceStatistics.vue`、`IpStatistics.vue`、`BusinessStatistics.vue`、`PersonnelManagement.vue`（`Login.vue` 不涉及）
  - 验收标准：`npm run lint:check && npm run build` 通过；导航与视觉零变化；重复结构只剩一处
  - 分支：`refactor/extract-layout`
- [ ] **T-020** 消除 `config.go` `ServerPort` 硬编码 ｜ P2 ｜ 负责: — ｜ 备注: 违反红线 3（配置外置）
  - 内容：`ServerPort` 默认值硬编码 `"5677"`；改为环境变量（如 `SERVER_PORT`）读取，默认值是否保留由人类定口径
  - 定位：`backend/config/config.go:43`（默认值赋值处，字段声明 `:22`）、`backend/main.go:23`（`addr := ":" + config.AppConfig.ServerPort`）、同步 `docker/.env.example`
  - 验收标准：端口经环境变量可配；`go build` + `golangci-lint run ./...` 通过；PROJECT_STATUS 技术债表该行删除
  - 分支：`fix/config-server-port`

## blocked

（无。填写格式：`- [ ] **T-xxx** … ｜ blocked 原因：具体条件` —— 泛泛的"困难"不算 blocked）

## done

| ID | 任务 | 完成时间 | 提交 | 备注 |
|----|------|----------|------|------|
| T-021 | 文档一致性自动校验脚本 `scripts/check_docs.sh`（5 项检查 links/tree/snapshot/counts/shas + AGENTS/ROADMAP 4 处接线） | 2026-10-06 | `81f442a` | 验收：基线输出 5 errors/6 warnings 与已知漂移完全一致不误报；注入坏锚点/坏快照 → exit 1，恢复后 exit 0；`bash -n` 通过；T-012/T-007 验收已接线 |
| T-018 | 后端 lint 清零 + golangci-lint v2 配置迁移（`.golangci.yml` v1→v2、修 20 个问题、DEVELOPMENT Linter 章节同步） | 2026-10-06 | `ec2bed6` | 验收：`go build` + `golangci-lint run ./...` exit 0；CSV 解析容错语义经临时冒烟测试确认不变 |
| T-001 | 协作文档体系改造（AGENTS 改为 AI 入口、ROADMAP 改为任务队列、删除 ONBOARDING、分支策略对齐 main+feature） | 2026-10-06 | `0237d40` | 验收：无 ONBOARDING 链接残留、AGENTS 含状态快照/会话协议/验证命令、md 链接锚点全通 |
| — | 2026-09 及之前的交付项 | — | 见 `git log` | 详细记录见 [PROJECT_STATUS.md](./PROJECT_STATUS.md#近期变更记录) |

> 归档规则：done 条目保留 3 个月后可移入上方汇总行；**ID 不复用**。

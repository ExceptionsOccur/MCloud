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
| 内容 | 任务目标与做法概述（一段式，可带子项） |
| 验收标准 | 客观可判定（命令通过 / 文件存在 / 行为可复现），禁止主观描述 |
| 定位 | 涉及文件清单，来源 `docs/CODE_INDEX.md`「功能→代码映射」；**领任务后必须核对/补齐**。优先写**函数名/路由路径**等稳定锚点；行号仅为快照（以 grep 复核为准，可省略），补不出先声明、不盲写 |
| 回写 | 完成后必须更新的文档（通常由 AGENTS.md 责任矩阵推导，此处列出例外） |
| 分支 | 功能分支名（`feat/fix/docs/refactor/<描述>`）；未写则按 AGENTS 红线 8 推导 |
| 文风 | done 备注只写验收结论（≤ 80 字）；PROJECT_STATUS 变更记录每条一行摘要（≤ 120 字），**禁止** ①②③ 枚举子改动与粘贴验收全文（细则见 [DEVELOPMENT.md](./DEVELOPMENT.md) 文档回写文风） |

**认领规则**：一次只领 **1 条**优先级最高（`P0` > `P1` > `P2`）的任务；领用 = 把该条移到 `in_progress` 并填写负责字段（在写任何代码之前先改本文件）。

**回顾锚点**：`上次回顾至 T-041`。`done` 表自该锚点起新增 ≥ 5 条时，须先执行 [AGENTS.md · 会话协议 E](../AGENTS.md#会话协议) 的周期回顾（按各任务「回写」字段定向核查文档↔代码符合度、补回填写漏，纯文档类任务跳过，偏差先报告人类再修），回顾完成并经人类确认后把锚点推进至该批最后一条 `done` ID。

---

## in_progress

（无）

## todo（按优先级）

- [ ] **T-002** 后端 `services/host_service.go` 拆分 ｜ P1 ｜ 负责: —
  - 依赖：已解除（T-018 lint 清零已完成）
  - 内容：拆为 `host_service.go`（CRUD/筛选）+ `host_batch_service.go`（文本/结构化批量）+ `host_csv_service.go`（行映射、导入导出）
  - 定位：`backend/services/host_service.go`（18 个函数，22KB）；回写 `docs/CODE_INDEX.md` 服务索引
  - 验收标准：`cd backend && go build ./...` 与 `golangci-lint run ./...` 通过；**对外 API 行为零变化**；`docs/CODE_INDEX.md` 服务索引同步
  - 分支：`refactor/split-host-service`
- [ ] **T-003** `controllers/csv.go` 业务逻辑下沉 ｜ P1 ｜ 负责: —
  - 依赖：已解除（T-018 lint 清零已完成）
  - 内容：行数校验、表头判断、去重等逻辑移入 service，controller 只剩绑定+调用+响应（红线 1）
  - 定位：`backend/controllers/csv.go`（文件后缀判断、行循环、行数校验、重复判断）→ 下沉至 `services/host_service.go` 的 CSV 部分
  - 验收标准：controller 中无业务校验代码；导入导出行为不变；`go build` + lint 通过
  - 注意：与 T-002 共享 `host_service.go`，建议排在 T-002 之后或同分支执行
  - 分支：`refactor/csv-controller-thin`
- [ ] **T-004** 前端 composables 抽取 ｜ P1 ｜ 负责: —
  - 内容：新建 `frontend/src/composables/`，抽 `useProbeWebSocket` / `useChartOption` / `usePagedTable`
  - 定位：抽取源 —— `views/IpStatistics.vue`（WebSocket+探测）、`views/ResourceStatistics.vue` / `views/BusinessStatistics.vue`（ECharts 配置）、`views/HostManagement.vue` + `stores/host.js`（分页）
  - 验收标准：`cd frontend && npm run lint:check && npm run build` 通过；抽离后各页面行为不变
- [ ] **T-005** `IpStatistics.vue`（835 行）拆分 ｜ P1 ｜ 负责: —
  - 依赖：T-004
  - 内容：拆出 `components/ip/`（网格局部、探测面板、网段管理），单文件 ≤ 400 行
  - 定位：`frontend/src/views/IpStatistics.vue`（全文 835 行）→ 新建 `frontend/src/components/ip/`
  - 验收标准：拆分后 `IpStatistics.vue` ≤ 400 行；`npm run lint:check && npm run build` 通过；功能无回归
- [ ] **T-006** 测试骨架 ｜ P1 ｜ 负责: —
  - 内容：后端 `services` 表驱动单测（先覆盖主机 CRUD、登录）+ 前端 Vitest 冒烟
  - 定位：新建 `backend/services/*_test.go`；前端 `frontend/package.json`（加 vitest 依赖与 script）+ `src/**/__tests__/`
  - 验收标准：`cd backend && go test ./...` 有用例非零且通过；`npm run test` 可运行
- [ ] **T-042** 批量编辑 `apply_time` 错误码与空值语义修复 ｜ P1 ｜ 负责: — ｜ 备注: 全量回顾 J1（2026-10-09）
  - 内容：`PUT /api/batch/hosts` 对 `apply_time=null`/非字符串返回裸 error → 50001 且 null 被拒；改为 null 与省略同义放行、非法值统一 40001（与单条口径一致）
  - 定位：`backend/services/host_service.go`（批量编辑 `isStr` 分支 ~L704）、`backend/controllers/batch.go`（L63 错误映射）、`docs/API.md` 日期时间格式章节（现按 50001 如实描述，修复后同步）
  - 验收标准：`apply_time=null` 不再 50001；非字符串/非法日期返回 40001；`go build` + `golangci-lint run ./...` 通过；`docs/API.md` 同步
  - 分支：`fix/batch-apply-time-code`
- [ ] **T-008** 数据导出报表（Excel/CSV） ｜ P2 ｜ 负责: — ｜ 备注: T-038 后收窄（2026-10-09）
  - 内容：全量导出已覆盖——主机 CSV（`GET /api/export`，BOM）+ 8 表 xlsx 备份（T-038 `/api/export/all`）；剩余缺口 = **按筛选条件导出**：`csv.Export` 裸 `Find(&hosts)` 不带筛选、前端 `exportCSV()` 无参，导出与列表所见不一致；补齐筛选参数透传（接口复用列表筛选 + SearchToolbar 传参），原「新建 report_service.go」定位作废
  - 定位：`backend/controllers/csv.go`（`Export` 裸查全表）、`backend/services/host_service.go`（列表筛选逻辑复用）、`frontend/src/api/csv.js`（`exportCSV()` 无参）、`frontend/src/components/SearchToolbar.vue`（`handleExport`）
  - 验收标准：带筛选导出的行数与列表一致；`go build` + `golangci-lint` + `lint:check` + `build` 通过；`docs/API.md` 同步参数
  - 回写：API.md `GET /api/export` 查询参数
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
  - 定位：`backend/services/stats_service.go` L101（`exec.CommandContext(ctx, "ping", ...)`）、`pingICMP`/`probeTCP`；调用链 `controllers/stats.go` → `Probe`、`controllers/websocket.go`
  - 验收标准：探测颜色状态机行为与 `docs/BUSINESS_LOGIC.md` 描述一致
  - 回写：BUSINESS_LOGIC「探测方式」
- [ ] **T-012** 文档一致性巡检 ｜ P2 ｜ 负责: — ｜ 备注: 适合小型会话
  - 定位：`docs/*.md`、`README.md`、`AGENTS.md`（协议条文部分需人类授权，见红线 7）；校验器 `scripts/check_docs.sh`（T-021 已交付）
  - 验收标准：`bash scripts/check_docs.sh` 退出码 0（links/tree/snapshot/counts/shas 全绿）；warn 一并处理；发现的错误全部修复
- [ ] **T-019** 前端抽 Layout/AppNav ｜ P2 ｜ 负责: — ｜ 备注: 定位已实测（2026-10-09）
  - 内容：各页内嵌同一段 `nav-tab` 导航结构（9 个功能页、每文件 8-9 处 `nav-tab` 引用），抽为共享 `components/AppNav.vue` 或 Layout（`App.vue` 根布局），消除重复粘贴
  - 定位：`frontend/src/views/` 下 9 个功能页（HostManagement / ResourceStatistics / IpStatistics / BusinessStatistics / PersonnelManagement / ZeroTrustLedger / MappingLedger / PublicIPManagement / DataBackup；`Login.vue` 不涉及）
  - 验收标准：`npm run lint:check && npm run build` 通过；导航与视觉零变化；重复结构只剩一处
  - 分支：`refactor/extract-layout`
- [ ] **T-020** 消除 `config.go` `ServerPort` 硬编码 ｜ P2 ｜ 负责: — ｜ 备注: 违反红线 3（配置外置）
  - 内容：`ServerPort` 默认值硬编码 `"5677"`；改为环境变量（如 `SERVER_PORT`）读取，默认值是否保留由人类定口径
  - 定位：`backend/config/config.go:43`（默认值赋值处，字段声明 `:22`）、`backend/main.go:23`（`addr := ":" + config.AppConfig.ServerPort`）、同步 `docker/.env.example`
  - 验收标准：端口经环境变量可配；`go build` + `golangci-lint run ./...` 通过；PROJECT_STATUS 技术债表该行删除
  - 分支：`fix/config-server-port`
- [ ] **T-022** WS 探测通道 token 有效性校验 ｜ P2 ｜ 负责: — ｜ 备注: 2026-10-06 人类立队（安全问题，后面处理）
  - 内容：`/api/ws/probe` 当前**只校验 `token` query 参数非空**（空则 `40101`），不校验 JWT 有效性——任何非空字符串即可建立 WebSocket 连接并发起探测；需复用 `middleware/jwt.go` 的校验逻辑，无效/过期/伪造 token 拒绝升级
  - 定位：`backend/routes/routes.go`（`/api/ws/probe` 处仅 `c.Query("token") != ""` 判断）、`backend/controllers/websocket.go`（`HandleProbeWS`；`upgrader.CheckOrigin` 恒 `true`，同链路可一并评估是否收紧）、`backend/middleware/jwt.go`（`JWTAuth()` 校验逻辑参考）、`docs/API.md`「WebSocket 探测帧协议」章节（同步改）
  - 验收标准：无效/过期/伪造 token 无法建立连接（`40101`）；有效 token 行为与现网一致；`go build` + `golangci-lint run ./...` 通过；API.md 与 BUSINESS_LOGIC 相关描述同步
  - 分支：`fix/ws-probe-jwt`
- [ ] **T-043** CORS Origin 白名单收紧 ｜ P2 ｜ 负责: — ｜ 备注: 全量回顾 J2（安全类，2026-10-09，与 T-022 同批后置）
  - 内容：`middleware/jwt.go` 当前对任意 Origin 回显 `Access-Control-Allow-Origin` + `credentials=true`，无白名单；改为环境变量可配的 Origin 白名单，非白名单不回显 CORS 头
  - 定位：`backend/middleware/jwt.go`（CORS 段 ~L57-72）、`docker/.env.example`、`docs/DEVELOPMENT.md` 安全准则 CORS 行
  - 验收标准：非白名单 Origin 不回显 CORS 头；白名单内预检/带凭证请求行为正常；`go build` + `golangci-lint run ./...` 通过；DEVELOPMENT 同步
  - 分支：`fix/cors-origin-whitelist`
- [ ] **T-044** 后端 CSV 上传大小限制 ｜ P2 ｜ 负责: — ｜ 备注: 全量回顾 J3（2026-10-09）
  - 内容：`controllers/csv.go` 导入无大小限制（16MB 仅为前端提示）；后端加限制（`http.MaxBytesReader`，上限对齐前端 16MB），超限返回 40001
  - 定位：`backend/controllers/csv.go`（Import/Export）、`frontend/src/components/ImportDialog.vue`（16MB 提示）、`docs/DEVELOPMENT.md` 安全准则上传行
  - 验收标准：超限导入返回 40001 且信息可读；正常导入不受影响；`go build` + `golangci-lint run ./...` 通过；DEVELOPMENT 同步
  - 分支：`fix/csv-upload-limit`

## blocked

（无。填写格式：`- [ ] **T-xxx** … ｜ blocked 原因：具体条件` —— 泛泛的"困难"不算 blocked）

## 已取消

| ID | 任务 | 取消时间 | 原因 |
|----|------|----------|------|
| T-037 | 零信任 `apply_time` 无时区格式解析失败 | 2026-10-09 | 人类裁定并入 T-040（全仓日期格式统一）一次执行 |
| T-007 | CI 流水线落地 | 2026-10-06 | 项目暂不引入 CI/CD；所有集成与检验由 agent 完成任务后本地执行「验证命令」（见 [AGENTS.md](../AGENTS.md#验证命令门禁)） |

## done

| ID | 任务 | 完成时间 | 提交 | 备注 |
|----|------|----------|------|------|
| T-046 | 零信任列表「申请主机+申请端口」合并为「申请资源」标签列 | 2026-10-09 | `9c1cbfa` | 验收：两列并一列标签 `主机名(ip:port)` 按主机名排序；前端门禁+check_docs 全绿；API/DB 零改动 |
| T-045 | 协议补强：周期回顾偏差核查跳过纯文档类任务 | 2026-10-09 | `5fbb249` | 验收：AGENTS E.2/DEVELOPMENT/ROADMAP 三处口径同步纯文档类任务跳过；check_docs 0 errors |
| T-041 | 协议补强：每完成 5 条任务的周期回顾 | 2026-10-09 | `4ccd241` | 验收：协议 E 触发/范围/处理三要素落地；锚点 T-040 两处一致；DEVELOPMENT 同步；check_docs 0 errors |
| T-040 | 全仓日期时间格式统一（吸收 T-037） | 2026-10-09 | `9522789` | 验收：零信任时间双格式兼容；主机申请时间全链路严格 `YYYY-MM-DD`；展示与文件名口径回写 API.md；门禁全绿 |
| T-039 | 空库启动迁移失败修复（UNIQUE 约束对齐 GORM uniqueIndex 口径） | 2026-10-09 | `c196f37` | 验收：空库启动+登录通过；4 表 `*_key` 转 `idx_`；外键/重复索引去重；dev 迁移 no-op；门禁全绿 |
| T-038 | 数据备份：8 个业务 sheet 统一导出(xlsx)/导入(upsert) | 2026-10-09 | `6644d5a` | 验收：8 sheet 列=设计清单；空库回灌关联完整重建；重复导入 0 增删；行级错误回滚库不变；门禁全绿 |
| T-036 | 零信任台账多组主机:端口配对 + 公网IP带出地区 | 2026-10-08 | `ab7e3d0` | 验收：targets 配对 CRUD+迁移；公网IP池校验带出地区；批量等长配对；引用40901；门禁+冒烟全绿 |
| T-035 | 映射台账公网IP改为资源池选择 | 2026-10-08 | `627ee8f` | 验收：公网IP下拉选资源池；表单无运营商/出口位置；列表从资源池带出 |
| T-034 | 域名台账改造为端口映射台账 | 2026-10-08 | `d26d6ad` | 验收：port_mappings 多端口等长；ip_mapped 自动；/mapping-ledger；迁 domains 数据 |
| T-033 | 域名/公网IP出口位置字段 | 2026-10-08 | `aac7fea` | 验收：exit_location；运营商之后；域名+公网IP+批量列；迁移+验证全绿 |
| T-032 | 零信任/域名台账批量添加 | 2026-10-08 | `4e96c09` | 验收：两台账 /batch 文本接口；主机按内网IP；域名跳过零信任全插；前端工具栏可用 |
| T-031 | 公网 IP 资源录入 | 2026-10-08 | `e0d84ac` | 验收：public_ips CRUD+IP唯一/格式校验；设置菜单+独立页；迁移+验证全绿 |
| T-015 | 公网 IP 字段改造 + 双台账关联 | 2026-10-08 | `130388a` | 验收：ip_mapped 布尔；详情展示两台账关联；CSV/表单/筛选同步；迁移+验证全绿 |
| T-030 | 域名台账字段改造 | 2026-10-08 | `e83db90` | 验收：isp/host_id/host_port；无 provider/expires_at；FK 校验+主机删除 40901；迁移+验证全绿 |
| T-029 | 零信任台账增加系统名称字段 | 2026-10-08 | `1f70851` | 验收：API 含 system_name；列/表单位于申请端口后；迁移+前后端验证全绿 |
| T-016 | Goose 迁移执行器实装 | 2026-10-08 | `82e0e07` | 验收：存量 dev 库首跑 9 迁移+幂等；build/lint/check_docs 过；口径反转已回写 |
| T-028 | 前端导航修正：台账入口与 Tab 一致性 | 2026-10-08 | `2e32a5b` | 验收：设置菜单无台账入口；各页 Tab 含零信任+域名；lint/build 通过 |
| T-027 | 前端 ESLint 警告清零 | 2026-10-08 | `ebb52de` | 验收：lint:check 0 problems + build 通过 |
| T-026 | 协议补强：多任务串行合并 | 2026-10-08 | `c5208c6` | 验收：AGENTS A.2/A.4 + 分支协议 + DEVELOPMENT 口径一致；check_docs 0 errors |
| T-014 | 域名台账 | 2026-10-08 | `4e0190c` | 验收：5 字段 CRUD + 域名唯一；前后端验证通过 |
| T-013 | 零信任台账 | 2026-10-08 | `6a6e760` | 验收：7 字段 CRUD + 主机 FK 校验；前后端验证通过 |
| T-025 | 协议补强：临时任务入队 + 提交权限时序 + 分支/接管规则 | 2026-10-06 | `1837874` | 验收：A.0/A.4/D 与红线 5 例外落地；提交号时序兼容 shas；check_docs.sh 0 errors |
| T-024 | 文档回写文风上限（防膨胀） | 2026-10-06 | `bcfb410` | 验收：`check_docs.sh` 0 errors；三处口径一致；超长条目已压缩 |
| T-023 | 文档卫生清理 + 移除 CI/CD 规划内容 | 2026-10-06 | `42d4c10` | 验收：`check_docs.sh` 0 errors；规划语境无 T-007/CI 残留；协议改动经人类授权 |
| T-017 | 文档事实修正与协议补强 | 2026-10-06 | `6187228` | 验收：`check_docs.sh` 0 errors；协议条文改动仅限任务列明项 |
| T-021 | 文档一致性自动校验脚本 `scripts/check_docs.sh` | 2026-10-06 | `81f442a` | 验收：5 项检查可用；注入坏数据可正确报错退出 |
| T-018 | 后端 lint 清零 + golangci-lint v2 配置迁移 | 2026-10-06 | `ec2bed6` | 验收：`go build` + `golangci-lint run ./...` exit 0 |
| T-001 | 协作文档体系改造（AGENTS/ROADMAP/删除 ONBOARDING） | 2026-10-06 | `0237d40` | 验收：无 ONBOARDING 残留；md 链接锚点全通 |
| — | 2026-09 及之前的交付项 | — | 见 `git log` | 详细记录见 [PROJECT_STATUS.md](./PROJECT_STATUS.md#近期变更记录) |

> 归档规则：done 条目保留 3 个月后可移入上方汇总行；**ID 不复用**。备注只写验收结论（≤ 80 字），见 [DEVELOPMENT.md](./DEVELOPMENT.md) 文档回写文风。

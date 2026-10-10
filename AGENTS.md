# AGENTS.md — MCloud 项目总纲（AI 协作入口）

> **本文件是所有 AI 协作会话的唯一必读入口**，核心状态与执行协议已内联，读完本文件即可开工，无需先读其他文档。
> 人类可读的项目介绍见 [README.md](./README.md)；需要细节时按「文档地图」跳转。
> **最后更新：2026-10-09**

## 会话协议

### A. 开工前（顺序执行）

**A.0 临时任务入队**（仅当人类指定的需求不在 ROADMAP `todo`/`blocked` 时执行，否则跳过本步）：

1. 分配 ID：取 ROADMAP 已出现的 `T-xxx` 最大值 + 1，**永不复用**
2. 在 ROADMAP `todo` 区新建条目，按字段约定填写：内容 / 优先级（默认 `P1`，人类可指定）/ 验收标准（客观可判定）/ 定位（依据 [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) 补齐，补不出先声明不盲写）/ 分支 / 回写
3. 确认门槛：验收标准主观歧义，或涉及敏感字段/口径（存储方式、外键、数据迁移策略等）→ **先向人类确认**，再进入 `in_progress`

然后按下列步骤执行：

1. **A.1** 读本文件的「状态快照」与「核心约定（红线）」，并与 [docs/ROADMAP.md](./docs/ROADMAP.md) 比对（todo 队列 / `in_progress` / `next_task` / 回顾锚点）——不一致时**先修快照**（以 ROADMAP 为准）
2. **A.2** 读 [docs/ROADMAP.md](./docs/ROADMAP.md)（任务状态的**唯一事实来源**），认领 1 条优先级最高的 `todo` 任务（A.0 入队的临时任务即为该条）。**人类一次指定多条任务时，仍一次只认领 1 条**；上一条必须完成串行合并（见下方 A.4 与「分支与提交协议」）后才可认领下一条。**认领前检查**：ROADMAP `done` 自「回顾锚点」起 ≥ 5 条时，先执行会话协议 E 的周期回顾
3. **A.3** **认领 = 先改文件再写代码**：把该任务移到 `in_progress`，填写负责标识与日期
4. **A.4** **创建分支**（认领后、写代码前）：`git checkout main && git pull && git checkout -b <分支名>`；分支名取任务「分支」字段，未写则按红线 8 推导（`feat/fix/docs/refactor/<简短描述>`）；一个分支只对应 1 条任务，**禁止直接 push `main`**。**基线约束**：多任务会话中，任务 N+1 的分支必须从**已合并任务 N 后的 `main`** 拉出；禁止在未合并的任务分支上并行开出下一条任务分支
5. **A.5** **定位代码**：查 [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) 的「功能→代码映射」「请求链路」与 [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) 目录结构，把涉及文件写进该任务的「定位」字段——任务已带「定位」则核对并以其为起点，没有则必须补齐；**补不出（找不到实现位置）时先在输出中声明，不得盲写代码**
6. **A.6** 按任务的「验收标准」倒推实现方案；涉及公共文件（`routes/`、`models/`、本文件）时先在输出中声明影响范围

### B. 结束前（回写协议，未回写视为任务未完成）

| # | 动作 | 目标文件 |
|---|------|----------|
| 1 | 任务状态 → `done`，填完成时间与提交号 | [docs/ROADMAP.md](./docs/ROADMAP.md) |
| 2 | 功能完成度、变更记录（日期 + 提交号 + **一行摘要 ≤120 字，禁止 ①②③ 枚举**） | [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) |
| 3 | 按下方「文档更新责任矩阵」同步受影响文档 | `docs/*.md` |
| 4 | 刷新本文件「状态快照」，全量对齐 ROADMAP（`last_updated`、todo 队列、`next_task`、`in_progress`、`blocked`、回顾锚点） | 本文件 |
| 5 | 全部「验证命令」跑绿（文档改动含 `bash scripts/check_docs.sh`） | — |

### 提交号回写时序（与红线 5 配套）

1. 代码 + 文档初稿完成（ROADMAP 任务仍在 `in_progress`，`done` 表「提交」列暂留 `-`）
2. 验证命令全绿（含 `check_docs.sh`；此时 `done` 表无本任务 SHA，`shas` 检查不涉及）
3. 人类已明确要求提交 → `git commit`（原子提交：代码 + 文档 + ROADMAP 状态变更；**禁止 push**，除非人类明说）
4. 用真实 SHA 回填 ROADMAP `done` 表「提交」列 + [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) 变更记录
5. 补一个 `docs:` 提交（仅含上述回填字段）
6. 再跑 `bash scripts/check_docs.sh` → 退出码 0（`shas` 校验真实 SHA 在 git 历史中存在）
7. 合并回 `main` 后删除功能分支（生命周期 ≤ 1 周）

### C. 冲突与降级规则

- 状态快照与 ROADMAP 不一致 → **以 ROADMAP 为准**，并修复快照
- 文档与代码不一致 → **以代码为准**，修复文档（`docs:` 提交）
- 验收标准无法满足 → 任务改 `blocked` 并写明**具体阻塞条件**（泛泛的"困难"不算），不要留在 `in_progress`
- 需求本身有歧义 → 停下来向人类确认，不要自行扩大范围

### D. 会话中断与接管

- 新会话开工时若 ROADMAP `in_progress` 非空：
  - 「负责」标识非当前会话/非人类指令的进行中任务 → **不得接管**，不得清理其分支；按 A.2 从 `todo` 取 1 条
  - 人类明确要求接管/继续 → 走认领流程，更新「负责」为当前会话标识
- 分支残留清理仅限：任务已 `done` 且人类要求，或该分支属于当前会话

### E. 周期回顾（每完成 5 条任务）

1. **触发**：ROADMAP `done` 表自「回顾锚点」（ROADMAP 字段约定区的 `上次回顾至 T-xxx`）起新增 **≥ 5 条**时，在认领下一条任务**之前**执行周期回顾
2. **范围**：仅核查该批 5 条任务——逐条按其 ROADMAP「回写」字段 +「文档更新责任矩阵」推导出的文档清单，定向核查**工作流文档与实际代码的符合度**（回填写漏、口径漂移、链接失效）；**跳过纯文档类任务**（验收内容仅改文档、不涉及 `backend/`/`frontend/` 代码者——其本身即在更新文档，核查文档同步属重复劳动）；不做全量文档巡检（全量属 `T-012` 类独立任务）
3. **处理**：发现偏差或漏回写 → 先输出偏差清单**报告人类**，经人类确认后再修复（`docs:` 提交），不自行扩大范围；需改代码才能修的问题按 A.0 入队新任务
4. **收尾**：偏差处理经人类确认后，把 ROADMAP 回顾锚点推进至该批最后一条 `done` 的任务 ID，并同步本文件「状态快照」锚点行（回写协议第 4 步）
5. 回顾是协议内建动作，**不占用任务 ID**；锚点以 ROADMAP 为准，本文件快照只是缓存

## 状态快照

> 这是 [ROADMAP](./docs/ROADMAP.md) 的缓存，回写协议第 4 步负责刷新它。

| 字段 | 值 |
|------|-----|
| last_updated | 2026-10-10 |
| in_progress | `T-051` 公网IP引用保护+映射防重（负责: opencode mimo-v2.6-flash-free 2026-10-10） |
| next_task | `T-050` hosts.disk 补录入口（= ROADMAP todo 首条，P1 最高优先级） |
| todo（P1） | `T-050` hosts.disk 补录入口 → `T-004` composables 抽取 → `T-005` IpStatistics 拆分 → `T-006` 测试骨架 |
| todo（P2） | `T-008` 报表导出、`T-009` bcrypt 升级、`T-011` ping 解耦、`T-012` 文档巡检、`T-019` 抽 Layout/AppNav（功能页已 10 个，按 10 页更新）、`T-020` config ServerPort 硬编码、`T-022` WS probe token 校验、`T-043` CORS 白名单、`T-044` CSV 上传限制、`T-052` targets 拆子表 |
| blocked | 无 |
| 回顾锚点 | 上次回顾至 `T-010`（`done` 自此起满 5 条触发周期回顾，见会话协议 E） |
| 功能完成度 | 功能模块 19 项：17 完成 / 2 待开发；技术债 7 项（详见 [PROJECT_STATUS · 功能完成度](./docs/PROJECT_STATUS.md#功能完成度)） |
| 已知风险 | 无测试、密码为 SHA-256、WS 探测仅校验 token 非空（`T-022`）（详见 [PROJECT_STATUS](./docs/PROJECT_STATUS.md#已知问题--技术债)） |

> T-036~T-041 各批次均已合并 main（提交号见 ROADMAP done 表）；全量回顾偏差修复与锚点推进见 ROADMAP「回顾锚点」。

## 项目速览

MCloud 是云平台主机资产信息管理系统：**Go (Gin) 后端 + Vue 3 SPA 前端 + PostgreSQL 16**。

```
浏览器 (Vue 3)  ──Axios──▶  /api        ┐
                ──WS────▶  /api/ws/probe ├─▶  Go (Gin) ─▶ PostgreSQL 16
                                        ┘
```

**分层（红线 1，不可破坏）**：`routes`（注册+JWT）→ `controllers`（仅绑定参数、调 service、`Success/Error` 响应）→ `services`（全部业务逻辑）→ `models`（GORM）。

**前端结构**：`api/`（Axios 封装）→ `stores/`（Pinia）→ `views/`（10 个功能页：7 顶部 Tab + 3 设置菜单页）→ `components/`。

| 关键事实 | 值 |
|----------|-----|
| 开发端口 | 前端 5173（Vite，`/api` 代理到 `backend:5677`）、后端 5677、Postgres 5432、生产 5678 |
| 默认账号 | `admin` / `Pass4MCloud`（种子见 `backend/database/postgres.go` 的 `seedAdmin()`） |
| 数据库表 | 10 张：`users` / `hosts` / `host_applications` / `persons` / `cloud_resources` / `ip_subnets` / `zero_trusts` / `port_mappings` / `public_ips` / `audit_logs` |
| 开发环境 | `cd docker && cp .env.example .env && docker-compose -f docker-compose.dev.yml up -d` |

## 核心约定（红线）

细节见 [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)，以下为硬性约束：

1. **分层**：`controllers` 不含业务逻辑，业务全部在 `services`
2. **统一响应**：只用 `controllers.Success` / `controllers.Error`，**禁止** `c.JSON`
3. **配置外置**：数据库连接、密钥一律走环境变量，**禁止硬编码**
4. **GORM 列名**：缩写字段（如 `CIDR`）必须显式 `gorm:"column:xxx"`，否则生成 `c_id_r` 错误列名
5. **提交**：Conventional Commits（`feat:` / `fix:` / `docs:` …）；**未经明确要求不主动 commit / push**
   - 例外：人类会话指令中明确要求提交（如「完成后提交」）时，agent 在验证命令全绿后可执行 `git commit`；`git push` 仍须人类明说
   - 未提提交时：验证通过后停在「待提交」，在输出中说明，由人类决定；提交号回写步骤见「提交号回写时序」
6. **敏感信息**：`.env`、密钥、密码、Token **不入库**
7. **改动聚焦 + 修改权限**：只改当前任务验收标准内的代码，**不顺手重构**。可改范围按文件类别判定：
   - **状态与同步类**（`docs/ROADMAP.md`、`docs/PROJECT_STATUS.md`、责任矩阵指定的同步文档）→ **必须**随任务回写，这是义务，不算违规
   - **普通代码与业务文档**（`backend/`、`frontend/`、`docs/API.md` 等）→ 仅限任务验收标准覆盖的范围
   - **协议条文**（本文件的「会话协议 / 红线 / 状态快照规则」、`docs/DEVELOPMENT.md` 的规范条文）→ **默认禁止**；只有任务验收标准**明确写明**要改它（即人类发起的 `docs:` 任务）才可改，否则在输出中提出建议、不动手
   - **构建文件**（`Dockerfile`、`docker-compose*.yml`、`.golangci.yml`、`.eslintrc.cjs`、`nginx.conf`）→ **未经人类明确要求禁止修改**
8. **分支**：从 `main` 拉 `feat/fix/docs/refactor` 短生命周期分支，PR 合并回 `main`；**禁止直接 push `main`**
9. **迁移**：模型变更必须随附迁移 SQL（`backend/migrations/`，`-- +goose` 格式）；运行时由启动时 `Migrate()` 先执行 goose（embed FS 打包进二进制），`AutoMigrate` 仅兜底；**SQL 必须与模型定义一致**（统一口径见 [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) 数据库迁移规范）
10. **门禁**：提交前「验证命令」必须全绿
11. **环境一致**：数据库只用 docker-compose 启动，禁止手动安装
12. **文档同步**：代码改完必须按「文档更新责任矩阵」回写文档；**未回写 = 任务未完成**

## 分支与提交协议

```
main ← 唯一主干，只接受 PR，禁止直接 push
  ├── feat/<描述>       功能
  ├── fix/<描述>        修复
  ├── docs/<描述>       文档
  └── refactor/<描述>   重构（T-002~T-005 类任务必须走此分支，不混入业务）
```

- 分支从 `main` 拉出，合并回 `main` 后立即删除，**生命周期 ≤ 1 周**
- 一个分支只对应 ROADMAP 中的 **1 条任务**；PR 标题 = Conventional Commits 格式，描述含：改了什么 / 为什么 / 影响范围（接口、表、页面）/ 关联任务 ID / 文档回写清单
- **多任务串行合并（人类一次指定多条任务时）**：
  1. 完成任务 N 的代码 + 文档 + 验证 + 提交（按「提交号回写时序」），**合并回 `main` 并删除任务 N 分支**后，才允许认领任务 N+1
  2. 任务 N+1 分支必须从合并后的 `main` 重新 `checkout -b`；禁止在任务 N 未合并时并行开 N+1 分支
  3. 后一条任务的文档回写（表数/SQL 数/完成度/快照）以**合并后的 main** 为基线计算，不得沿用上一条合并前的中间态
  4. 目的：避免多任务并行分支在 `routes/`、导航、ARCHITECTURE/API 计数、AGENTS 快照等公共文件上产生冲突与计数漂移
- 合并方式优先 squash；合并完成即触发「回写协议」第 1、2 步
- 本项目**暂不引入 CI/CD**；所有集成与检验由 agent 在完成任务后**本地执行「验证命令」**，全部通过才算任务完成

## 文档更新责任矩阵

> 谁改代码，谁回写文档；PR/任务结束时未同步视为未完成。

| 变更内容 | 必须同步 |
|----------|----------|
| 新增/修改 API 接口、路由 | [docs/API.md](./docs/API.md) + [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) |
| 新增/修改数据库字段、表 | [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) 数据模型 + `backend/migrations/` 迁移 SQL |
| 新增/删除前端页面、路由 | [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) + [README.md](./README.md) 页面与导航 |
| 修改关键业务逻辑（探测、统计口径、认证） | [docs/BUSINESS_LOGIC.md](./docs/BUSINESS_LOGIC.md) |
| 修改规范、流程、红线 | [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) + 本文件 |
| 目录/文件结构变化 | [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) 目录结构 + [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) |
| 任务状态变化 | [docs/ROADMAP.md](./docs/ROADMAP.md)（唯一事实来源） |
| 功能合并、技术债增减 | [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) 完成度 + 变更记录 |

## 验证命令（门禁）

按改动范围执行，**全部通过才能结束任务**：

```bash
# 后端改动
cd backend && go build ./... && golangci-lint run ./...

# 前端改动（门禁用 lint:check；npm run lint 带 --fix 会改写文件，仅用于本地修复）
cd frontend && npm run lint:check && npm run build

# 文档改动（链接/锚点、目录树、状态快照、计数、提交号一致性全量校验）
bash scripts/check_docs.sh
# 数据库改动
cd backend && go build ./...   # 确认模型可编译 + 迁移 SQL 文件已按 YYYYMMDDHHMMSS_xxx.sql 命名
```

**工具前置**（环境搭好一次即可）：

- Go 1.22+、Node 18+；数据库与容器一律走 `docker/`（红线 11）
- golangci-lint 安装在 `$(go env GOPATH)/bin`，该目录不在 `PATH` 时先 `export PATH=$PATH:$(go env GOPATH)/bin`；须用能读 v2 配置的版本（v1 会报 `unsupported version of the configuration`，v2.14.0 已验证）

## 文档地图

> 按需跳转，不要一次性读全部。**默认只需本文件**。

| 文档 | 内容 | 何时读 |
|------|------|--------|
| [README.md](./README.md) | 功能特性、快速开始（面向人类） | 需要产品视角 / 部署说明 |
| [docs/ROADMAP.md](./docs/ROADMAP.md) | **任务队列（唯一事实来源）** | 领任务、回写状态 |
| [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) | 完成度、变更记录、技术债、注意事项 | 判断某功能是否已存在、避坑 |
| [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | 目录结构、分层、10 张表模型 | 定位模块、改模型 |
| [docs/API.md](./docs/API.md) | 响应格式、错误码、全部路由与参数 | 增改/调用接口 |
| [docs/BUSINESS_LOGIC.md](./docs/BUSINESS_LOGIC.md) | 认证、搜索、CSV、IP 探测状态机、统计口径 | 改业务逻辑 |
| [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) | 编码规范、提交规范、测试、安全 | 写代码时的细则 |
| [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) | 文件职责、功能→代码映射、请求链路 | 找实现位置（最省 token 的入口） |

## 术语表

| 术语 | 含义 |
|------|------|
| 主机 host | 云主机/物理机资产，核心表 `hosts` |
| 申请信息 | `host_applications`，与 `hosts` 一对一，删主机必须级联删除 |
| 人员 person | `persons` 表，经 `hosts.person_id` 关联主机 |
| 零信任台账 | `zero_trusts` 表，申请单位/账户名/接入目标(`targets`，host_id:port 多组配对)/公网IP入口(接入地区由资源池带出)/时间/备注；主机被引用禁止删除 |
| 域名台账 | 已升级为端口映射台账 `port_mappings`（T-034）：公网IP↔内网主机多端口映射，域名可选 |
| `is_db_server` | BOOLEAN，API 统一返回 `true/false` |
| 云资源总览 | `cloud_resources`，按区域手工录入的总量资源 |
| 裸金属 | 物理服务器，**资源统计一律排除**（避免重复计算） |
| IP 网段 | `ip_subnets`，用户自建，仅接受 /24，不预置数据 |
| 探测颜色 | 🟢 空闲 / 🔴 已用有非空记录 / 🟡 已用但无记录 |
| 「已停止」 | 统计口径 = `已停止` + `已关机` 合并 |
| 位图 | IP 统计页 10 列 × 26 行共 256 格可视化 |

## 常用命令

| 目的 | 命令 |
|------|------|
| 开发环境启动 | `cd docker && cp .env.example .env && docker-compose -f docker-compose.dev.yml up -d` |
| 仅数据库 | `cd docker && docker-compose -f docker-compose.dev.yml up -d postgres` |
| 后端编译 / Lint | `cd backend && go build ./...` / `golangci-lint run ./...` |
| 前端开发 / 构建 / Lint | `cd frontend && npm run dev` / `npm run build` / `npm run lint:check`（检查；修复用 `npm run lint`） |
| 重置 admin 密码 | `cd backend && go run ./cmd/resetpw/main.go [新密码]` |
| 生产部署 | `cd docker && cp .env.example .env && docker-compose -f docker-compose.prod.yml up -d --build` |

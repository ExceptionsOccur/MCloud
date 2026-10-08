# 开发规范与准则

> 本文档定义 MCloud 项目的编码规范、提交准则与协作要求。修改代码前必读。
> 文档入口见 [AGENTS.md](../AGENTS.md)，代码位置见 [CODE_INDEX.md](./CODE_INDEX.md)。

## 目录

- [核心原则](#核心原则)
- [Go 代码规范](#go-代码规范)
- [前端代码规范](#前端代码规范)
- [提交规范](#提交规范)
- [测试规范](#测试规范)
- [安全准则](#安全准则)

---

## 核心原则

1. **分层清晰**：controllers 不含业务逻辑，services 承载全部业务逻辑，routes 只做路由注册
2. **统一响应**：所有 API 通过 `controllers.Success` / `controllers.Error` 返回，禁止直接 `c.JSON`
3. **配置外置**：敏感信息与可变配置一律走环境变量，禁止硬编码
4. **约定优先**：命名、目录、格式遵循项目既有约定，不引入风格不一致的写法
5. **最小改动 + 修改权限**：改动聚焦当前需求，不顺手重构无关代码。文件按类别判定权限：`docs/ROADMAP.md`、`docs/PROJECT_STATUS.md` 及责任矩阵指定的同步文档**必须随任务回写**；协议条文（[AGENTS.md](../AGENTS.md) 的会话协议/红线、本文件规范）**仅当任务验收标准明确写明才可改**；构建文件（`Dockerfile`、`docker-compose*.yml`、`.golangci.yml`、`.eslintrc.cjs`、`nginx.conf`）**未经人类明确要求禁止修改**。完整分类见 AGENTS.md 红线 7
6. **可验证**：提交前确保编译通过、关键路径自测通过

---

## Go 代码规范

### 项目分层

```
routes/      → 路由注册（无业务逻辑）
middleware/  → 中间件（JWT、CORS）
controllers/ → 请求解析 + 响应格式化（无业务逻辑）
services/    → 核心业务逻辑
models/      → 数据模型 + GORM 标签
utils/       → 无状态工具函数
config/      → 配置加载
database/    → 数据库连接与迁移
```

### 命名规范

| 类别 | 规范 | 示例 |
|------|------|------|
| 包名 | 全小写，无下划线 | `controllers` 不是 `controller` |
| 文件名 | 全小写 + 下划线分隔 | `host_service.go` |
| 结构体 | PascalCase | `HostService` |
| 导出函数/方法 | PascalCase | `GetHostByID` |
| 私有函数/方法 | camelCase | `hostExists` |
| 常量 | PascalCase / CamelCase | `MaxFailedAttempts` |
| JSON 字段 | snake_case | `private_ip`, `apply_unit` |

### Controller 规范

Controller 统一为 Gin Handler，只做三件事：

```go
func SomeAction(c *gin.Context) {
    // 1. 绑定参数（ShouldBindJSON / ShouldBindQuery）
    // 2. 调用 service 层
    // 3. 通过 controllers.Success / controllers.Error 返回统一响应
}
```

- **不包含业务逻辑**：数据校验、数据库操作、事务一律放 service
- 参数绑定失败返回 `40001`
- 资源不存在返回 `40401`
- 数据冲突（如 IP 重复）返回 `40901`

### Service 规范

- Service 承载所有业务逻辑：数据校验、GORM 操作、事务管理、错误处理
- 返回值统一为「数据 + error」或「结果 + error」：

```go
func GetHostByID(id uint) (*models.Host, error)
func CreateHost(data CreateHostRequest) (uint, error)
```

- 业务规则错误使用 `errors.New("可读的中文提示")`，由 controller 决定错误码

### 错误处理

所有错误必须通过 `controllers.Error(c, code, message)` 返回：

```go
// ✅ 正确
return controllers.Error(c, 40901, "内网IP已存在")

// ❌ 禁止
c.JSON(400, gin.H{"error": "..."})
```

### 数据库查询规范

- 使用 GORM 链式调用构建查询
- 动态 WHERE 条件使用 `if cond != "" { db = db.Where(...) }` 模式
- 分页统一使用 `Offset` + `Limit`
- 关联查询使用 `Preload` 或 `Joins`
- 所有查询必须处理 `RecordNotFound`
- **缩写字段必须显式指定列名**：GORM 会把 `CIDR` 命名为 `c_id_r`，须加 `gorm:"column:xxx"`

### 配置管理

- 所有配置通过 `config` 包从环境变量加载
- 禁止硬编码数据库连接串、密钥等敏感信息
- 使用 `github.com/joho/godotenv` 加载 `.env`
- `.env` 不提交到 Git

### 数据库迁移规范

**统一口径（2026-10-08 T-016 反转：运行时执行 SQL）**：

| 层面 | 机制 | 说明 |
|------|------|------|
| 运行时 | `backend/migrations/*.sql`（embed FS）+ `github.com/pressly/goose/v3` | 启动时 `Migrate()` 先 `goose.Up` 执行 `migrations/*.sql`（SQL 打包进二进制） |
| 兜底 | `database/postgres.go` → `AutoMigrate()` + `seedAdmin()` | goose 之后仍执行 AutoMigrate 补齐 SQL 未覆盖的模型变更；种子管理员留在代码 |

> 一句话：**SQL 是运行时迁移的唯一事实来源（goose 执行），AutoMigrate 仅兜底，两者不冲突**。SQL 必须与模型定义一致；已执行的迁移文件禁止修改。

**迁移文件规则**：

1. 迁移文件存放在 `backend/migrations/`，命名格式：`YYYYMMDDHHMMSS_<描述>.sql`
2. 每次模型变更（新增字段、改类型、加索引、加约束）必须写对应的 SQL 迁移文件（`-- +goose Up` / `-- +goose Down`）
3. 迁移文件一旦提交，**禁止修改或删除**（已执行的迁移不可变）
4. 运行时**先**执行 goose（`Migrate()` 内 `goose.Up`），**再** `AutoMigrate()` 兜底——**SQL 内容必须与模型定义保持一致**，否则 goose 落地结构与 GORM 模型会漂移

**多人协作流程**：

```
1. 拉取最新 main
2. 创建迁移文件（如 20260322120000_add_status_index.sql）
3. 本地启动，确认 goose 应用后的结构与模型/SQL 描述一致
4. 提交 PR（迁移文件 + 模型变更 + 业务代码一起）
5. 合并后其他开发者拉取代码，启动时由 goose 应用新结构（AutoMigrate 兜底未覆盖项）
```

**冲突预防**：

- 同一时间段内，避免多人同时修改同一张表的 schema
- 如需修改核心表（`hosts`、`users`），先在 Issue 中说明变更内容
- 迁移文件命名的时间戳使用 UTC 时间，避免时区冲突

---

## 前端代码规范

### Vue 3 组件规范

- 统一使用 `<script setup>`（Composition API）
- 组件文件名 PascalCase（`HostTable.vue`）
- Props 用 `defineProps`，Emits 用 `defineEmits`
- 组件内状态用 `ref()` / `reactive()`，跨组件状态用 Pinia store

### Element Plus 使用规范

- 弹窗统一 `el-dialog` + `v-model`
- 表单统一 `el-form` + `el-form-item` + `label-width`
- 表格统一 `el-table` + `el-table-column`，分页用 `el-pagination`
- 下拉框用 `el-select` + `el-option`
- 操作反馈用 `ElMessage.success()` / `ElMessage.error()`
- 二次确认用 `ElMessageBox.confirm()`

### API 请求规范

- 统一在 `src/api/index.js` 创建 Axios 实例
- 请求拦截器注入 `Authorization: Bearer <token>`
- 响应拦截器统一处理错误码：
  - `40101` / `40102` → 清除 token，跳转登录页
  - 其他错误 → `ElMessage.error(message)`
- API 函数命名：`getHosts`、`createHost`、`updateHost`、`deleteHost`
- 文件上传使用 `FormData`，不手动设置 `Content-Type`

### 路由守卫

```javascript
router.beforeEach((to, from, next) => {
    const authStore = useAuthStore()
    if (to.meta.requiresAuth && !authStore.token) {
        next('/login')
    } else if (to.path === '/login' && authStore.token) {
        next('/')
    } else {
        next()
    }
})
```

### 状态管理（Pinia）

- token 存储在 `localStorage`，刷新时恢复
- 登录成功存 token，登出清除
- store 按领域拆分：`auth` / `host` / `cloudResource` / `stats` / `business`

### CSS 规范

- 组件样式使用 `<style scoped>`，全局样式放 `styles/global.css`
- 表格内容居中（`text-align: center`）
- 状态徽章 CSS class 映射（定义在 `styles/global.css`）：

| class | 颜色 |
|-------|------|
| `status-运行中` | 绿色 |
| `status-已停止` | 红色 |
| `status-已关机` | 灰色 |
| `status-待确认` | 橙色 |

- CPU 架构标签（定义在 `components/HostTable.vue` scoped style）：`arch-C86`（蓝）/ `arch-X86`（橙）/ `arch-ARM`（绿）

---

## 提交规范

### 提交信息格式

采用 Conventional Commits 风格，格式：`<type>: <简短描述>`

| type | 用途 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `refactor` | 重构（不改功能） |
| `perf` | 性能优化 |
| `style` | 样式/格式调整 |
| `docs` | 文档变更 |
| `chore` | 构建/依赖/杂项 |
| `test` | 测试相关 |

示例：
```
feat: 新增业务统计页面，按项目/公司/人员聚合资源
fix: 修复ping超时参数单位错误，改用context控制100ms超时
docs: AGENTS.md 补充IP网段管理API
```

### 提交准则

1. **原子提交**：一次提交聚焦一个逻辑变更，不混入无关改动
2. **可编译**：提交前确保 `go build ./...` 通过、前端无语法错误
3. **不提交敏感信息**：`.env`、密钥、密码、Token 一律不入库
4. **不提交产物**：`dist/`、编译二进制、`node_modules/` 通过 `.gitignore` 排除
5. **只在需要时提交**：未经明确要求不主动 commit / push；例外条款与提交号回写时序以 [AGENTS.md](../AGENTS.md) 红线 5 与「回写协议」为准（人类明确要求提交时，验证全绿后可 commit，禁止 push 除非明说）
6. **先看后提**：提交前 `git status` + `git diff` 确认改动范围

### 文档回写文风（防膨胀）

状态文档只保留「可导航的摘要」，细节归 git，避免 PROJECT_STATUS / ROADMAP 无限膨胀：

| 位置 | 上限 | 禁止 |
|------|------|------|
| `PROJECT_STATUS.md` 变更记录 | 每条 **≤ 120 字**，一行摘要：`sha` + type + 任务 ID + 结果 | ①②③ 枚举子改动；粘贴验收全文；重复 commit message |
| `ROADMAP.md` done 备注 | 只写验收结论（命令通过 / 行为确认），**≤ 80 字** | 实现步骤流水账；与变更记录重复的长文 |

- 归档：done 条目 / 变更记录超过 **3 个月**移入汇总行或删除（ROADMAP 归档规则、PROJECT_STATUS 变更记录说明）
- 细节优先级：commit message（完整）> ROADMAP done 备注（验收）> PROJECT_STATUS 变更记录（一行）> AGENTS 快照（无历史）

### 提交前自检清单

- [ ] 改动是否聚焦、无无关文件
- [ ] 后端 `go build ./...` 是否通过
- [ ] 后端 `golangci-lint run ./...` 是否通过
- [ ] 前端 `npm run lint:check` 是否通过（门禁用 `lint:check`；`npm run lint` 带 `--fix` 会改写文件，仅用于本地修复）
- [ ] 是否引入硬编码配置/密钥
- [ ] 数据库模型变更是否已写迁移 SQL
- [ ] API 约定（响应格式、错误码）是否一致
- [ ] 文档（API/数据模型/变更记录）是否同步更新

---

## Linter 与验证门禁

本项目**暂不引入 CI/CD**；所有集成与检验由 agent 在完成任务后**本地执行验证命令**（清单见 [AGENTS.md](../AGENTS.md#验证命令门禁)），全部通过才算任务完成。

提交代码前，必须在本地通过 linter 检查。

### 后端（Go）

```bash
cd backend && golangci-lint run ./...
```

- 配置文件：`backend/.golangci.yml`（**golangci-lint v2 格式**，`version: "2"`；`gosimple` 已并入 `staticcheck`，`gofmt`/`goimports` 移入 `formatters`）
- 启用的检查：`errcheck`、`govet`、`staticcheck`、`unused`、`ineffassign`、`misspell` + 格式化 `gofmt`、`goimports`（`local-prefixes: mcloud`）
- **工具前置**：golangci-lint 安装在 `$(go env GOPATH)/bin`，若该目录不在 `PATH` 需先 `export PATH=$PATH:$(go env GOPATH)/bin`；v1 版本无法读取 v2 配置（会以 `unsupported version of the configuration` 退出）
- 提交前执行：`go build ./...` + `golangci-lint run ./...`

### 前端（Vue/JS）

```bash
cd frontend && npm run lint:check
```

- 配置文件：`frontend/.eslintrc.cjs`
- 两个脚本的区别：`lint:check` = 纯检查（`eslint . --ext .js,.vue`，**门禁用这个**）；`lint` = 同参数 + `--fix` **会改写文件**，仅用于本地批量修复
- 提交前执行：`npm run lint:check` + `npm run build`

---

## 环境一致性

所有开发者使用统一的本地开发环境，**禁止手动安装 PostgreSQL**。

**构建文件约定**：

| 环境 | 文件 | 用途 |
|------|------|------|
| 开发 | `docker/docker-compose.dev.yml` | 日常开发（DB + 单容器应用） |
| 生产 | `docker/docker-compose.prod.yml` | 线上部署（DB + 单容器应用） |

> 本地开发一律使用 `docker-compose.dev.yml`，禁止使用 `docker-compose.prod.yml`。

### 一键启动（开发环境）

```bash
# 1. 准备环境变量
cd docker && cp .env.example .env

# 2. 启动全部服务（数据库 + 应用），首次需构建镜像
docker-compose -f docker-compose.dev.yml up -d

# 仅启动数据库（本地运行后端/前端时使用）
docker-compose -f docker-compose.dev.yml up -d postgres

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f

# 停止服务
docker-compose -f docker-compose.dev.yml down
```

### 架构说明

单容器应用 = Nginx（5678 端口）+ Go 后端（5677 端口），Nginx 负责：
- 静态文件服务（Vue 构建产物）
- `/api` 反向代理到 Go 后端
- WebSocket 代理（IP 探测）

```
浏览器 → :5678 (Nginx) ─┬─ /api/*  → 127.0.0.1:5677 (Go)
                          ├─ /api/ws/* → 127.0.0.1:5677 (Go WebSocket)
                          └─ /*       → 静态文件 (Vue 产物)
```

### 环境变量

- 开发环境：`docker/.env`（从 `docker/.env.example` 复制）
- 生产环境：`docker/.env`（从 `docker/.env.example` 复制，**必须修改 JWT_SECRET**）
- `.env` 不提交到 Git（已在 `.gitignore` 中）

### 生产环境部署

```bash
cd docker
cp .env.example .env   # 复制后修改 JWT_SECRET 等
docker-compose -f docker-compose.prod.yml up -d --build
```

- 应用容器暴露 80 端口（Nginx）
- 后端 `SERVER_MODE=release`，关闭调试日志
- 建议在容器前加 SSL 终端（如 Traefik / Caddy）

---

## 任务协调与变更可见性

多人并行开发时，必须保证变更的可见性，避免冲突。每个任务有明确的「开始」和「结束」动作。

### 任务开始

接到任务后、动手写代码前，必须完成以下步骤：

1. **同步最新代码**：`git checkout main && git pull`
2. **临时需求先入队**：需求不在 ROADMAP 队列时，先按 [AGENTS.md · 会话协议 A.0](../AGENTS.md#会话协议) 的「临时任务入队」新建任务条目并确认验收标准，再继续
3. **创建功能分支**：`git checkout -b feat/<简短描述>` 或 `fix/<简短描述>`（分支步骤细则见 AGENTS.md 协议 A 第 4 步）
4. **确认影响范围**：阅读任务需求，判断涉及哪些模块（后端/前端/数据库），列出可能改动的文件
5. **如有数据模型变更**：先说明新字段/新表的设计，获得确认后再动手
6. **同步环境**：`cd docker && docker-compose -f docker-compose.dev.yml up -d` 确保数据库可用

### 任务进行中

1. 频繁拉取 `main` 最新代码，减少合并冲突
2. 每次提交前检查 `git status`，确认无无关文件混入
3. 修改公共文件（`routes/routes.go`、`models/` 下的模型、各页导航）时，主动通知相关开发者
4. 一个功能分支只做一件事，不混入无关改动
5. **多任务串行合并**：人类一次指定多条任务时，完成任务 N（提交 + 合并 `main` + 回填提交号 + 验证绿）后才开始任务 N+1；N+1 分支从合并后的 `main` 拉出，禁止并行开多任务分支（细则见 AGENTS.md「分支与提交协议」）

### 任务结束

代码完成后、提交 PR 前，必须完成以下步骤：

1. **本地自检**：
   - [ ] 后端 `go build ./...` 通过
   - [ ] 后端 `golangci-lint run ./...` 通过
   - [ ] 前端 `npm run lint:check` 通过
   - [ ] 前端 `npm run build` 通过
   - [ ] 数据库模型变更已写迁移 SQL（`backend/migrations/YYYYMMDDHHMMSS_xxx.sql`）
2. **文档同步**（完整对照见 [AGENTS.md · 文档更新责任矩阵](../AGENTS.md#文档更新责任矩阵)）：
   - 新增/修改 API 接口 → 更新 `docs/API.md`
   - 新增/修改数据库字段 → 更新 `docs/ARCHITECTURE.md` 数据模型章节
   - 新增/删除路由 → 更新 `docs/CODE_INDEX.md`
   - 重要变更 → 更新 `docs/PROJECT_STATUS.md` 变更记录
   - 任务状态 → 更新 `docs/ROADMAP.md`（领任务 → `in_progress`，完成 → `done`；它是任务状态的唯一事实来源）
3. **提交 PR**：
   - PR 描述包含：改了什么、为什么改、影响范围（哪些接口/表/页面受影响）
   - 人类指令明确要求提交时：验证全绿后 `git commit`（禁止 push 除非明说）；提交号回填 ROADMAP `done` 与 PROJECT_STATUS 变更记录，再补 `docs:` 提交并跑绿 `check_docs.sh`（时序见 AGENTS.md「提交号回写时序」）
   - 合并回 `main`，删除功能分支
4. **多任务会话**：合并回 `main` 并删除本任务分支后，方可认领下一条任务；下一条分支从更新后的 `main` 拉出（串行合并，见上）

---

## 测试规范

### 后端测试

- Service 层单元测试：`testing` + 表驱动
- Controller 层集成测试：`httptest` + `gin.CreateTestContext`
- 数据库测试：使用测试库，每个用例前重置数据
- 核心覆盖：认证（登录/锁定/改密）、主机 CRUD（IP 唯一性）、批量操作（跳过重复）、CSV 往返

### 前端测试

- 组件渲染测试：Vitest + Vue Test Utils
- API mock：MSW 或 vitest mock

### 冒烟测试流程

```
登录 → 获取 token → 新增主机 → 查询列表 → 编辑主机 → 删除主机 → 导出 CSV → 导入 CSV → 验证数据
```

---

## 安全准则

- 密码哈希使用 SHA-256 + 盐（格式 `salt$hash`），生产环境建议升级 bcrypt
- JWT 密钥从环境变量读取，禁止硬编码
- `.env` 不提交到 Git
- API 输入必须校验，SQL 通过 GORM 参数化查询防注入
- CORS：开发允许 `localhost:5678`，生产限制具体域名
- 文件上传限制：最大 16MB，仅允许 `.csv` 后缀

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
5. **最小改动**：改动聚焦当前需求，不顺手重构无关代码；未经明确要求，不得修改约束文件（`AGENTS.md`、`docs/*.md`）、构建文件（`Dockerfile`、`docker-compose*.yml`、`.golangci.yml`、`.eslintrc.cjs`、`nginx.conf`）
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

多人同时修改数据模型时，必须通过迁移文件协调，**禁止仅依赖 AutoMigrate**。

**迁移文件规则**：

1. 迁移文件存放在 `backend/migrations/`，命名格式：`YYYYMMDDHHMMSS_<描述>.sql`
2. 每次模型变更（新增字段、改类型、加索引、加约束）必须写对应的 SQL 迁移文件
3. 迁移文件一旦提交，**禁止修改或删除**（已执行的迁移不可变）
4. `database/postgres.go` 中的 `AutoMigrate()` 仅用于本地开发快速验证，**不可作为唯一迁移手段**

**多人协作流程**：

```
1. 拉取最新 develop
2. 创建迁移文件（如 20260322120000_add_status_index.sql）
3. 在本地执行迁移验证
4. 提交 PR（迁移文件 + 模型变更 + 业务代码一起）
5. 合并后其他开发者拉取代码，启动时自动执行新迁移
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
5. **只在需要时提交**：未经明确要求不主动 commit / push
6. **先看后提**：提交前 `git status` + `git diff` 确认改动范围

### 提交前自检清单

- [ ] 改动是否聚焦、无无关文件
- [ ] 后端 `go build ./...` 是否通过
- [ ] 后端 `golangci-lint run ./...` 是否通过
- [ ] 前端 `npm run lint` 是否通过
- [ ] 是否引入硬编码配置/密钥
- [ ] 数据库模型变更是否已写迁移 SQL
- [ ] API 约定（响应格式、错误码）是否一致
- [ ] 文档（API/数据模型/变更记录）是否同步更新

---

## Linter 与 CI 门禁

提交代码前，必须在本地通过 linter 检查。CI 流水线会自动执行以下检查，不通过则拒绝合并。

### 后端（Go）

```bash
cd backend && golangci-lint run ./...
```

- 配置文件：`backend/.golangci.yml`
- 必须通过的规则：`errcheck`、`govet`、`staticcheck`、`unused`、`gosimple`
- 提交前执行：`go build ./...` + `golangci-lint run`

### 前端（Vue/JS）

```bash
cd frontend && npm run lint
```

- 配置文件：`frontend/.eslintrc.cjs`
- 提交前执行：`npm run lint` + `npm run build`

### CI 流水线（PR 触发）

```
PR 提交
  → Lint（Go + JS 并行）
  → Build（后端编译 + 前端构建）
  → 测试（待补充）
  → 全部通过才可合并
```

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

1. **同步最新代码**：`git checkout develop && git pull`
2. **创建功能分支**：`git checkout -b feat/<简短描述>` 或 `fix/<简短描述>`
3. **确认影响范围**：阅读任务需求，判断涉及哪些模块（后端/前端/数据库），列出可能改动的文件
4. **如有数据模型变更**：先说明新字段/新表的设计，获得确认后再动手
5. **同步环境**：`cd docker && docker-compose -f docker-compose.dev.yml up -d` 确保数据库可用

### 任务进行中

1. 频繁拉取 `develop` 最新代码，减少合并冲突
2. 每次提交前检查 `git status`，确认无无关文件混入
3. 修改公共文件（`routes/routes.go`、`models/` 下的模型）时，主动通知相关开发者
4. 一个功能分支只做一件事，不混入无关改动

### 任务结束

代码完成后、提交 PR 前，必须完成以下步骤：

1. **本地自检**：
   - [ ] 后端 `go build ./...` 通过
   - [ ] 后端 `golangci-lint run ./...` 通过
   - [ ] 前端 `npm run lint` 通过
   - [ ] 前端 `npm run build` 通过
   - [ ] 数据库模型变更已写迁移 SQL（`backend/migrations/YYYYMMDDHHMMSS_xxx.sql`）
2. **文档同步**：
   - 新增/修改 API 接口 → 更新 `docs/API.md`
   - 新增/修改数据库字段 → 更新 `docs/ARCHITECTURE.md` 数据模型章节
   - 新增/删除路由 → 更新 `docs/CODE_INDEX.md`
   - 重要变更 → 更新 `docs/PROJECT_STATUS.md` 变更记录
3. **提交 PR**：
   - PR 描述包含：改了什么、为什么改、影响范围（哪些接口/表/页面受影响）
   - 合并回 `develop`，删除功能分支

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

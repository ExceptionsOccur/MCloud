# AGENTS.md

> MCloud 项目总纲。本文件是 AI 协作与开发的**统一入口**，按需跳转到对应文档。

## 项目速览

MCloud 是一个云平台主机资产信息管理系统，采用 **Go (Gin) 后端 + Vue 3 SPA 前端** 架构，数据存储于 PostgreSQL。

系统围绕主机资产提供：主机资产管理、资源统计、IP 统计、业务统计等能力。功能特性与使用说明见 [README.md](./README.md)。

```
浏览器 (Vue 3)  ──Axios──▶  /api        ┐
                ──WS────▶  /api/ws/probe ├─▶  Go (Gin) ─▶ PostgreSQL 16
                                        ┘
```

## 技术栈

| 层次 | 技术 | 说明 |
|------|------|------|
| 后端 | Go 1.22+ / Gin | Web 框架 + 路由 + 中间件 |
| ORM | GORM v2 | 建模 + 查询 + AutoMigrate |
| 数据库 | PostgreSQL 16 | 主数据存储 |
| 前端 | Vue 3 + Vite | SPA |
| UI | Element Plus | 组件库 |
| 状态/路由 | Pinia + Vue Router 4 | |
| HTTP | Axios | API 封装 |
| 图表 | ECharts + vue-echarts | 饼图 / 柱状图 |
| 实时通信 | gorilla/websocket | IP 连通性探测 |
| 认证 | JWT (golang-jwt/v5) | 无状态 Token |
| 密码 | SHA-256 + 盐 | `salt$hash` |

## 文档导航

按需查阅，避免一次性读取全部：

| 文档 | 内容 | 何时查阅 |
|------|------|----------|
| [README.md](./README.md) | 功能特性、快速开始、部署 | 了解系统能做什么 / 启动项目 |
| [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | 目录结构、分层、数据模型（5 张表） | 定位模块 / 改模型 |
| [docs/API.md](./docs/API.md) | 响应格式、错误码、全部路由、参数 | 增改接口 / 调用接口 |
| [docs/BUSINESS_LOGIC.md](./docs/BUSINESS_LOGIC.md) | 认证、搜索、CSV、IP 探测、各类统计逻辑 | 改业务逻辑 |
| [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) | 编码规范、提交规范、测试、安全准则 | 写代码 / 提交前自检 |
| [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) | 文件职责、功能→代码映射、请求链路 | 快速找到实现位置 |
| [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) | 完成度、变更记录、技术债、注意事项 | 了解进度 / 避坑 |

## 核心约定（红线）

以下为必须遵守的硬性约定，细节见 [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)：

1. **分层**：`controllers` 只做参数绑定 + 调用 service + 格式化响应，**不含业务逻辑**；业务逻辑全部在 `services`
2. **统一响应**：所有 API 通过 `controllers.Success` / `controllers.Error` 返回，**禁止** `c.JSON`
3. **配置外置**：数据库连接、密钥等一律走环境变量，**禁止硬编码**
4. **GORM 列名**：缩写字段（如 `CIDR`）必须显式写 `gorm:"column:xxx"`，否则会生成 `c_id_r` 之类的错误列名
5. **提交规范**：Conventional Commits（`feat:` / `fix:` / `docs:` …）；**未经明确要求不主动 commit / push**
6. **敏感信息**：`.env`、密钥、密码、Token **不入库**
7. **改动聚焦**：只改与当前需求相关的代码，不顺手重构；未经明确要求，**不得修改**约束文件（`AGENTS.md`、`docs/*.md`）、构建文件（`Dockerfile`、`docker-compose*.yml`、`.golangci.yml`、`.eslintrc.cjs`、`nginx.conf`）
8. **分支策略**：从 `develop` 拉 feature 分支，完成后合并回 `develop`；**禁止直接 push `main` / `develop`**
9. **数据库迁移**：模型变更必须写迁移 SQL（见 `backend/migrations/`），**禁止仅依赖 AutoMigrate**
10. **CI 门禁**：提交前本地运行 linter + 编译检查（`golangci-lint run` / `npm run lint`），确保通过
11. **环境一致**：本地开发使用 `docker-compose -f docker-compose.dev.yml up -d` 启动，禁止手动安装数据库实例

## 分支与协作

```
main          ← 生产分支，受保护，只接受 PR 合并
  └── develop ← 集成分支，日常开发的目标分支
       ├── feat/xxx   ← 功能分支，从 develop 拉出
       ├── fix/xxx    ← 修复分支
       └── docs/xxx   ← 文档分支
```

**分支规则**：

1. 功能分支命名：`feat/<简短描述>`（如 `feat/business-stats`）、`fix/<简短描述>`、`docs/<简短描述>`
2. 功能分支从 `develop` 拉出，完成后合并回 `develop`
3. `develop` 稳定后合并到 `main` 并打 version tag
4. 禁止直接 push `main` / `develop`
5. 功能分支生命周期不超过 1 周，超时需拆分或重新评估

**变更可见性**：

1. 修改数据模型（`models/`）或 API 路由（`routes/`）时，必须在 PR 描述中说明影响范围
2. 新增/修改 API 接口必须同步更新 `docs/API.md`
3. 新增/修改数据库字段必须同步更新 `docs/ARCHITECTURE.md` 的数据模型章节

## 快速开始

```bash
cd docker && cp .env.example .env && docker-compose -f docker-compose.dev.yml up -d
```

默认账号：`admin` / `Pass4MCloud`。完整说明见 [README.md](./README.md#快速开始)。

## 常用命令

| 目的 | 命令 |
|------|------|
| 后端编译检查 | `cd backend && go build ./...` |
| 后端 Lint | `cd backend && golangci-lint run ./...` |
| 重置 admin 密码 | `cd backend && go run ./cmd/resetpw/main.go` |
| 指定密码重置 | `cd backend && go run ./cmd/resetpw/main.go <新密码>` |
| 前端开发 | `cd frontend && npm run dev` |
| 前端构建 | `cd frontend && npm run build` |
| 前端 Lint | `cd frontend && npm run lint` |
| 启动数据库（仅 DB） | `cd docker && docker-compose -f docker-compose.dev.yml up -d postgres` |
| 开发环境启动 | `cd docker && cp .env.example .env && docker-compose -f docker-compose.dev.yml up -d` |
| 生产环境部署 | `cd docker && cp .env.example .env && docker-compose -f docker-compose.prod.yml up -d --build` |

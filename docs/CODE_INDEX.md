# 代码索引

本文件用于快速定位功能对应的代码位置。目录约定与分层规范见 [AGENTS.md](../AGENTS.md)，开发规范见 [DEVELOPMENT.md](./DEVELOPMENT.md)。

> 定位以**文件/函数名**为准；行号快照见 [ROADMAP.md](./ROADMAP.md) 各任务「定位」字段，以 grep 复核为准。

## 目录

- [后端代码索引](#后端代码索引)
- [前端代码索引](#前端代码索引)
- [功能 → 代码映射](#功能--代码映射)
- [请求链路](#请求链路)
- [数据库表 → 代码](#数据库表--代码)

---

## 后端代码索引

### 入口与配置

| 文件 | 职责 | 关键内容 |
|------|------|----------|
| `backend/main.go` | 程序入口 | `config.Load` → `database.Connect` → `database.Migrate` → `routes.SetupRoutes` → 启动 Gin |
| `backend/config/config.go` | 配置加载 | `Config` 结构体、`Load()`（从环境变量/`.env` 读取） |
| `backend/database/postgres.go` | 数据库连接与迁移 | `Connect()`、`Migrate()`（goose.Up + AutoMigrate 兜底 + seedAdmin）、`seedAdmin()` |
| `backend/migrations/embed.go` | 迁移 SQL embed FS | `//go:embed *.sql` → `migrations.FS`，供 goose 加载 |
| `backend/routes/routes.go` | 路由注册 | `SetupRoutes()`，公开路由 + JWT 鉴权分组 + WebSocket |

### 数据模型（models/）

| 文件 | 模型 | 对应表 |
|------|------|--------|
| `backend/models/user.go` | `User` | `users` |
| `backend/models/host.go` | `Host`（含 `Application` 关联） | `hosts` |
| `backend/models/host_application.go` | `HostApplication` | `host_applications` |
| `backend/models/cloud_resource.go` | `CloudResource` | `cloud_resources` |
| `backend/models/ip_subnet.go` | `IPSubnet` | `ip_subnets` |
| `backend/models/person.go` | `Person` | `persons` |
| `backend/models/zero_trust.go` | `ZeroTrust` | `zero_trusts` |
| `backend/models/domain.go` | `Domain` | `domains` |
| `backend/models/public_ip.go` | `PublicIP` | `public_ips` |

### 控制器（controllers/）

> 仅做参数绑定/校验 + 调用 service + 格式化响应，不含业务逻辑。

| 文件 | 职责 |
|------|------|
| `controllers/response.go` | 统一响应封装：`Success()` / `Error()` + `Response` 结构 |
| `controllers/auth.go` | 登录、登出、获取用户信息、修改密码 |
| `controllers/host.go` | 主机 CRUD（`Filter`/`Get`/`Create`/`Update`/`Delete`） |
| `controllers/batch.go` | 批量添加、批量编辑 |
| `controllers/csv.go` | CSV 导入、导出、下载模板 |
| `controllers/cloud_resource.go` | 云资源列表、按区域更新 |
| `controllers/stats.go` | IP 使用情况、单点探测、业务统计 |
| `controllers/subnet.go` | IP 网段 CRUD |
| `controllers/person.go` | 人员 CRUD（`List`/`Create`/`Update`/`Delete`） |
| `controllers/zero_trust.go` | 零信任台账 CRUD |
| `controllers/domain.go` | 域名台账 CRUD |
| `controllers/public_ip.go` | 公网IP资源台账 CRUD |
| `controllers/websocket.go` | WebSocket 探测通道（并发处理探测请求） |

### 业务服务（services/）

| 文件 | 职责 | 关键函数 |
|------|------|----------|
| `services/auth_service.go` | 认证逻辑、失败锁定 | `Login`（5 次失败锁定 15 分钟）、`ChangePassword` |
| `services/host_service.go` | 主机 CRUD、筛选、批量、CSV 行解析 | `Filter`、`Create`、`BatchUpdate`、`ParseCSVRowToCreateHost` |
| `services/cloud_resource_service.go` | 云资源总览读写 | `List`、`Update`（按 region upsert） |
| `services/stats_service.go` | IP 使用统计 + 连通性探测 | `GetIPUsage`、`Probe`、`pingICMP`、`probeTCP` |
| `services/business_stats.go` | 业务统计聚合 | `GetBusinessStats`（项目/公司/人员三维聚合） |
| `services/subnet_service.go` | IP 网段管理 | `List`/`Create`/`Update`/`Delete`、`normalizeCIDR`（/24 校验） |
| `services/person_service.go` | 人员管理 | `List`/`Create`/`Update`/`Delete`、`normalizePerson`；删除时校验关联主机数 |
| `services/zero_trust_service.go` | 零信任台账 | `List`/`Create`/`Update`/`Delete`/`BatchCreateText`；`HostReferencedByZeroTrust` |
| `services/domain_service.go` | 域名台账 | `List`/`Create`/`Update`/`Delete`/`BatchCreateText` |
| `services/public_ip_service.go` | 公网IP资源台账 | `List`/`Create`/`Update`/`Delete`、`normalizePublicIP`（IP 唯一/格式校验） |

### 中间件与工具

| 文件 | 职责 |
|------|------|
| `middleware/jwt.go` | `JWTAuth()` JWT 鉴权、`CORSMiddleware()` 跨域 |
| `utils/password.go` | 密码哈希：`HashPassword`、`FormatPasswordHash`、`VerifyPassword` |
| `utils/csv.go` | CSV 解析/生成、多编码检测 |

### 命令工具（cmd/）

| 文件 | 职责 |
|------|------|
| `cmd/resetpw/main.go` | admin 密码重置（随机高熵 / 指定密码），同时清零失败次数与锁定 |

---

## 前端代码索引

### 入口与路由

| 文件 | 职责 |
|------|------|
| `frontend/src/main.js` | 注册 Pinia、Vue Router、Element Plus、图标、全局样式 |
| `frontend/src/App.vue` | 根组件（`<router-view />`） |
| `frontend/src/router/index.js` | 路由配置 + 登录守卫 |
| `frontend/vite.config.js` | 开发代理 `/api` → 后端（`ws: true` 支持 WebSocket） |

### API 封装（api/）

| 文件 | 对应后端接口 |
|------|--------------|
| `api/index.js` | Axios 实例 + 请求/响应拦截器（注入 token、401 跳登录） |
| `api/auth.js` | `/auth/login`、`/auth/me`、`/auth/password` |
| `api/host.js` | `/hosts/*`、`/batch/hosts` |
| `api/csv.js` | `/import`、`/export`、`/template` |
| `api/cloud_resource.js` | `/cloud-resources` |
| `api/stats.js` | `/stats/ip-usage`、`/stats/probe`、`/stats/business` |
| `api/subnet.js` | `/ip-subnets` |
| `api/person.js` | `/persons` |
| `api/zero_trust.js` | `/zero-trusts` |
| `api/domain.js` | `/domains` |
| `api/public_ip.js` | `/public-ips` |

### 状态管理（stores/）

| 文件 | 职责 |
|------|------|
| `stores/auth.js` | token、用户信息、登录/登出 |
| `stores/host.js` | 主机列表、分页、筛选条件 |
| `stores/cloudResource.js` | 云资源 + 按区域聚合的已用/运行中/已停止统计 |
| `stores/stats.js` | IP 使用情况（网段 + 已用/空记录 IP） |
| `stores/business.js` | 业务统计数据（三维聚合结果） |

### 页面（views/）

| 文件 | 路由 | 职责 |
|------|------|------|
| `views/Login.vue` | `/login` | 登录页 |
| `views/HostManagement.vue` | `/` | 主机管理主页（工具栏 + 表格 + 各弹窗） |
| `views/ResourceStatistics.vue` | `/statistics` | 资源统计（汇总表格 + 饼图） |
| `views/IpStatistics.vue` | `/ip-statistics` | IP 统计（位图 + 探测 + 网段管理） |
| `views/BusinessStatistics.vue` | `/business-statistics` | 业务统计（概览 + 柱状图 + 表格 + 下钻弹窗） |
| `views/PersonnelManagement.vue` | `/personnel` | 人员管理（搜索 + 新增/编辑/删除，含关联主机数校验） |
| `views/PublicIPManagement.vue` | `/public-ip` | 公网IP资源台账（搜索 + 新增/编辑/删除，设置菜单入口） |
| `views/ZeroTrustLedger.vue` | `/zero-trust` | 零信任台账（搜索 + 新增/编辑/删除 + 批量添加，申请主机外键） |
| `views/DomainLedger.vue` | `/domain-ledger` | 域名台账（搜索 + 新增/编辑/删除 + 批量添加） |

### 组件（components/）

| 文件 | 职责 | 使用页面 |
|------|------|----------|
| `SearchToolbar.vue` | 搜索框、筛选下拉、操作按钮 | HostManagement |
| `HostTable.vue` | 主机列表表格 + 分页 | HostManagement |
| `HostFormDialog.vue` | 新增/编辑/详情弹窗（技术+申请双 Tab，含人员选择/手输自动建人员） | HostManagement |
| `BatchAddDialog.vue` | 批量添加弹窗 | HostManagement |
| `BatchEditDialog.vue` | 批量编辑弹窗（含人员选择/手输自动建人员） | HostManagement |
| `ImportDialog.vue` | CSV 导入弹窗 | HostManagement |
| `ChangePasswordDialog.vue` | 修改密码弹窗 | 全部页面 |
| `CloudResourceDialog.vue` | 云资源录入弹窗 | 全部页面（设置菜单） |
| `SubnetManageDialog.vue` | IP 网段管理弹窗 | IpStatistics |

### 工具与样式

| 文件 | 职责 |
|------|------|
| `utils/index.js` | `downloadBlob`、`formatTime`、各下拉选项常量 |
| `styles/global.css` | 全局样式、状态/标签 CSS class |

---

## 仓库工具

| 文件 | 职责 |
|------|------|
| `scripts/check_docs.sh` | 文档一致性校验 5 项（links/tree/snapshot/counts/shas），AGENTS「验证命令」文档改动项，支持 `--only`/`--list` |

---

## 功能 → 代码映射

| 功能 | 后端 | 前端 |
|------|------|------|
| **登录/认证** | `controllers/auth.go` `services/auth_service.go` `middleware/jwt.go` | `views/Login.vue` `stores/auth.js` `api/auth.js` |
| **主机 CRUD** | `controllers/host.go` `services/host_service.go` | `views/HostManagement.vue` `components/HostTable.vue` `components/HostFormDialog.vue` `stores/host.js` |
| **主机筛选/搜索** | `services/host_service.go` → `Filter` | `components/SearchToolbar.vue` |
| **批量增改** | `controllers/batch.go` `services/host_service.go` → `BatchCreate`/`BatchUpdate` | `components/BatchAddDialog.vue` `components/BatchEditDialog.vue` |
| **CSV 导入导出** | `controllers/csv.go` `services/host_service.go` `utils/csv.go` | `components/ImportDialog.vue` `api/csv.js` |
| **云资源录入** | `controllers/cloud_resource.go` `services/cloud_resource_service.go` | `components/CloudResourceDialog.vue` `stores/cloudResource.js` |
| **资源统计** | `controllers/cloud_resource.go` + `GET /api/hosts` | `views/ResourceStatistics.vue` `stores/cloudResource.js` |
| **IP 使用位图** | `controllers/stats.go` → `IPUsage` `services/stats_service.go` → `GetIPUsage` | `views/IpStatistics.vue` `stores/stats.js` |
| **IP 连通性探测** | `controllers/websocket.go` `controllers/stats.go` → `Probe` `services/stats_service.go` → `Probe` | `views/IpStatistics.vue`（WebSocket + 颜色状态机） |
| **IP 网段管理** | `controllers/subnet.go` `services/subnet_service.go` `models/ip_subnet.go` | `components/SubnetManageDialog.vue` `api/subnet.js` |
| **人员管理** | `controllers/person.go` `services/person_service.go` `models/person.go` | `views/PersonnelManagement.vue` `api/person.js` |
| **公网IP资源录入** | `controllers/public_ip.go` `services/public_ip_service.go` `models/public_ip.go` | `views/PublicIPManagement.vue` `api/public_ip.js` |
| **零信任台账** | `controllers/zero_trust.go` `services/zero_trust_service.go` `models/zero_trust.go` | `views/ZeroTrustLedger.vue` `api/zero_trust.js` `components/ZeroTrustBatchAddDialog.vue` |
| **域名台账** | `controllers/domain.go` `services/domain_service.go` `models/domain.go` | `views/DomainLedger.vue` `api/domain.js` `components/DomainBatchAddDialog.vue` |
| **主机关联人员** | `services/host_service.go`（`person_id` 事务写入） | `components/HostFormDialog.vue` `components/BatchEditDialog.vue`（选择/手输自动新增） |
| **业务统计** | `controllers/stats.go` → `BusinessStats` `services/business_stats.go` | `views/BusinessStatistics.vue` `stores/business.js` |
| **密码重置工具** | `cmd/resetpw/main.go` `utils/password.go` | — |

---

## 请求链路

### 主机列表查询

```
GET /api/hosts
  → routes/routes.go（JWT 鉴权）
  → controllers/host.go           Filter()
  → services/host_service.go      Filter()（动态 WHERE + Preload Application + 分页）
  → models/host.go                返回 []Host
```

### IP 单点探测（WebSocket）

```
前端点击格子
  → ws://.../api/ws/probe?token=xxx
  → routes/routes.go（校验 query token）
  → controllers/websocket.go      HandleProbeWS() → handleProbeMessage()（每请求独立 goroutine）
  → services/stats_service.go     Probe()（ICMP + TCP22/3389 → 颜色状态机，仅返回颜色不写库）
  → 回推 { ip, color }
```

### 业务统计聚合

```
GET /api/stats/business
  → controllers/stats.go          BusinessStats()
  → services/business_stats.go     GetBusinessStats()
       ├─ 读取全部 Host + Preload Application
       ├─ 按 project / company / applicant 三维聚合
       └─ 汇总资源、运行状态、维度关联、主机明细
```

### 云资源保存

```
PUT /api/cloud-resources
  → controllers/cloud_resource.go     Update()
  → services/cloud_resource_service.go Update()（按 region 查询，存在则更新，否则创建）
```

---

## 数据库表 → 代码

| 表 | 模型 | 主要操作位置 |
|----|------|--------------|
| `users` | `models/user.go` | `services/auth_service.go`、`cmd/resetpw/main.go` |
| `hosts` | `models/host.go` | `services/host_service.go`、`services/stats_service.go`、`services/business_stats.go` |
| `host_applications` | `models/host_application.go` | `services/host_service.go`（随主机事务创建/更新/删除） |
| `cloud_resources` | `models/cloud_resource.go` | `services/cloud_resource_service.go` |
| `ip_subnets` | `models/ip_subnet.go` | `services/subnet_service.go`、`services/stats_service.go`（读取网段） |
| `persons` | `models/person.go` | `services/person_service.go`；`hosts.person_id` 由 `services/host_service.go` 写入 |
| `zero_trusts` | `models/zero_trust.go` | `services/zero_trust_service.go`；删除主机前经 `HostReferencedByZeroTrust` 校验 |
| `domains` | `models/domain.go` | `services/domain_service.go` |
| `public_ips` | `models/public_ip.go` | `services/public_ip_service.go` |

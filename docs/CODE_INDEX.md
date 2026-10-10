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
| `backend/routes/routes.go` | 路由协调 | `SetupRoutes()`：CORS + JWT 受保护组统一收口 + 分发各域 |
| `backend/routes/{auth,host,data_exchange,cloud_resource,stats,subnet,person,public_ip,zero_trust,port_mapping,audit,websocket}.go` | 各域路由注册 | `registerX(g)`；公开域仅登录/登出 + WS probe |

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
| `backend/models/port_mapping.go` | `PortMapping` | `port_mappings` |
| `backend/models/public_ip.go` | `PublicIP` | `public_ips` |
| `backend/models/audit_log.go` | `AuditLog` | `audit_logs` |

### 控制器（controllers/）

> 仅做参数绑定/校验 + 调用 service + 格式化响应，不含业务逻辑。

| 文件 | 职责 |
|------|------|
| `controllers/response.go` | 统一响应封装：`Success()` / `Error()` + `Response` 结构 |
| `controllers/auth.go` | 登录、登出、获取用户信息、修改密码 |
| `controllers/host.go` | 主机 CRUD（`Filter`/`Get`/`Create`/`Update`/`Delete`） |
| `controllers/batch.go` | 批量添加、批量编辑 |
| `controllers/csv.go` | CSV 导入、导出、下载模板 |
| `controllers/data_exchange.go` | 数据备份：8 表 xlsx 导出（`Export`）、导入（`Import`，20MB 限制 + 错误码映射） |
| `controllers/cloud_resource.go` | 云资源列表、按区域更新 |
| `controllers/stats.go` | IP 使用情况、单点探测、业务统计 |
| `controllers/subnet.go` | IP 网段 CRUD |
| `controllers/person.go` | 人员 CRUD（`List`/`Create`/`Update`/`Delete`） |
| `controllers/zero_trust.go` | 零信任台账 CRUD |
| `controllers/port_mapping.go` | 端口映射台账 CRUD + 整组防重判定（`ErrPortMappingExists` → `40901`） |
| `controllers/public_ip.go` | 公网IP资源台账 CRUD + 被引用校验（改IP/删除被映射/零信任引用 → `40901`） |
| `controllers/audit_log.go` | 审计日志查询（分页筛选） |
| `controllers/audit_context.go` | 审计上下文：`auditContext(c)` 取 operator + request_id |
| `controllers/websocket.go` | WebSocket 探测通道（并发处理探测请求） |

### 业务服务（services/）

| 文件 | 职责 | 关键函数 |
|------|------|----------|
| `services/auth_service.go` | 认证逻辑、失败锁定 | `Login`（5 次失败锁定 15 分钟）、`ChangePassword` |
| `services/host_service.go` | 主机 CRUD 与筛选 | `Filter`、`Create`、`Update`、`Delete`、`validatePersonID` |
| `services/host_batch_service.go` | 主机批量增改（结构化 JSON + 纯文本行）与批量编辑 | `BatchCreate`、`BatchCreateFromText`、`BatchUpdate`、`parsePersonID` |
| `services/host_csv_service.go` | 主机 CSV 导入导出流程（后缀/编码/行循环）与行映射 | `ImportCSV`、`ExportCSV`、`CSVTemplate`、`ParseCSVRowToCreateHost`、`ExportToCSVRows` |
| `services/cloud_resource_service.go` | 云资源总览读写 | `List`、`Update`（按 region upsert） |
| `services/stats_service.go` | IP 使用统计 + 连通性探测 | `GetIPUsage`、`Probe`、`pingICMP`、`probeTCP` |
| `services/business_stats.go` | 业务统计聚合 | `GetBusinessStats`（项目/公司/人员三维聚合） |
| `services/subnet_service.go` | IP 网段管理 | `List`/`Create`/`Update`/`Delete`、`normalizeCIDR`（/24 校验） |
| `services/person_service.go` | 人员管理 | `List`/`Create`/`Update`/`Delete`、`normalizePerson`；删除时校验关联主机数 |
| `services/zero_trust_service.go` | 零信任台账 | `List`/`Create`/`Update`/`Delete`/`BatchCreateText`；`normalizeTargetPairs`（配对校验）；`HostReferencedByZeroTrust`/`ZeroTrustsByHost` |
| `services/port_mapping_service.go` | 端口映射台账 | `List`/`Create`/`Update`/`Delete`/`BatchCreateText`；`ErrPortMappingExists` 整组判重（`host_id+public_ip+external_ports`）；`refreshHostIPMapped` |
| `services/public_ip_service.go` | 公网IP资源台账 | `List`/`Create`/`Update`/`Delete`、`normalizePublicIP`（IP 唯一/格式校验）、`countPublicIPRefs`/`ErrPublicIPReferenced` 引用保护 |
| `services/audit_service.go` | 审计日志 | `Record`（写入，失败不阻断业务）、`List`（分页筛选）、`GetOperator`、`NewRequestID` |
| `services/data_exchange.go` | 数据备份公共定义 | `dataExchangeSheets`（8 sheet + 中文列头）、`ImportReport`/`SheetStat`/`RowError` |
| `services/export_service.go` | 数据备份导出 | `Export`（8 sheet 依赖序写入）、`ExportFilename` |
| `services/import_service.go` | 数据备份导入 | `Import`（单事务 + 回滚）、`parseWorkbook`（结构校验）、`apply*`（按自然键 upsert） |

### 中间件与工具

| 文件 | 职责 |
|------|------|
| `middleware/jwt.go` | `JWTAuth()` JWT 鉴权、`CORSMiddleware()` 跨域 |
| `utils/password.go` | 密码哈希：`HashPassword`、`FormatPasswordHash`、`VerifyPassword` |
| `utils/csv.go` | CSV 解析/生成、多编码检测 |
| `utils/datetime.go` | 日期时间统一口径：`ValidateDateOnly`（严格 `YYYY-MM-DD`）、`ParseFlexibleTime`/`FlexibleTime`（RFC3339 为主 + 无时区兼容）、展示常量 `DateTimeLayout` |

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
| `api/dataExchange.js` | `/export/all`、`/import/all` |
| `api/cloud_resource.js` | `/cloud-resources` |
| `api/stats.js` | `/stats/ip-usage`、`/stats/probe`、`/stats/business` |
| `api/subnet.js` | `/ip-subnets` |
| `api/person.js` | `/persons` |
| `api/zero_trust.js` | `/zero-trusts` |
| `api/port_mapping.js` | `/port-mappings` |
| `api/public_ip.js` | `/public-ips` |
| `api/audit.js` | `/audit-logs` |

### 状态管理（stores/）

| 文件 | 职责 |
|------|------|
| `stores/auth.js` | token、用户信息、登录/登出 |
| `stores/host.js` | 主机列表 CRUD + 分页态（分页/筛选通用逻辑由 `composables/usePagedTable` 提供） |
| `stores/cloudResource.js` | 云资源 + 按区域聚合的已用/运行中/已停止统计 |
| `stores/stats.js` | IP 使用情况（网段 + 已用/空记录 IP） |
| `stores/business.js` | 业务统计数据（三维聚合结果） |

### 组合式函数（composables/，T-004）

| 文件 | 职责 |
|------|------|
| `composables/useProbeWebSocket.js` | 探测 WS 连接管理：连接/3s 重连/发送/卸载清理，`onMessage` 回调交回调用方处理业务帧 |
| `composables/useChartOption.js` | ECharts 按需注册（CanvasRenderer 内置）+ 返回 `VChart`（模板 `<v-chart>`） |
| `composables/usePagedTable.js` | 分页列表通用态：页码/页大小/筛选/加载/总数、`fetchPage`/`setFilter`/`setPage`/`resetFilters` |
| `composables/useProbeGrid.js` | 位图探测状态机（T-005）：颜色/单点探测/批量测试/网段计数，`onUnmounted` 清理 resolver |

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
| `views/ZeroTrustLedger.vue` | `/zero-trust` | 零信任台账（搜索 + 「申请资源」标签列 `主机名(ip:port)` 按主机名排序 + 配对行表单 + 公网IP资源池带出接入地区 + 批量添加） |
| `views/MappingLedger.vue` | `/mapping-ledger` | 端口映射台账（多端口标签 + 批量添加） |
| `views/DataBackup.vue` | `/data-backup` | 数据备份（8 表 xlsx 导出/导入 + 导入报告展示，设置菜单入口） |
| `views/AuditLog.vue` | `/audit-logs` | 审计日志（操作人/动作/资源筛选 + 分页 + 详情弹窗） |

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
| `CloudResourceDialog.vue` | 云资源录入弹窗 | 除 DataBackup 外的设置菜单页面 |
| `ip/SubnetManageDialog.vue` | IP 网段管理弹窗 | IpStatistics |
| `ip/ProbePanel.vue` | 探测服务连接状态 + 「管理网段」入口（T-005） | IpStatistics |
| `ip/SubnetCard.vue` | 单网段卡片：统计/批量进度/位图格子（T-005） | IpStatistics |
| `ZeroTrustBatchAddDialog.vue` | 零信任批量添加弹窗（文本粘贴，9 列） | ZeroTrustLedger |
| `MappingBatchAddDialog.vue` | 端口映射批量添加弹窗（文本粘贴，6 列） | MappingLedger |

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
| **审计日志** | `controllers/audit_log.go` `services/audit_service.go` `models/audit_log.go` `routes/audit.go` | `views/AuditLog.vue` `api/audit.js` |
| **主机 CRUD** | `controllers/host.go` `services/host_service.go` | `views/HostManagement.vue` `components/HostTable.vue` `components/HostFormDialog.vue` `stores/host.js` `composables/usePagedTable.js` |
| **主机筛选/搜索** | `services/host_service.go` → `Filter` | `components/SearchToolbar.vue` |
| **批量增改** | `controllers/batch.go` `services/host_batch_service.go` → `BatchCreate`/`BatchUpdate` | `components/BatchAddDialog.vue` `components/BatchEditDialog.vue` |
| **CSV 导入导出** | `controllers/csv.go` `services/host_csv_service.go` `utils/csv.go` | `components/ImportDialog.vue` `api/csv.js` |
| **数据备份 xlsx 导出/导入** | `controllers/data_exchange.go` `services/export_service.go` `services/import_service.go` | `views/DataBackup.vue` `api/dataExchange.js` |
| **云资源录入** | `controllers/cloud_resource.go` `services/cloud_resource_service.go` | `components/CloudResourceDialog.vue` `stores/cloudResource.js` |
| **资源统计** | `controllers/cloud_resource.go` + `GET /api/hosts` | `views/ResourceStatistics.vue` `stores/cloudResource.js` `composables/useChartOption.js` |
| **IP 使用位图** | `controllers/stats.go` → `IPUsage` `services/stats_service.go` → `GetIPUsage` | `views/IpStatistics.vue` `components/ip/SubnetCard.vue` `stores/stats.js` |
| **IP 连通性探测** | `controllers/websocket.go` `controllers/stats.go` → `Probe` `services/stats_service.go` → `Probe` | `composables/useProbeGrid.js`（颜色状态机/批量） `composables/useProbeWebSocket.js`（连接/重连/发送） `components/ip/ProbePanel.vue`（状态指示） |
| **IP 网段管理** | `controllers/subnet.go` `services/subnet_service.go` `models/ip_subnet.go` | `components/ip/SubnetManageDialog.vue` `api/subnet.js` |
| **人员管理** | `controllers/person.go` `services/person_service.go` `models/person.go` | `views/PersonnelManagement.vue` `api/person.js` |
| **公网IP资源录入** | `controllers/public_ip.go` `services/public_ip_service.go` `models/public_ip.go` | `views/PublicIPManagement.vue` `api/public_ip.js` |
| **公网IP引用保护 + 映射整组防重** | `services/public_ip_service.go` → `countPublicIPRefs`/`ErrPublicIPReferenced`（`40901`）、`services/port_mapping_service.go` → `ErrPortMappingExists`（`40901`）、`migrations/20261010000004` 唯一索引 | `api/public_ip.js` `api/port_mapping.js`（错误码提示） |
| **零信任台账** | `controllers/zero_trust.go` `services/zero_trust_service.go` `models/zero_trust.go` | `views/ZeroTrustLedger.vue` `api/zero_trust.js` `components/ZeroTrustBatchAddDialog.vue` |
| **端口映射台账** | `controllers/port_mapping.go` `services/port_mapping_service.go` `models/port_mapping.go` | `views/MappingLedger.vue` `api/port_mapping.js` `components/MappingBatchAddDialog.vue` |
| **主机关联人员** | `services/host_service.go`（`person_id` 事务写入） | `components/HostFormDialog.vue` `components/BatchEditDialog.vue`（选择/手输自动新增） |
| **业务统计** | `controllers/stats.go` → `BusinessStats` `services/business_stats.go` | `views/BusinessStatistics.vue` `stores/business.js` `composables/useChartOption.js` |
| **密码重置工具** | `cmd/resetpw/main.go` `utils/password.go` | — |

---

## 请求链路

> 路由按域拆分（T-010）：JWT 统一挂载收口于 `routes/routes.go` 受保护组，各域注册在 `routes/<域>.go`；下文链路行标注域注册文件。

### 主机列表查询

```
GET /api/hosts
  → routes/host.go                 （JWT，收口 routes.go）
  → controllers/host.go           Filter()
  → services/host_service.go      Filter()（动态 WHERE + applicant_empty 分支 + Preload Application/Person + 分页）
  → models/host.go                返回 []Host
```

### IP 单点探测（WebSocket）

```
前端点击格子
  → ws://.../api/ws/probe?token=xxx
  → routes/websocket.go            （校验 query token）
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

### CSV 导入导出

```
POST /api/import（multipart file）
  → routes/host.go                 （JWT，收口 routes.go）
  → controllers/csv.go               Import()（绑定文件 + 读入字节 + 错误码映射）
  → services/host_csv_service.go     ImportCSV()（后缀校验 + DetectAndDecode 识别编码 + ParseCSV + 行循环）
  → services/host_csv_service.go     ParseCSVRowToCreateHost() → Create()（成功/跳过/错误计数）

GET /api/export
  → routes/host.go                 （JWT，收口 routes.go）
  → controllers/csv.go               Export()
  → services/host_csv_service.go     ExportCSV()（Preload Application 读全表）
  → services/host_csv_service.go     ExportToCSVRows() → utils/csv.go BuildCSVOutput()（BOM + CSVHeaders 表头）
```

### 零信任台账 CRUD

```
GET/POST /api/zero-trusts、PUT/DELETE /api/zero-trusts/:id、POST /api/zero-trusts/batch
  → routes/zero_trust.go           （JWT，收口 routes.go）
  → controllers/zero_trust.go        List()/Create()/Update()/Delete()/BatchCreateText()
  → services/zero_trust_service.go
       ├─ List()（keyword 模糊 + parseStoredTargets 回填主机简要 + 公网IP资源池带出接入地区）
       ├─ Create()/Update()（normalizeTargetPairs 配对校验 + 主机存在校验 + 公网IP须在资源池）
       └─ BatchCreateText()（逐行 9 列：内网IP 定位主机、IP 与端口等长配对、合法行全量插入）
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
| `zero_trusts` | `models/zero_trust.go` | `services/zero_trust_service.go`；删除主机前经 `HostReferencedByZeroTrust` 校验；主机详情经 `ZeroTrustsByHost` 回填 |
| `port_mappings` | `models/port_mapping.go` | `services/port_mapping_service.go`；删除主机前经 `HostReferencedByMapping` 校验 |
| `public_ips` | `models/public_ip.go` | `services/public_ip_service.go` |
| `audit_logs` | `models/audit_log.go` | `services/audit_service.go`；各 CUD service 方法经 `Record` 写入；`controllers/audit_log.go` 查询 |

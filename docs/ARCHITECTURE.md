# 架构与数据模型

> 本文档描述 MCloud 的目录结构、分层架构与数据库模型。
> 文档入口见 [AGENTS.md](../AGENTS.md)，代码位置见 [CODE_INDEX.md](./CODE_INDEX.md)。

## 目录

- [整体架构](#整体架构)
- [目录结构](#目录结构)
- [数据模型](#数据模型)

---

## 整体架构

```
┌─────────────────────────────┐
│  浏览器 (Vue 3 SPA)          │
│  Element Plus + Pinia        │
│  Axios ──┐        WebSocket │
└──────────┼────────────┬──────┘
           │ /api       │ /api/ws/probe
           ▼            ▼
┌─────────────────────────────┐
│  Vite Dev Server (代理)       │
└──────────┬──────────────────┘
           ▼
┌─────────────────────────────┐
│  Go 后端 (Gin)               │
│  routes → middleware         │
│    → controllers             │
│    → services                │
│    → models (GORM)           │
└──────────┬──────────────────┘
           ▼
┌─────────────────────────────┐
│  PostgreSQL 16               │
└─────────────────────────────┘
```

- **前端**：单页应用，`/api` 请求经 Axios，IP 探测走 WebSocket 长连接
- **后端**：分层架构，启动时 goose 执行 `migrations/*.sql`（embed）+ AutoMigrate 兜底 + 种子数据
- **数据库**：PostgreSQL，8 张表

---

## 目录结构

> 快照日期：2026-10-06（由 `scripts/check_docs.sh` 的 tree 检查守护：树中文件必须存在、已跟踪文件必须登记）

```
go/
├── README.md                      # 功能特性与使用说明
├── AGENTS.md                      # 项目总纲（统一入口）
├── scripts/                       # 仓库自动化脚本
│   └── check_docs.sh              # 文档一致性校验（链接/目录树/状态快照/计数/提交号）
│
├── docs/                          # 工程文档
│   ├── ROADMAP.md                 # 任务队列（任务状态唯一事实来源）
│   ├── DEVELOPMENT.md             # 开发规范与准则
│   ├── ARCHITECTURE.md            # 架构与数据模型（本文件）
│   ├── API.md                     # API 约定
│   ├── BUSINESS_LOGIC.md          # 关键业务逻辑
│   ├── CODE_INDEX.md              # 代码索引
│   └── PROJECT_STATUS.md          # 功能完成度、变更记录与技术债
│
├── docker/                        # 容器化配置
│   ├── docker-compose.dev.yml     # 开发环境（DB + 后端热重载 + 前端 HMR）
│   ├── docker-compose.prod.yml    # 生产环境（DB + 单容器应用）
│   └── .env.example               # 环境变量模板（复制为 .env 后使用）
│
├── backend/                       # Go 后端 + Docker 构建
│   ├── main.go                    # 入口：启动 Gin、连接 DB、运行迁移、注册路由
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile                 # 多阶段构建（Node 前端 + Go 后端 → Alpine 运行）
│   ├── Dockerfile.dev             # 开发镜像（air 热重载）
│   ├── .golangci.yml              # golangci-lint v2 配置
│   ├── nginx.conf                 # Nginx 配置（SPA + /api 反代 + WebSocket）
│   ├── run.sh                     # 容器启动脚本（Nginx + Go）
│   │
│   ├── config/
│   │   └── config.go              # 配置结构体（从环境变量加载）
│   │
│   ├── database/
│   │   └── postgres.go            # GORM 连接初始化 + 迁移执行
│   │
│   ├── models/
│   │   ├── user.go                # User 模型
│   │   ├── host.go                # Host 模型
│   │   ├── host_application.go    # HostApplication 模型
│   │   ├── person.go              # Person 模型（人员信息）
│   │   ├── cloud_resource.go      # CloudResource 模型
│   │   ├── ip_subnet.go           # IPSubnet 模型（IP 网段管理）
│   │   ├── zero_trust.go          # ZeroTrust 模型（零信任台账）
│   │   └── domain.go              # Domain 模型（域名台账）
│   │
│   ├── controllers/
│   │   ├── response.go            # 统一响应辅助函数
│   │   ├── auth.go                # 登录、登出、改密码、用户信息
│   │   ├── host.go                # 主机 CRUD
│   │   ├── batch.go               # 批量添加/编辑
│   │   ├── csv.go                 # CSV 导入/导出/模板
│   │   ├── cloud_resource.go      # 云资源总览
│   │   ├── person.go              # 人员 CRUD
│   │   ├── zero_trust.go          # 零信任台账 CRUD
│   │   ├── domain.go              # 域名台账 CRUD
│   │   ├── stats.go               # 统计（IP 使用、探测、业务统计）
│   │   ├── subnet.go              # IP 网段 CRUD
│   │   └── websocket.go           # WebSocket IP 探测通道
│   │
│   ├── middleware/
│   │   └── jwt.go                 # JWT 认证中间件 + CORS 中间件
│   │
│   ├── services/
│   │   ├── auth_service.go        # 认证业务逻辑
│   │   ├── host_service.go        # 主机业务逻辑（CRUD + 批量 + CSV，待拆分）
│   │   ├── cloud_resource_service.go # 云资源总览业务逻辑
│   │   ├── person_service.go      # 人员业务逻辑
│   │   ├── zero_trust_service.go  # 零信任台账业务逻辑
│   │   ├── domain_service.go      # 域名台账业务逻辑
│   │   ├── stats_service.go       # IP 使用统计 + 连通性探测
│   │   ├── business_stats.go      # 业务统计聚合（项目/公司/人员）
│   │   └── subnet_service.go      # IP 网段业务逻辑（/24 校验）
│   │
│   ├── routes/
│   │   └── routes.go              # 路由注册（公开 + 鉴权 + WebSocket）
│   │
│   ├── cmd/
│   │   └── resetpw/main.go        # 密码重置工具（随机/指定密码）
│   │
│   ├── utils/
│   │   ├── password.go            # 密码哈希工具（SHA-256 + 盐）
│   │   └── csv.go                 # CSV 解析/生成
│   │
│   └── migrations/                # 迁移 SQL（启动时 goose 执行，embed 进二进制；AutoMigrate 兜底）
│
└── frontend/                      # Vue 3 前端
    ├── index.html
    ├── package.json
    ├── package-lock.json
    ├── .eslintrc.cjs              # ESLint 配置（eslint:recommended + vue 插件）
    ├── .eslintignore              # ESLint 忽略：dist/、node_modules/
    ├── vite.config.js             # 代理 /api 到后端（ws: true 支持 WebSocket）
    │
    └── src/
        ├── main.js                # 入口：注册 Vue Router、Pinia、Element Plus
        ├── App.vue                # 根组件（路由出口）
        │
        ├── router/
        │   └── index.js           # 路由配置 + 路由守卫
        │
        ├── stores/                # Pinia 状态：auth / host / cloudResource / stats / business
        ├── api/                   # API 封装：index / auth / host / csv / cloud_resource / person / zero_trust / domain / stats / subnet
        ├── views/                 # 页面：Login / HostManagement / ResourceStatistics / IpStatistics / BusinessStatistics / PersonnelManagement / ZeroTrustLedger / DomainLedger
        ├── components/            # 组件：工具栏/表格/各类弹窗
        │
        ├── utils/
        │   └── index.js           # 通用工具函数
        │
        └── styles/
            └── global.css         # 全局样式
```

---

## 数据模型

数据库共 8 张表：

| 表 | 模型 | 说明 |
|----|------|------|
| `users` | `models/user.go` | 用户认证 |
| `persons` | `models/person.go` | 人员信息（姓名/联系方式/单位名称） |
| `hosts` | `models/host.go` | 主机技术信息 |
| `host_applications` | `models/host_application.go` | 主机申请信息（与 hosts 一对一） |
| `cloud_resources` | `models/cloud_resource.go` | 云资源总览（按区域） |
| `ip_subnets` | `models/ip_subnet.go` | IP 网段管理 |
| `zero_trusts` | `models/zero_trust.go` | 零信任接入申请台账 |
| `domains` | `models/domain.go` | 域名台账 |

### users 表

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 用户名 | `username` | VARCHAR(64) | NOT NULL, UNIQUE |
| 密码哈希 | `password_hash` | VARCHAR(256) | NOT NULL |
| 失败尝试次数 | `failed_attempts` | INTEGER | DEFAULT 0 |
| 锁定截止时间 | `locked_until` | TIMESTAMPTZ | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |
| 最后登录时间 | `last_login` | TIMESTAMPTZ | |

### persons 表（人员信息）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 姓名 | `name` | VARCHAR(64) | NOT NULL |
| 联系方式 | `contact` | VARCHAR(64) | |
| 单位名称 | `unit` | VARCHAR(128) | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### hosts 表（技术信息）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 区域 | `region` | VARCHAR(64) | NOT NULL |
| 实例ID | `instance_id` | VARCHAR(128) | |
| 主机名称 | `name` | VARCHAR(128) | NOT NULL |
| 内网IP | `private_ip` | VARCHAR(45) | NOT NULL, UNIQUE |
| 公网IP | `public_ip` | VARCHAR(45) | |
| 资产类型 | `asset_type` | VARCHAR(32) | 可选值：虚拟机、裸金属服务器 |
| 操作系统 | `os` | VARCHAR(64) | |
| CPU核数 | `cpu` | INTEGER | |
| CPU架构 | `cpu_arch` | VARCHAR(16) | 可选值：C86、X86、ARM |
| 内存(GB) | `memory` | INTEGER | |
| 磁盘(GB) | `disk` | INTEGER | |
| 系统盘(GB) | `system_disk` | INTEGER | |
| 数据盘(GB) | `data_disk` | INTEGER | |
| 环境类型 | `env_type` | VARCHAR(16) | 可选值：测试、生产 |
| 是否数据库服务器 | `is_db_server` | BOOLEAN | DEFAULT FALSE |
| 状态 | `status` | VARCHAR(32) | 可选值：运行中、已停止、已关机、待确认 |
| 开放端口 | `open_ports` | TEXT | 如 "80,443,22/tcp" |
| 标签 | `tags` | TEXT | |
| 人员ID | `person_id` | BIGINT | 可空，FK → persons(id) |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### host_applications 表（申请信息）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 主机ID | `host_id` | INTEGER | NOT NULL, UNIQUE, FK → hosts(id) ON DELETE CASCADE |
| 申请单位 | `apply_unit` | VARCHAR(128) | |
| 申请人 | `applicant` | VARCHAR(64) | |
| 申请人联系方式 | `applicant_contact` | VARCHAR(64) | |
| 所属项目 | `project` | VARCHAR(128) | |
| 申请理由 | `apply_reason` | TEXT | |
| 申请配置 | `apply_config` | TEXT | |
| 申请时间 | `apply_time` | VARCHAR(32) | |
| 对象存储大小 | `object_storage_size` | VARCHAR(32) | 如 "500GB"、"1TB" |
| 备注 | `remark` | TEXT | |

### cloud_resources 表（云资源总览）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 区域 | `region` | VARCHAR(64) | NOT NULL, UNIQUE |
| 物理CPU(核) | `physical_cpu` | INTEGER | DEFAULT 0 |
| vCPU(核) | `vcpu` | INTEGER | DEFAULT 0 |
| 内存(G) | `memory` | INTEGER | DEFAULT 0 |
| 存储(G) | `storage` | INTEGER | DEFAULT 0 |
| 裸金属(台) | `bare_metal` | INTEGER | DEFAULT 0 |
| GPU卡数 | `gpu_card_count` | INTEGER | DEFAULT 0 |
| 对象存储(G) | `object_storage` | INTEGER | DEFAULT 0 |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### ip_subnets 表（IP 网段管理）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 网段 | `cidr` | VARCHAR(32) | NOT NULL, UNIQUE |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### zero_trusts 表（零信任台账）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 申请单位 | `apply_unit` | VARCHAR(128) | NOT NULL |
| 账户名 | `account_name` | VARCHAR(64) | NOT NULL |
| 申请人联系方式 | `contact` | VARCHAR(64) | |
| 申请主机 | `host_id` | INTEGER | NOT NULL, FK → hosts(id)，应用层禁止删除被引用主机 |
| 申请端口 | `port` | INTEGER | NOT NULL, 1-65535 |
| 申请时间 | `apply_time` | TIMESTAMPTZ | NOT NULL |
| 备注 | `remark` | TEXT | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### domains 表（域名台账）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | SERIAL | PRIMARY KEY |
| 域名 | `domain` | VARCHAR(255) | NOT NULL, UNIQUE |
| 解析公网IP | `public_ip` | VARCHAR(45) | |
| 服务商 | `provider` | VARCHAR(128) | |
| 到期时间 | `expires_at` | TIMESTAMPTZ | 可空 |
| 备注 | `remark` | TEXT | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |
| 更新时间 | `updated_at` | TIMESTAMPTZ | DEFAULT NOW() |

### 字段约束说明

- `private_ip` 唯一约束，防止重复内网 IP
- `instance_id` 可选，用于关联云平台实例
- `is_db_server` 使用 PostgreSQL 原生 BOOLEAN
- `host_applications.host_id` 外键关联 `hosts.id`，级联删除
- `hosts.person_id` 外键关联 `persons(id)`，可空；人员被主机引用时禁止删除（应用层校验，返回 `40901`）
- `persons.name` 必填，姓名 + 联系方式 + 单位完全重复视为同一人员
- `ip_subnets.cidr` 仅允许 `/24` IPv4 网段，写入时自动规范化为网络地址（末位归 0）
- 系统**不预置默认网段**，由用户在「IP统计 → 管理网段」中维护
- `zero_trusts.host_id` 外键关联 `hosts.id`；主机被零信任台账引用时禁止删除（应用层校验，返回 `40901`）
- `domains.domain` 唯一约束，同域名不允许重复登记
- **GORM 列名陷阱**：`CIDR` 字段默认会被命名为 `c_id_r`，模型已显式指定 `gorm:"column:cidr"`

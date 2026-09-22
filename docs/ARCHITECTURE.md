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
- **后端**：分层架构，启动时 GORM AutoMigrate 自动建表 + 种子数据
- **数据库**：PostgreSQL，5 张业务表

---

## 目录结构

```
go/
├── README.md                      # 功能特性与使用说明
├── AGENTS.md                      # 项目总纲（统一入口）
├── docs/                          # 工程文档
│   ├── DEVELOPMENT.md             # 开发规范与准则
│   ├── ARCHITECTURE.md            # 架构与数据模型（本文件）
│   ├── API.md                     # API 约定
│   ├── BUSINESS_LOGIC.md          # 关键业务逻辑
│   ├── CODE_INDEX.md              # 代码索引
│   └── PROJECT_STATUS.md          # 开发情况与技术债
│
├── docker/                        # 容器化配置
│   ├── docker-compose.dev.yml     # 开发环境（DB + 单容器应用）
│   └── docker-compose.prod.yml    # 生产环境（DB + 单容器应用）
│
├── backend/                       # Go 后端 + Docker 构建
│   ├── main.go                    # 入口：启动 Gin、连接 DB、运行迁移、注册路由
│   ├── go.mod
│   ├── go.sum
│   ├── .env.example               # 环境变量模板
│   ├── Dockerfile                 # 多阶段构建（Node 前端 + Go 后端 → Alpine 运行）
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
│   │   ├── cloud_resource.go      # CloudResource 模型
│   │   └── ip_subnet.go           # IPSubnet 模型（IP 网段管理）
│   │
│   ├── controllers/
│   │   ├── response.go            # 统一响应辅助函数
│   │   ├── auth.go                # 登录、登出、改密码、用户信息
│   │   ├── host.go                # 主机 CRUD
│   │   ├── batch.go               # 批量添加/编辑
│   │   ├── csv.go                 # CSV 导入/导出/模板
│   │   ├── cloud_resource.go      # 云资源总览
│   │   ├── stats.go               # 统计（IP 使用、探测、业务统计）
│   │   ├── subnet.go              # IP 网段 CRUD
│   │   └── websocket.go           # WebSocket IP 探测通道
│   │
│   ├── middleware/
│   │   ├── jwt.go                 # JWT 认证中间件
│   │   └── cors.go                # CORS 跨域中间件
│   │
│   ├── services/
│   │   ├── auth_service.go        # 认证业务逻辑
│   │   ├── host_service.go        # 主机业务逻辑
│   │   ├── cloud_resource_service.go # 云资源总览业务逻辑
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
│   └── migrations/                # Goose SQL 迁移文件（历史保留，实际用 AutoMigrate）
│
└── frontend/                      # Vue 3 前端
    ├── index.html
    ├── package.json
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
        ├── api/                   # API 封装：index / auth / host / csv / cloud_resource / stats / subnet
        ├── views/                 # 页面：Login / HostManagement / ResourceStatistics / IpStatistics / BusinessStatistics
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

数据库共 5 张表：

| 表 | 模型 | 说明 |
|----|------|------|
| `users` | `models/user.go` | 用户认证 |
| `hosts` | `models/host.go` | 主机技术信息 |
| `host_applications` | `models/host_application.go` | 主机申请信息（与 hosts 一对一） |
| `cloud_resources` | `models/cloud_resource.go` | 云资源总览（按区域） |
| `ip_subnets` | `models/ip_subnet.go` | IP 网段管理 |

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

### 字段约束说明

- `private_ip` 唯一约束，防止重复内网 IP
- `instance_id` 可选，用于关联云平台实例
- `is_db_server` 使用 PostgreSQL 原生 BOOLEAN
- `host_applications.host_id` 外键关联 `hosts.id`，级联删除
- `ip_subnets.cidr` 仅允许 `/24` IPv4 网段，写入时自动规范化为网络地址（末位归 0）
- 系统**不预置默认网段**，由用户在「IP统计 → 管理网段」中维护
- **GORM 列名陷阱**：`CIDR` 字段默认会被命名为 `c_id_r`，模型已显式指定 `gorm:"column:cidr"`

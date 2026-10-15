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
- **数据库**：PostgreSQL，11 张表

---

## 目录结构

> 快照日期：2026-10-10（由 `scripts/check_docs.sh` 的 tree 检查守护：树中文件必须存在、已跟踪文件必须登记）

```
go/
├── README.md                      # 功能特性与使用说明
├── AGENTS.md                      # 项目总纲（统一入口）
├── .gitignore                     # Git 忽略规则（.env、dist/、node_modules、IDE/OS 文件）
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
│   │   ├── zero_trust_target.go   # ZeroTrustTarget 模型（接入目标子表，T-052；端口列 ports 表达式 T-056）
│   │   ├── port_mapping.go        # PortMapping 模型（端口映射台账）
│   │   ├── public_ip.go           # PublicIP 模型（公网IP资源台账）
│   │   └── audit_log.go           # AuditLog 模型（审计日志）
│   │
│   ├── controllers/
│   │   ├── response.go            # 统一响应辅助函数
│   │   ├── auth.go                # 登录、登出、改密码、用户信息
│   │   ├── host.go                # 主机 CRUD
│   │   ├── batch.go               # 批量添加/编辑
│   │   ├── csv.go                 # CSV 导入/导出/模板
│   │   ├── data_exchange.go       # 数据备份 xlsx 导出/导入（T-038）
│   │   ├── cloud_resource.go      # 云资源总览
│   │   ├── person.go              # 人员 CRUD
│   │   ├── zero_trust.go          # 零信任台账 CRUD
│   │   ├── port_mapping.go        # 端口映射台账 CRUD
│   │   ├── public_ip.go           # 公网IP资源台账 CRUD
│   │   ├── stats.go               # 统计（IP 使用、探测、业务统计）
│   │   ├── subnet.go              # IP 网段 CRUD
│   │   ├── audit_log.go           # 审计日志查询
│   │   ├── audit_context.go       # 审计上下文（operator + request_id）
│   │   ├── websocket.go           # WebSocket IP 探测通道
│   │   └── csv_test.go            # CSV 导入 16MB 超限/未超限穿透/缺字段测试（T-044）
│   │
│   ├── middleware/
│   │   ├── jwt.go                 # JWT 认证中间件 + CORS 白名单中间件 + ValidateToken（T-022/T-043）
│   │   └── jwt_test.go            # ValidateToken 单测（有效/过期/伪造/none 算法，T-022）
│   │
│   ├── services/
│   │   ├── auth_service.go        # 认证业务逻辑
│   │   ├── host_service.go        # 主机业务逻辑（CRUD + 筛选）
│   │   ├── host_batch_service.go  # 主机批量业务逻辑（结构化/文本批量 + 批量编辑）
│   │   ├── host_csv_service.go    # 主机 CSV 行映射与导入导出
│   │   ├── cloud_resource_service.go # 云资源总览业务逻辑
│   │   ├── person_service.go      # 人员业务逻辑
│   │   ├── zero_trust_service.go  # 零信任台账业务逻辑
│   │   ├── port_mapping_service.go # 端口映射台账业务逻辑（多端口校验 + ip_mapped 重算）
│   │   ├── public_ip_service.go   # 公网IP资源台账业务逻辑
│   │   ├── stats_service.go       # IP 使用统计 + 连通性探测
│   │   ├── business_stats.go      # 业务统计聚合（项目/公司/人员）
│   │   ├── subnet_service.go      # IP 网段业务逻辑（/24 校验）
│   │   ├── audit_service.go       # 审计日志（Record/List，操作人快照 + request_id 聚合）
│   │   ├── data_exchange.go       # 数据备份公共定义（8 sheet 规格 + 导入报告）
│   │   ├── export_service.go      # 数据备份导出（8 sheet xlsx）
│   │   ├── import_service.go      # 数据备份导入（自然键 upsert + 单事务回滚）
│   │   ├── testdb_test.go         # 测试基建：sqlite 内存库注入 database.DB（T-006）
│   │   ├── host_service_test.go   # 主机 CRUD/筛选表驱动单测（T-006）
│   │   ├── auth_service_test.go   # 登录/锁定表驱动单测（T-006）
│   │   ├── host_csv_service_test.go # 筛选导出行数一致性单测（T-008）
│   │   ├── stats_service_test.go  # 探测状态机/TCP/ICMP 单测（T-011）
│   │   ├── zero_trust_service_test.go # 接入目标子表 CRUD/按主机搜索/多端口表达式/批量端口组单测（T-052/T-056）
│   │   ├── port_mapping_service_test.go # 端口协议/范围归一化、展开配对、整组判重单测（T-055）
│   │   └── data_exchange_test.go  # 空库 xlsx 回灌（零信任子表/映射协议列往返）单测（T-052/T-055）
│   │
│   ├── routes/
│   │   ├── routes.go              # 路由协调：CORS + JWT 受保护组收口 + 分发各域
│   │   ├── auth.go                # 认证域（登录/登出公开，me/改密鉴权）
│   │   ├── host.go                # 主机域（CRUD/批量/CSV 导入导出）
│   │   ├── data_exchange.go       # 数据备份域（xlsx 整体导出/导入）
│   │   ├── cloud_resource.go      # 云资源域
│   │   ├── stats.go               # 统计域（IP/业务统计、探测）
│   │   ├── subnet.go              # IP 网段域
│   │   ├── person.go              # 人员域
│   │   ├── public_ip.go           # 公网IP域
│   │   ├── zero_trust.go          # 零信任台账域
│   │   ├── port_mapping.go        # 端口映射台账域
│   │   ├── audit.go               # 审计日志域（查询）
│   │   ├── websocket.go           # WS 探测通道（query token JWT 有效性校验，T-022）
│   │   └── websocket_test.go      # WS token 缺失/无效拒绝 + 有效建连测试（T-022）
│   │
│   ├── cmd/
│   │   └── resetpw/main.go        # 密码重置工具（随机/指定密码）
│   │
│   ├── utils/
│   │   ├── password.go            # 密码哈希工具（bcrypt，兼容旧 SHA-256 平滑升级，T-009）
│   │   ├── csv.go                 # CSV 解析/生成
│   │   ├── datetime.go            # 日期时间统一口径（解析/校验/格式化，T-040）
│   │   └── password_test.go       # 密码哈希单测（bcrypt 往返/旧格式兼容，T-006/T-009）
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
        ├── composables/           # 组合式函数（T-004/T-005）：useProbeWebSocket / useChartOption / usePagedTable / useProbeGrid
        ├── api/                   # API 封装：index / auth / host / csv / dataExchange / cloud_resource / person / zero_trust / port_mapping / public_ip / stats / subnet / audit
        ├── views/                 # 页面：Login / HostManagement / ResourceStatistics / IpStatistics / BusinessStatistics / PersonnelManagement / PublicIPManagement / ZeroTrustLedger / MappingLedger / DataBackup / AuditLog
        ├── components/            # 组件：顶部导航 AppNav（T-019）/工具栏/表格/各类弹窗（ip/：网段管理/探测面板/网段卡片）
        │
        ├── utils/
        │   ├── index.js           # 通用工具函数
        │   └── __tests__/         # Vitest 冒烟单测（T-006，npm run test）
        │
        └── styles/
            └── global.css         # 全局样式
```

> **迁移索引口径（T-039）**：唯一索引统一采用 GORM 期望的 `idx_` 前缀命名（`idx_<表>_<列>`），历史 `*_key` / `uni_` 重复约束由迁移收敛（`backend/migrations/20261009000001_align_unique_indexes_and_fks.sql`）。
> **迁移↔实库对齐（T-048）**：`port_mappings`/`public_ips` 主键 `integer`→`bigint`（含序列）、6 表 `created_at` 补 `DEFAULT now()`、`uni_port_mappings_domain`→`idx_port_mappings_domain`、`users.failed_attempts`/`host_applications.host_id` 对齐 `bigint`、`host_applications` 补 `created_at`/`updated_at`、FK 口径统一实库 `NO ACTION`（对齐迁移 `20261010000002_align_schema_drift.sql`，幂等可重跑）。**不做**：软删除、CHECK、普通索引补齐、`people_pkey` 改名。

---

## 数据模型

数据库共 11 张表：

| 表 | 模型 | 说明 |
|----|------|------|
| `users` | `models/user.go` | 用户认证 |
| `persons` | `models/person.go` | 人员信息（姓名/联系方式/单位名称） |
| `hosts` | `models/host.go` | 主机技术信息 |
| `host_applications` | `models/host_application.go` | 主机申请信息（与 hosts 一对一） |
| `cloud_resources` | `models/cloud_resource.go` | 云资源总览（按区域） |
| `ip_subnets` | `models/ip_subnet.go` | IP 网段管理 |
| `zero_trusts` | `models/zero_trust.go` | 零信任接入申请台账 |
| `zero_trust_targets` | `models/zero_trust_target.go` | 零信任接入目标子表（一对「主机:端口表达式」一行，T-052/T-056） |
| `port_mappings` | `models/port_mapping.go` | 端口映射台账（公网IP↔内网主机多端口；域名可选） |
| `public_ips` | `models/public_ip.go` | 公网IP资源台账（IP/运营商/出口位置/备注） |
| `audit_logs` | `models/audit_log.go` | 审计日志（操作人快照/动作/资源/diff/request_id；只增不删） |

> 库中另有 goose 运行时表 `goose_db_version`（记录迁移版本，由 goose 维护），不计入业务表。

### users 表

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 用户名 | `username` | VARCHAR(64) | NOT NULL, UNIQUE |
| 密码哈希 | `password_hash` | VARCHAR(256) | NOT NULL | bcrypt（登录兼容旧 `salt$hash` 自动升级，T-009） |
| 失败尝试次数 | `failed_attempts` | BIGINT | DEFAULT 0 |
| 锁定截止时间 | `locked_until` | TIMESTAMPTZ | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |
| 最后登录时间 | `last_login` | TIMESTAMPTZ | |

### persons 表（人员信息）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 姓名 | `name` | VARCHAR(64) | NOT NULL |
| 联系方式 | `contact` | VARCHAR(64) | |
| 单位名称 | `unit` | VARCHAR(128) | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### hosts 表（技术信息）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 区域 | `region` | VARCHAR(64) | NOT NULL |
| 实例ID | `instance_id` | VARCHAR(128) | |
| 主机名称 | `name` | VARCHAR(128) | NOT NULL |
| 内网IP | `private_ip` | VARCHAR(45) | NOT NULL, UNIQUE |
| 是否映射公网 | `ip_mapped` | BOOLEAN | NOT NULL, DEFAULT FALSE |
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
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 主机ID | `host_id` | BIGINT | NOT NULL, UNIQUE（`idx_host_applications_host_id`），FK → hosts(id)（**无 ON DELETE 动作**，不级联） |
| 申请单位 | `apply_unit` | VARCHAR(128) | |
| 申请人 | `applicant` | VARCHAR(64) | |
| 申请人联系方式 | `applicant_contact` | VARCHAR(64) | |
| 所属项目 | `project` | VARCHAR(128) | |
| 申请理由 | `apply_reason` | TEXT | |
| 申请配置 | `apply_config` | TEXT | |
| 申请时间 | `apply_time` | VARCHAR(32) | |
| 对象存储大小 | `object_storage_size` | VARCHAR(32) | 如 "500GB"、"1TB" |
| 备注 | `remark` | TEXT | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |
| 更新时间 | `updated_at` | TIMESTAMPTZ | DEFAULT NOW() |

### cloud_resources 表（云资源总览）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
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
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 网段 | `cidr` | VARCHAR(32) | NOT NULL, UNIQUE |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### zero_trusts 表（零信任台账）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 申请单位 | `apply_unit` | VARCHAR(128) | NOT NULL |
| 账户名 | `account_name` | VARCHAR(64) | NOT NULL |
| 申请人 | `applicant` | VARCHAR(64) | 选填（T-060，与账户名区分） |
| 申请人联系方式 | `contact` | VARCHAR(64) | |
| 公网IP入口 | `public_ip` | VARCHAR(45) | 选填，须在 `public_ips` 资源池；接入地区读时带出 |
| 系统名称 | `system_name` | VARCHAR(128) | 选填 |
| 申请时间 | `apply_time` | TIMESTAMPTZ | NOT NULL |
| 备注 | `remark` | TEXT | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

> **接入目标已拆子表（T-052）**：原 `targets`（`host_id:port` 逗号串）列删除，接入目标以 `zero_trust_targets` 配对行存储（1NF）；API `targets` 为派生字段（模型 `gorm:"-"`，读时由子表组装，对齐 `port_mappings` 的 `isp`/`exit_location` 派生做法）。派生串格式 T-056 起为 `host_id:ports`（对间以 `;` 分隔、对内端口表达式保留逗号），T-052 期间为逗号分隔单口。

### zero_trust_targets 表（零信任接入目标）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 零信任ID | `zero_trust_id` | BIGINT | NOT NULL，索引 `idx_zero_trust_targets_zero_trust_id`，FK → zero_trusts(id)，**ON DELETE RESTRICT** |
| 主机ID | `host_id` | BIGINT | NOT NULL，索引 `idx_zero_trust_targets_host_id`，FK → hosts(id)，**ON DELETE RESTRICT** |
| 端口 | `ports` | TEXT | NOT NULL，端口表达式（单口 `22`/范围 `1000-10005`/列表 `22,443,5432` 可混合，各端口 1-65535，T-056；T-052~055 期间为 INTEGER 单口 `port` 列） |

> 一对「主机:端口表达式」一行；迁移 `20261010000005_zero_trust_targets.sql` 建表并回填存量 `targets` 串（格式非法或主机不存在的行跳过），随后删除 `zero_trusts.targets` 列；迁移 `20261010000007_zero_trust_targets_ports.sql`（T-056）把 `port` INTEGER 列改名并转为 `ports` TEXT 表达式列（条件守卫幂等；DO 块须包裹 `-- +goose StatementBegin/End`，否则 goose 切分器按体内分号截断导致启动失败，T-057）。删除被引用主机/台账记录时，应用层经子表计数校验返回 `40901`，数据库层双 FK RESTRICT 兜底（本表按 T-052 规格用 RESTRICT，晚于 T-048 的既有 NO ACTION 口径）。

### port_mappings 表（端口映射台账）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 公网IP | `public_ip` | VARCHAR(45) | NOT NULL |
| 内网主机 | `host_id` | BIGINT | NOT NULL, FK → hosts(id) |
| 外网端口 | `external_ports` | TEXT | NOT NULL（默认 `''`），逗号分隔 token：单端口或 `A-B` 连续范围（1-65535、起始≤结束），规范化存储（T-055） |
| 外网协议 | `external_protocols` | TEXT | NOT NULL，与外网端口 **token** 等长逗号并行（`tcp`/`udp`，缺省 tcp，T-055） |
| 内网端口 | `internal_ports` | TEXT | NOT NULL（默认 `''`），与外网端口 token 并行存储；与外网端口**展开后数量**一致、按展开顺序一一对应（范围可与单口混搭） |
| 内网协议 | `internal_protocols` | TEXT | NOT NULL，与内网端口 **token** 等长逗号并行（`tcp`/`udp`，缺省 tcp，T-055） |
| 域名 | `domain` | VARCHAR(255) | 可选；非空时唯一（唯一索引 `idx_port_mappings_domain`） |
| 备注 | `remark` | TEXT | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |
| 更新时间 | `updated_at` | TIMESTAMPTZ | DEFAULT NOW() |

> **运营商/出口位置为派生字段（T-049）**：`isp`/`exit_location` 两列已删除（历史传递依赖，值与资源池漂移），API 响应字段保留、恒从 `public_ips` 资源池读时带出（模型 `gorm:"-"`，对齐 `zero_trusts` 的接入地区做法）；搜索经 `public_ips` 匹配。

### public_ips 表（公网IP资源台账）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 公网IP | `ip` | VARCHAR(45) | NOT NULL, UNIQUE |
| 运营商 | `isp` | VARCHAR(128) | |
| 出口位置 | `exit_location` | VARCHAR(128) | IP 所在地，选填 |
| 备注 | `remark` | TEXT | |
| 创建时间 | `created_at` | TIMESTAMPTZ | DEFAULT NOW() |

### audit_logs 表（审计日志）

| 字段 | 列名 | 类型 | 约束 |
|------|------|------|------|
| ID | `id` | BIGSERIAL | PRIMARY KEY |
| 操作人ID | `operator_id` | BIGINT | NOT NULL（快照，非外键） |
| 操作人名 | `operator_name` | VARCHAR(64) | NOT NULL DEFAULT ''（快照） |
| 动作 | `action` | VARCHAR(32) | NOT NULL（create/update/delete/batch_update/import_csv/import_xlsx） |
| 资源类型 | `resource_type` | VARCHAR(32) | NOT NULL（host/person/public_ip/zero_trust/port_mapping/cloud_resource/ip_subnet/data_exchange） |
| 资源ID | `resource_id` | BIGINT | 可空（批量/导入类条目） |
| 详情 | `detail` | JSONB | `{"before":{...},"after":{...}}` 前后值 diff |
| 请求ID | `request_id` | VARCHAR(64) | NOT NULL DEFAULT ''（同一 HTTP 请求内多条审计共用） |
| 创建时间 | `created_at` | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

索引：`operator_id` / `action` / `resource_type` / `resource_id` / `request_id` / `created_at DESC`（仅普通索引，**无 UNIQUE 约束**、**无 users 外键**，只增不删）。

### 字段约束说明

- `private_ip` 唯一约束，防止重复内网 IP
- `instance_id` 可选，用于关联云平台实例
- `is_db_server` 使用 PostgreSQL 原生 BOOLEAN
- `host_applications.host_id` 外键关联 `hosts.id`（约束 `fk_hosts_application`），**无 ON DELETE 动作、不级联**；删除主机由应用层在事务内先删申请信息再删主机（`services/host_service.go` 的 `Delete`）
- `hosts.person_id` 外键关联 `persons(id)`，可空；人员被主机引用时禁止删除（应用层校验，返回 `40901`）
- `persons.name` 必填，姓名 + 联系方式 + 单位完全重复视为同一人员
- `ip_subnets.cidr` 仅允许 `/24` IPv4 网段，写入时自动规范化为网络地址（末位归 0）
- 系统**不预置默认网段**，由用户在「IP统计 → 管理网段」中维护
- `zero_trusts` 接入目标经 `zero_trust_targets` 子表存储（一对「主机:端口表达式」一行，`ports` 为单口/范围/列表表达式，双 FK `ON DELETE RESTRICT`）；主机被零信任台账引用时禁止删除（应用层子表去重计数校验，返回 `40901`）；API `targets` 为读时组装的派生字段；`public_ip` 须在公网IP资源池，接入地区（`exit_location`）读时带出不落库
- `port_mappings.host_id` 外键关联 `hosts.id`；外网/内网端口**展开后**数量必须一致（`A-B` 范围按 B-A+1 计，范围可与单口混搭配对），协议列表分别与对应端口 **token** 列表等长（仅 `tcp`/`udp`，缺省补 `tcp`）；主机被映射引用时禁止删除
- `port_mappings` 整组唯一 `(host_id, public_ip, external_ports, external_protocols)`（唯一索引 `idx_port_mappings_host_ip_ports`：`T-051` 落前三列、`T-055` 纳入外网协议列并回填存量为 tcp，迁移 `20261010000006` 幂等；应用层 Create/Update 判重返回 `40901`，同端口不同协议不互斥）
- `hosts.ip_mapped` 由映射台账自动重算（有映射=true，无=false），主机表单不可手改
- `public_ips.ip` 唯一约束，公网 IP 资源池录入；删除/修改 IP 时若被 `port_mappings`/`zero_trusts` 引用（应用层值引用校验）返回 `40901`（`T-051`）
- **GORM 列名陷阱**：`CIDR` 字段默认会被命名为 `c_id_r`，模型已显式指定 `gorm:"column:cidr"`

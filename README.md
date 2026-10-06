# MCloud 资产管理系统

MCloud 是一个面向运维人员的资产信息管理系统，提供 Web 网页访问，用于表格化统计各类资产信息，并在资产管理基础上提供数据统计分析能力。

## 功能特性

### 主机资产管理

- **主机信息维护**：技术信息（区域、实例ID、主机名、IP、规格、OS、状态等）与申请信息（申请单位、申请人、项目等）的增删查改
- **批量操作**：批量添加、批量编辑（按字段选择更新）
- **CSV 导入导出**：支持 UTF-8 / GBK / GB2312 多编码自动识别，导出带 BOM 兼容 Excel，提供标准导入模板
- **多条件筛选**：跨 12 个字段的关键词模糊搜索 + 环境类型/资产类型/CPU架构/状态/区域/数据库服务器/申请人等筛选
- **分页展示**：规格列聚合展示（`4vCPU/8G/100G/50G`），标签化显示状态、环境、CPU架构、数据库标识

### 资源统计

- 按**区域**（region-a / region-b）常显云端总资源
- 汇总表格展示每个区域各资源的 **总量 / 已用 / 剩余(含百分比) / 运行中 / 已停止**
- 每个区域独立 2×2 饼图：**vCPU / 内存 / 存储 / 裸金属** 使用占比
- 已用量从 `hosts` 表实时聚合，区分运行状态：
  - 运行中（`运行中`）
  - 已停止（`已停止` + `已关机` 合并）
  - 待确认不计入
- 资源统计（vCPU、内存、存储）**均排除裸金属服务器**，避免重复计算

### IP 统计

- 按 **/24 网段**以位图形式（10 列 × 26 行坐标布局）展示 256 个 IP 的使用情况
- 三种颜色状态直观区分：

  | 颜色 | 含义 |
  |------|------|
  | 🟢 绿色 | IP 未使用 + 表中无记录 |
  | 🔴 红色 | IP 已使用 + 表中有非空记录 |
  | 🟡 黄色 | IP 已使用/待确认 + 表中无记录或记录为空 |

- **单点探测**：点击任意格子，后端发送 ICMP + TCP(22/3389) 探测，按状态机自动更新颜色
- **全量测试**：按网段一键并发探测全部 256 个 IP，带进度条与已探测高亮边框
- **实时返回**：探测结果通过 WebSocket 推送，支持并发探测与断线自动重连
- **网段管理**：网段在线增删改，限定 /24 掩码，自动规范化为网络地址

### 业务统计

- 按**项目 / 公司 / 人员**三个维度聚合资源用量
- 概览卡片：项目数、公司数、人员数、主机总数
- TOP10 横向柱状图：支持维度（项目/公司/人员）与指标（主机数/vCPU/内存/存储）切换
- 明细表格：各分组的资源用量、主机数、运行状态及维度间关联（如项目→公司、公司→项目列表）
- **点击下钻**：点击柱状图任意柱子，弹窗展示该分组下的主机列表（主机名 / IP / 对应项目）

### 配置管理

- **云资源录入**：右上角设置菜单入口，按区域 Tab 手动录入各区域的物理CPU/vCPU/内存/存储/裸金属/GPU/对象存储
- **IP 网段管理**：IP统计页入口，在线维护 /24 网段列表

## 技术栈

| 层次 | 技术 |
|------|------|
| 后端框架 | Go 1.22+ / Gin |
| ORM / 迁移 | GORM v2（启动时 `AutoMigrate` 兜底建表）+ SQL 迁移文件（归档/评审，不自动执行） |
| 数据库 | PostgreSQL 16 |
| 前端框架 | Vue 3 + Vite |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |
| 路由 | Vue Router 4 |
| HTTP 客户端 | Axios |
| 图表 | ECharts + vue-echarts |
| 实时通信 | gorilla/websocket |
| 认证 | JWT (golang-jwt/v5) |
| 密码哈希 | SHA-256 + 盐（`salt$hash`） |

## 快速开始

### 1. 启动服务（推荐 Docker）

```bash
cd docker
cp .env.example .env
docker-compose -f docker-compose.dev.yml up -d
```

启动后：
- 前端（开发）：`http://localhost:5173`（Vite，`/api` 自动代理到后端）
- 后端 API（开发）：`http://localhost:5677`
- 数据库：`localhost:5432`
- 生产环境（`docker-compose.prod.yml`）应用端口为 `5678`

### 2. 登录

```
用户名: admin
默认密码: Pass4MCloud
```

> 密码可在「设置 → 修改密码」中修改，也可用 `cmd/resetpw` 工具重置。

### 环境变量

参考 `docker/.env.example`，通过 `.env` 文件配置：

## 页面与导航

系统共 5 个功能页面：4 个顶部导航 Tab + 人员录入页（设置菜单进入）：

| 路由 | 页面 | 说明 |
|------|------|------|
| `/` | 主机管理 | 主机列表 + 筛选 + 分页（默认页） |
| `/statistics` | 资源统计 | 云端总资源 + 使用占比饼图 |
| `/ip-statistics` | IP 统计 | /24 网段位图 + 连通性探测 |
| `/business-statistics` | 业务统计 | 项目/公司/人员资源聚合 |
| `/personnel` | 人员管理 | 人员录入与维护，关联主机资产 |

右上角设置菜单：**云资源录入** / **人员录入** / **修改密码** / **退出登录**

## 生产构建

```bash
cd docker
cp .env.example .env
vi .env  # 修改 JWT_SECRET 为随机值
docker-compose -f docker-compose.prod.yml up -d --build
```

- 应用容器暴露 5678 端口（Nginx + Go 单容器）
- 后端 `SERVER_MODE=release`
- 建议在容器前加 SSL 终端（如 Traefik / Caddy）

## 密码重置

```bash
cd backend

# 生成 10 位高熵随机密码
go run ./cmd/resetpw/main.go

# 指定密码
go run ./cmd/resetpw/main.go Pass4MCloud
```

## 开发者文档

> **本项目以 AI 协作为主**：AI 会话只需读 [AGENTS.md](./AGENTS.md)（状态快照 + 执行协议已内联），即可知道当前进度与下一步工作；任务状态的唯一事实来源是 [docs/ROADMAP.md](./docs/ROADMAP.md)。

- [AGENTS.md](./AGENTS.md) —— 项目总纲（AI 协作入口：状态快照、会话协议、红线、验证命令）
- [docs/ROADMAP.md](./docs/ROADMAP.md) —— 任务队列（T-xxx、验收标准、状态流转）
- [docs/PROJECT_STATUS.md](./docs/PROJECT_STATUS.md) —— 完成度、变更记录、技术债
- [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) —— 架构与数据模型
- [docs/API.md](./docs/API.md) —— API 约定
- [docs/BUSINESS_LOGIC.md](./docs/BUSINESS_LOGIC.md) —— 关键业务逻辑
- [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) —— 开发规范与准则
- [docs/CODE_INDEX.md](./docs/CODE_INDEX.md) —— 代码索引

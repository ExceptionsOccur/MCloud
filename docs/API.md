# API 约定

> 本文档定义 MCloud 的接口规范与完整路由。文档入口见 [AGENTS.md](../AGENTS.md)。

## 目录

- [统一响应格式](#统一响应格式)
- [路由总览](#路由总览)
- [WebSocket 探测帧协议](#websocket-探测帧协议)
- [筛选参数说明](#筛选参数说明)
- [批量操作请求体](#批量操作请求体)

---

## 统一响应格式

所有 API 返回统一 JSON 格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code = 0` 表示成功
- `code != 0` 表示失败，`message` 包含错误描述

### 错误码规范

| 错误码 | 含义 | HTTP 状态 |
|--------|------|-----------|
| `40001` | 参数校验失败 | 400 |
| `40101` | 未登录 / Token 无效 | 401 |
| `40102` | 账户已锁定 | 401 |
| `40401` | 资源不存在 | 404 |
| `40901` | 数据冲突（如 IP 重复） | 409 |
| `50001` | 服务器内部错误 | 500 |

---

## 路由总览

### 认证（公开，无需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| POST | `/api/auth/login` | `auth.Login` | 登录，返回 JWT token |
| POST | `/api/auth/logout` | `auth.Logout` | 登出（客户端清 token） |

### 认证（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/auth/me` | `auth.Me` | 获取当前用户信息 |
| POST | `/api/auth/password` | `auth.ChangePassword` | 修改密码 |

### 主机管理（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/hosts` | `host.Filter` | 分页 + 搜索 + 筛选列表 |
| GET | `/api/hosts/:id` | `host.Get` | 单条详情 |
| POST | `/api/hosts` | `host.Create` | 新增主机 |
| PUT | `/api/hosts/:id` | `host.Update` | 编辑主机 |
| DELETE | `/api/hosts/:id` | `host.Delete` | 删除主机 |

### 批量操作（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| POST | `/api/batch/hosts` | `batch.BatchCreate` | 批量添加（JSON 数组） |
| POST | `/api/batch/hosts/text` | `batch.BatchCreateText` | 批量添加（纯文本，每行一条记录，逗号分隔） |
| PUT | `/api/batch/hosts` | `batch.BatchUpdate` | 批量编辑 |

### CSV 操作（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| POST | `/api/import` | `csv.Import` | 上传 CSV 批量导入 |
| GET | `/api/export` | `csv.Export` | 导出 CSV |
| GET | `/api/template` | `csv.Template` | 下载导入模板 |

### 云资源总览（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/cloud-resources` | `cloudResource.List` | 查询所有区域资源 |
| PUT | `/api/cloud-resources` | `cloudResource.Update` | 更新指定区域资源（按 region upsert） |

### 统计分析（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/stats/ip-usage` | `stats.IPUsage` | IP 使用情况（按网段返回已用/空记录 IP） |
| POST | `/api/stats/probe` | `stats.Probe` | 探测单个 IP 连通性（ICMP + TCP 22/3389） |
| GET | `/api/stats/business` | `stats.BusinessStats` | 业务统计（按项目/公司/人员聚合资源用量） |
| GET | `/api/ws/probe` | `HandleProbeWS` | WebSocket 探测通道（token 通过 query 传递） |

### IP 网段管理（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/ip-subnets` | `subnet.List` | 查询所有 IP 网段 |
| POST | `/api/ip-subnets` | `subnet.Create` | 新增网段（仅 /24，自动规范化网络地址） |
| PUT | `/api/ip-subnets/:id` | `subnet.Update` | 修改网段 |
| DELETE | `/api/ip-subnets/:id` | `subnet.Delete` | 删除网段 |

### 人员管理（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/persons` | `person.List` | 人员列表，`keyword` 模糊匹配姓名/联系方式/单位名称，返回含 `host_count`（关联主机数） |
| POST | `/api/persons` | `person.Create` | 新增人员（`name` 必填） |
| PUT | `/api/persons/:id` | `person.Update` | 修改人员 |
| DELETE | `/api/persons/:id` | `person.Delete` | 删除人员；被主机引用时返回 `40901` |

### 公网IP资源台账（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/public-ips` | `publicIP.List` | 公网IP列表，`keyword` 模糊匹配 IP/运营商/出口位置/备注 |
| POST | `/api/public-ips` | `publicIP.Create` | 新增公网IP（`ip` 必填、唯一、格式校验；`isp`/`exit_location`/`remark` 选填） |
| PUT | `/api/public-ips/:id` | `publicIP.Update` | 修改公网IP记录 |
| DELETE | `/api/public-ips/:id` | `publicIP.Delete` | 删除公网IP记录（允许直接删除） |

### 零信任台账（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/zero-trusts` | `zeroTrust.List` | 零信任申请列表，`keyword` 模糊匹配申请单位/账户名/联系方式/系统名称/备注/主机名/IP |
| POST | `/api/zero-trusts` | `zeroTrust.Create` | 新增申请（申请单位、账户名、申请主机、申请端口、申请时间必填；`system_name` 选填） |
| POST | `/api/zero-trusts/batch` | `zeroTrust.BatchCreateText` | 批量添加（文本粘贴，主机按内网IP定位；合法行全部插入） |
| PUT | `/api/zero-trusts/:id` | `zeroTrust.Update` | 修改申请记录 |
| DELETE | `/api/zero-trusts/:id` | `zeroTrust.Delete` | 删除申请记录 |

### 域名台账（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/domains` | `domain.List` | 域名台账列表，`keyword` 模糊匹配域名/公网IP/运营商/出口位置/备注/主机名/IP |
| POST | `/api/domains` | `domain.Create` | 新增域名记录（`domain` 必填，唯一；`host_id`/`host_port`/`isp`/`exit_location` 选填） |
| POST | `/api/domains/batch` | `domain.BatchCreateText` | 批量添加（文本粘贴；域名已存在跳过；内网IP可选关联主机） |
| PUT | `/api/domains/:id` | `domain.Update` | 修改域名记录 |
| DELETE | `/api/domains/:id` | `domain.Delete` | 删除域名记录 |

---

## WebSocket 探测帧协议

`GET /api/ws/probe?token=<JWT>`（`routes.go` 将连接升级为 WebSocket；浏览器 WS 请求无法携带 `Authorization` 头，故 token 走 query）。`token` 缺失返回 `40101`（当前仅校验非空，不校验 JWT 有效性——已知限制，`T-022` 已入队修复）。

双向均为 **Text frame + JSON**：

**客户端 → 服务端**（单帧探测请求，对应 `WSProbeRequest`）：

```json
{ "ip": "172.17.128.5", "color": "green" }
```

- `ip` 或 `color` 缺一 → 服务端静默丢弃（无响应帧）
- 非法 JSON → 静默丢弃

**服务端 → 客户端**（探测结果，对应 `WSProbeResponse`）：

```json
{ "ip": "172.17.128.5", "color": "yellow" }
```

- `color` 为探测后的**新颜色**（状态机见 [BUSINESS_LOGIC · IP 连通性探测](./BUSINESS_LOGIC.md#ip-连通性探测颜色状态机)）
- 探测执行失败（`stats.Probe` 返回 error）→ 不回帧
- 多帧并发：每条请求在独立 goroutine 中探测，回帧经互斥锁**串行写**避免帧交织；读循环遇错即退出并关闭连接，一条连接可连续发送多帧复用

---

## 筛选参数说明

`GET /api/hosts` 支持以下 query 参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码，默认 1 |
| `page_size` | int | 每页条数，默认 20，最大 100 |
| `keyword` | string | 模糊搜索（跨 12 个字段） |
| `env_type` | string | 环境类型筛选（测试/生产） |
| `asset_type` | string | 资产类型筛选（虚拟机/裸金属服务器） |
| `cpu_arch` | string | CPU架构筛选（C86/X86/ARM） |
| `is_db_server` | string | 数据库服务器筛选（0/1） |
| `status` | string | 状态筛选（运行中/已停止/已关机/待确认） |
| `region` | string | 区域筛选（region-a/region-b） |
| `applicant_empty` | string | 申请人筛选：`1`=未填写，`0`=已填写 |

### 搜索字段覆盖

`keyword` 模糊匹配以下字段（hosts 与 host_applications OR 联合搜索）：

- hosts 表：`region`, `instance_id`, `name`, `private_ip`, `os`, `status`, `tags`
- host_applications 表：`apply_unit`, `applicant`, `project`, `remark`

共 11 个字段，使用 `ILIKE '%keyword%'`（`public_ip` 已改为 `ip_mapped` 布尔，不参与关键词搜索）。

---

## 批量操作请求体

### 批量添加

```json
POST /api/batch/hosts
{
  "hosts": [
    {
      "region": "region-a",
      "name": "web-server-01",
      "private_ip": "192.168.1.10",
      "ip_mapped": true,
      "asset_type": "虚拟机",
      "env_type": "生产",
      "cpu": 4,
      "memory": 8,
      "system_disk": 50,
      "data_disk": 100,
      "applicant": "张三",
      "status": "运行中"
    }
  ]
}
```

跳过 `private_ip` 已存在的条目，返回成功/跳过/错误数量。

### 纯文本批量添加

```json
POST /api/batch/hosts/text
{
  "text": "区域,实例ID,主机名称,内网IP\nregion-a,ins-001,web-01,192.168.1.10"
}
```

- 每行一条记录，字段以逗号分隔，列顺序与 CSV 模板一致（`区域,实例ID,主机名称,内网IP,是否映射公网,资产类型,操作系统,CPU核数,CPU架构,内存(GB),系统盘(GB),数据盘(GB),环境类型,是否数据库服务器,状态,开放端口,标签,申请单位,申请人,申请人联系方式,所属项目,申请理由,申请配置,申请时间,对象存储大小,备注`）
- 至少需要前 3 列（区域、主机名称、内网IP），尾部列可省略（自动补空）
- 支持双引号包裹含逗号的字段；空行与 `#` 开头的注释行跳过；首行为表头时自动跳过
- 列数超过 26 或不足 3 列的行计入 `errors`
- 返回 `{ "success": n, "skipped": n, "errors": n, "line_errors": ["第5行: ..."] }`，`line_errors` 最多 10 条

### 批量编辑

```json
PUT /api/batch/hosts
{
  "ids": [1, 2, 3],
  "data": {
    "region": "region-a",
    "env_type": "生产",
    "project": "新项目"
  }
}
```

- 仅修改 `data` 中包含的字段，留空字段不修改
- `data` 中的字段自动按表分离：hosts 表字段更新 `hosts`，host_applications 表字段更新 `host_applications`
- `data.person_id` 为数字时批量关联人员，为 `null` 时批量解除关联；人员不存在返回 `40001`

### IP 探测请求

```json
POST /api/stats/probe
{
  "ip": "172.17.128.5",
  "color": "green"
}
```

返回新颜色：`{ "code": 0, "data": { "color": "yellow" } }`

### IP 网段请求

```json
POST /api/ip-subnets
{ "cidr": "172.17.128.0/24" }
```

- 仅接受 `/24` IPv4 网段，否则返回 `40001`
- `172.17.128.99/24` 会被规范化为 `172.17.128.0/24`

### 人员请求体

```json
POST /api/persons
{ "name": "张三", "contact": "13800000000", "unit": "某某研究院" }
```

- `name` 必填，`contact` / `unit` 可选
- 姓名 + 联系方式 + 单位完全重复时返回 `40001`

### 公网IP资源台账请求体

```json
POST /api/public-ips
{
  "ip": "203.0.113.10",
  "isp": "电信",
  "exit_location": "上海",
  "remark": "办公出口"
}
```

- `ip` 必填、全局唯一、须为合法 IPv4/IPv6；重复或格式错误返回 `40001`
- `isp` / `exit_location` / `remark` 可选
- 删除无引用校验，可直接删除

### 零信任台账请求体

```json
POST /api/zero-trusts
{
  "apply_unit": "某某研究院",
  "account_name": "zhangsan",
  "contact": "13800000000",
  "host_id": 1,
  "port": 22,
  "system_name": "统一门户",
  "apply_time": "2026-10-08T10:00:00",
  "remark": "临时开通"
}
```

- `apply_unit` / `account_name` / `host_id` / `port` 必填；`port` 范围 1-65535
- `system_name` 选填（VARCHAR 128），列表列与表单位于「申请端口」之后
- `host_id` 为 `hosts.id` 外键；主机不存在返回 `40001`
- `apply_time` 可选，缺省为服务端当前时间
- 列表返回嵌套 `host` 对象（名称/IP）；删除主机时若被台账引用返回 `40901`

### 台账批量添加请求体

```json
POST /api/zero-trusts/batch
{ "text": "申请单位,账户名,联系方式,内网IP,申请端口,系统名称,申请时间,备注\n某某研究院,zhangsan,138,192.168.1.10,22,统一门户" }
```

```json
POST /api/domains/batch
{ "text": "域名,解析公网IP,运营商,出口位置,内网IP,主机端口,备注\nwww.example.com,203.0.113.10,电信,上海,,," }
```

- 零信任列顺序：`申请单位,账户名,联系方式,内网IP,申请端口,系统名称,申请时间,备注`（至少前 4 列）；主机按内网IP定位，不存在则该行失败；合法行全部插入
- 域名列顺序：`域名,解析公网IP,运营商,出口位置,内网IP,主机端口,备注`（至少域名）；域名已存在跳过；内网IP为空不关联主机
- 返回 `{success, skipped, errors, line_errors[]}`，与主机批量接口同结构

### 域名台账请求体

```json
POST /api/domains
{
  "domain": "www.example.com",
  "public_ip": "1.2.3.4",
  "isp": "电信",
  "exit_location": "上海",
  "host_id": 3,
  "host_port": 8080,
  "remark": "业务主站"
}
```

- `domain` 必填且全局唯一，重复返回 `40001`
- `isp`（运营商）/ `exit_location`（出口位置）/ `host_id` / `host_port` / `public_ip` / `remark` 可选
- `host_id` 为 `hosts.id` 外键；填了不存在的主机返回 `40001`
- 列表返回嵌套 `host` 对象（名称/IP）；删除主机时若被域名台账引用返回 `40901`
- **无** `provider` / `expires_at` 字段（`T-030` 已重命名/删除）

### 主机的人员关联

`POST /api/hosts`、`PUT /api/hosts/:id` 支持 `person_id` 字段：

```json
{ "person_id": 1 }
```

- `person_id` 为 `null` 表示解除关联（仅 `PUT` 生效）
- 指定的人员不存在时返回 `40001`
- 列表与详情接口返回 `person` 对象（已关联时）
- 前端「申请人」为下拉 + 手输：选择已有人员直接带出联系方式/单位；手输新人员时前端先调用 `POST /api/persons` 写入人员库，再用返回的 `id` 作为 `person_id`

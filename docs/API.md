# API 约定

> 本文档定义 MCloud 的接口规范与完整路由。文档入口见 [AGENTS.md](../AGENTS.md)。

## 目录

- [统一响应格式](#统一响应格式)
- [日期时间格式](#日期时间格式)
- [路由总览](#路由总览)
- [WebSocket 探测帧协议](#websocket-探测帧协议)
- [筛选参数说明](#筛选参数说明)
- [批量操作请求体](#批量操作请求体)
- [数据备份 xlsx 导出/导入](#数据备份-xlsx-导出导入)

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
- `data` 为空时字段省略（`omitempty`）
- 导出/模板类接口（`GET /api/export`、`GET /api/template`、`GET /api/export/all`）返回二进制附件（CSV / xlsx），非 JSON

### 错误码规范

| 错误码 | 含义 | HTTP 状态 |
|--------|------|-----------|
| `40001` | 参数与业务校验失败（含部分唯一冲突，如公网IP/人员/网段重复、日期格式） | 400 |
| `40101` | 未登录 / Token 无效 / 账户已锁定 | 401 |
| `40401` | 资源不存在 | 404 |
| `40901` | 数据冲突与引用保护（如内网IP重复、被引用实体禁删改、映射整组重复） | 409 |
| `50001` | 服务器内部错误 | 500 |

---

## 日期时间格式

- **业务申请时间只记录到日期（T-065）**：主机 `host_applications.apply_time` 与零信任 `zero_trusts.apply_time` 均为纯日期 `YYYY-MM-DD`（零信任非空必填、缺省服务器当天；主机空值合法）；提交非空非法值一律拒绝（单条 `40001`，批量为行级错误，错误信息含 `YYYY-MM-DD`）
- **xlsx 导入宽容归一（T-065）**：导入的「申请时间」接受常见日期写法（`2026-10-09`、`2026/9/10`、`2026-9-1`、`2026.9.10`、`2026年9月10日`、`20260910`），解析成功后统一归一化存为 `YYYY-MM-DD`；两位年份（如 `26/9/10`）与带时刻写法不猜、报行级错误（整体回滚）
- **批量编辑**（`PUT /api/batch/hosts`）的 `apply_time`：字符串 `YYYY-MM-DD`（空串合法）、`null`（与省略同义，不更新该字段）；字符串但非 `YYYY-MM-DD` 或非字符串值返回 `40001`
- **导入**：xlsx 导入在报告 `errors[]` 中返回含 `YYYY-MM-DD` 的行级错误（存在行级错误时整体回滚）；CSV 导入只累计 `errors` 计数、不返回行级消息，坏行不入库
- **审计类时间保留完整时间戳**：各表 `created_at`/`updated_at`、审计日志、用户 `last_login`/`locked_until` 不受本口径影响；接口一律返回 RFC3339（带时区，如 `2026-10-16T02:25:31Z`，DB 存 UTC），由前端 `formatTime` 转浏览器本地时区展示
- **展示与导出**：审计/登录类时间展示统一 `YYYY-MM-DD HH:mm:ss`（前端本地时区）；业务申请时间展示与导出均为 `YYYY-MM-DD`（CSV 与 xlsx 备份）；备份文件名日期统一 `YYYY-MM-DD`

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
| GET | `/api/hosts` | `host.Filter` | 分页 + 搜索 + 筛选列表，按创建时间倒序（`id DESC`，最新创建在前，T-072）；响应含 `internal_count`/`external_count` 云内/云外汇总计数（T-078，现改由 `/api/hosts/stats` 提供，T-086） |
| GET | `/api/hosts/stats` | `host.Stats` | 全库主机计数 `{total, internal_count, external_count}`，**不受任何筛选条件影响**（T-086，供列表页脚） |
| GET | `/api/hosts/regions` | `host.ListRegions` | 区域候选值：hosts 中非空 `region` 去重（T-079 起表单/筛选的「区域」下拉只从这里取值 + 允许表单手输新区域） |
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
| POST | `/api/import` | `csv.Import` | 上传 CSV 批量导入（请求体 ≤16MB，超限返回 `40001`，T-044） |
| GET | `/api/export` | `csv.Export` | 导出 CSV（支持列表同名筛选参数，导出全部命中行、无分页） |
| GET | `/api/template` | `csv.Template` | 下载导入模板 |

### 数据备份（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/export/all` | `dataExchange.Export` | 导出 8 个业务 sheet 为单个**明文** xlsx（附件下载）；前端拿到字节后加密再下载（T-074） |
| POST | `/api/import/all` | `dataExchange.Import` | 上传 xlsx 整体导入（multipart 字段 `file`；加密 Excel 由前端先解密；upsert + 单事务回滚） |

### 审计日志（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/audit-logs` | `audit.List` | 分页查询审计日志；查询参数 `page`（默认 1）/`page_size`（默认 20，`>100` 回退 20）/`operator_name`/`action`/`resource_type`/`request_id`/`keyword` |
| GET | `/api/audit-logs/resource` | `audit.ResourceHistory` | 查询某资产的变更记录（资产详情「变更记录」时间线，T-066）；必填 `resource_type` + `resource_id`，返回按时间倒序的条目数组（`operator_name`/`action`/`detail`/`created_at` 等），缺参返回 `40001` |

- CUD/导入/批量操作自动写入 `audit_logs`（操作人快照 + 动作 + 资源 + 前后值 diff + 同请求共用 `request_id`）
- `action` 枚举：`create`/`update`/`delete`/`batch_update`/`import_csv`/`import_xlsx`；`resource_type` 枚举：`host`/`person`/`public_ip`/`zero_trust`/`port_mapping`/`cloud_resource`/`ip_subnet`/`data_exchange`/`ai_setting`
- `detail` 不含 password_hash 等敏感明文；只增不删；无 UNIQUE 约束、不设 users 外键
- 不覆盖：登录/改密事件、级联写（`refreshHostIPMapped`/登录计数）、WS 探测

### AI 接入配置（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/ai/settings` | `ai.GetSettings` | 读取 AI 接入配置，返回 `{api_url, api_key_masked, api_key_set, model}`；`api_key` 仅回显末四位（如 `****1234`），绝不返回明文（T-067） |
| PUT | `/api/ai/settings` | `ai.SaveSettings` | 保存配置（`api_url`/`model` 必填）；`api_key` 传空表示**保留原值**（前端脱敏回显无法回传明文），非空则加密后更新。变更写审计（`resource_type=ai_setting`），审计不落 API Key 明文 |
| POST | `/api/ai/extract` | `ai.Extract` | 上传申请表单（multipart 字段 `file`，`.docx`/`.xlsx`，≤20MB），调用大模型按文档内容识别类型（`host_apply` 主机申请 / `zero_trust_apply` 零信任申请）并提取字段；返回 `{doc_type, doc_type_name, fields, warnings, text_preview, file_token}`，**不写业务库**（T-068）。原文件暂存至后端 `incoming/` 目录并返回 `file_token` |
| POST | `/api/ai/archive` | `ai.Archive` | 归档已成功入库的源文件（body `{file_token}`）：将文件从 `incoming/` 移入 `archived/<YYYY-MM-DD>/`，返回 `{archived_path}`；token 非法/不存在/已归档返回 `40001`（T-068） |

- 配置为**单条全局记录**（`ai_settings` 表主键恒为 1），全站共用
- `api_key` 以 **AES-GCM 密文**落库，密钥由环境变量 `AI_SECRET_KEY` 经 SHA-256 派生；未配置该变量时降级明文存储并记录告警（T-067）
- AI 上传文件根目录由环境变量 `AI_UPLOAD_DIR` 配置（默认 `./data/ai_uploads`）；`extract` 落盘 `<AI_UPLOAD_DIR>/incoming/`，`archive` 移至 `<AI_UPLOAD_DIR>/archived/<YYYY-MM-DD>/`（T-068）
- `extract` 错误码：未配置 AI 接口 / 不支持的文档类型 / 解析失败 / 大模型调用失败均返回 `40001`

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
| GET | `/api/ws/probe` | `HandleProbeWS` | WebSocket 探测通道（token 通过 query 传递）† |

† 该路由未挂 JWT 中间件组，升级前在 `routes/websocket.go` 校验 query token 的 JWT 有效性（与 `Authorization` 同源，`T-022`）：`token` 缺失或无效/过期/伪造均返回 `40101`，详见[WebSocket 探测帧协议](#websocket-探测帧协议)。

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
| PUT | `/api/public-ips/:id` | `publicIP.Update` | 修改公网IP记录；**修改 IP 值**且该 IP 已被映射台账引用时返回 `40901` |
| DELETE | `/api/public-ips/:id` | `publicIP.Delete` | 删除公网IP记录；被 `port_mappings` 引用时返回 `40901`（`T-051`；零信任 `T-061` 起不再关联公网IP） |

### 零信任台账（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/zero-trusts` | `zeroTrust.List` | 零信任申请列表，`keyword` 模糊匹配申请单位/账户名/账号姓名/申请人/账户联系方式/系统名称/备注/接入地区/主机名/IP |
| GET | `/api/zero-trusts/exit-locations` | `zeroTrust.ListExitLocations` | 接入地区候选值：取 `public_ips.exit_location` 去重非空集（`T-061`，供前端下拉筛选） |
| POST | `/api/zero-trusts` | `zeroTrust.Create` | 新增申请（申请单位、账户名、`targets` 配对接入目标必填，`targets[]` 为 `{host_id, ports, protocol}`、`ports` 支持单口/范围/列表表达式且兼容旧 `port` 数字入参，`protocol` 缺省 tcp，T-056/T-065；申请时间、`account_name_zh`（账号姓名，T-063）、`system_name`、`exit_location`（多值，逗号分隔，T-065）选填，时间格式见[日期时间格式](#日期时间格式)）；**账户名与已有记录一致时合并进该记录**（targets 与接入地区并集去重、其余空字段补录非空保留，合并目标为 id 最小一条，T-058/T-065） |
| POST | `/api/zero-trusts/batch` | `zeroTrust.BatchCreateText` | 批量添加（文本粘贴，内网IP与申请端口两列等长位置配对；申请端口列含 `;` 按 `;` 分组表达多IP各自多端口，如 `22,443;8080,9000`，无 `;` 且仅 1 个端口组时广播到全部 IP（多IP相同端口，T-059），单IP整列逗号直接视为该主机多端口（免双层引号），旧平铺/双层引号语义保留兼容，T-058；同账户名**且同接入地区**的行合并进已有记录（T-077）；合法行全部插入） |
| PUT | `/api/zero-trusts/:id` | `zeroTrust.Update` | 修改申请记录 |
| DELETE | `/api/zero-trusts/:id` | `zeroTrust.Delete` | 删除申请记录 |

### 端口映射台账（需 JWT）

| 方法 | 路由 | Controller | 说明 |
|------|------|-----------|------|
| GET | `/api/port-mappings` | `portMapping.List` | 映射列表，`keyword` 模糊匹配公网IP/域名/端口/协议/备注/主机名/IP，以及资源池运营商/出口位置；`isp`/`exit_location` 为派生字段（`T-049` 起 `port_mappings` 不存列），响应值恒等于 `public_ips` 资源池 |
| POST | `/api/port-mappings` | `portMapping.Create` | 新增映射（公网IP须在资源池、内网主机、外网/内网端口必填；端口 token 支持单口或 `A-B` 连续范围（1-65535 且起始≤结束），**展开后**外网/内网数量一致（范围可与单口混搭配对）；`external_protocols`/`internal_protocols` 选填，与对应端口 **token** 等长、仅 `tcp`/`udp`，缺省补 `tcp`，T-055；域名可选；请求体传入的 `isp`/`exit_location` 被忽略）；整组 `(host_id, public_ip, external_ports, external_protocols)` 重复返回 `40901`（`T-051`/`T-055`） |
| POST | `/api/port-mappings/batch` | `portMapping.BatchCreateText` | 批量添加（文本粘贴；列：公网IP,内网IP,外网端口,内网端口,域名,备注[,外网协议,内网协议]，协议两列选填缺省 tcp，T-055）；重复整组计入 skipped |
| PUT | `/api/port-mappings/:id` | `portMapping.Update` | 修改映射记录；整组冲突返回 `40901` |
| DELETE | `/api/port-mappings/:id` | `portMapping.Delete` | 删除映射记录；删除后重算主机 `ip_mapped` |

---

## WebSocket 探测帧协议

`GET /api/ws/probe?token=<JWT>`（`routes/websocket.go` 将连接升级为 WebSocket；浏览器 WS 请求无法携带 `Authorization` 头，故 token 走 query）。升级前校验 token：缺失返回 `40101`（未提供认证信息）；无效/过期/伪造同样返回 `40101`（`middleware.ValidateToken`，与 HTTP `Authorization` 共用同一校验，`T-022`），校验通过才升级连接。

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

`GET /api/hosts` 与 `GET /api/export`（`T-008` 起同参筛选导出，导出忽略 `page`/`page_size`、返回全部命中行）支持以下 query 参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码，默认 1 |
| `page_size` | int | 每页条数，默认 20；>100 时回退默认 20 |
| `keyword` | string | 模糊搜索（跨 11 个字段） |
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
      "object_storage": 500,
      "system_disk": 50,
      "data_disk": 100,
      "is_external": false,
      "applicant": "张三",
      "status": "运行中"
    }
  ]
}
```

跳过 `private_ip` 已存在的条目，返回 `{ "success": n, "skipped": n, "errors": n, "line_errors": ["第1条: ..."] }`，`line_errors` 为行级错误消息数组，最多 10 条

### 纯文本批量添加

```json
POST /api/batch/hosts/text
{
  "text": "区域,实例ID,主机名称,内网IP\nregion-a,ins-001,web-01,192.168.1.10"
}
```

- 每行一条记录，字段以逗号分隔，列顺序与 CSV 模板一致（`区域,实例ID,主机名称,内网IP,是否映射公网,资产类型,操作系统,CPU核数,CPU架构,内存(GB),系统盘(GB),数据盘(GB),环境类型,是否数据库服务器,状态,标签,申请单位,申请人,申请人联系方式,所属项目,申请理由,申请配置,申请时间,备注,对象存储(GB),是否云外主机`；`是否云外主机` 为尾列（T-078，取值 是/否，缺省否），`对象存储(GB)`（T-072 前名为 `磁盘(GB)`）与旧文件缺列时按 0/云内处理；T-080 起删除申请信息「对象存储大小」列，总列数 27）
- 至少需要前 3 列（区域、实例ID、主机名称），尾部列可省略（自动补空）
- 支持双引号包裹含逗号的字段；空行与 `#` 开头的注释行跳过；首行为表头时自动跳过
- 列数超过 `len(CSVHeaders)`（当前 27）或不足 3 列的行计入 `errors`；前 3 列须为区域/实例ID/主机名称
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
- 删除被 `port_mappings` 引用的公网IP返回 `40901`（T-051），无引用可直接删除

### 零信任台账请求体

```json
POST /api/zero-trusts
{
  "apply_unit": "某某研究院",
  "account_name": "zhangsan",
  "account_name_zh": "张三丰",
  "applicant": "张三",
  "contact": "13800000000",
  "targets": [
    { "host_id": 1, "ports": "22,1000-10005", "protocol": "tcp" },
    { "host_id": 2, "ports": "53", "protocol": "udp" }
  ],
  "exit_location": "南宁",
  "system_name": "统一门户",
  "apply_time": "2026-10-08",
  "remark": "临时开通"
}
```

- `apply_unit` / `account_name` / `targets` 必填；`targets` 为「主机:端口表达式」**配对**数组（一次申请聚合多组，可表达 A 多端口、B 单端口等任意组合，存储于 `zero_trust_targets` 子表一对一行；API `targets` 为读时组装的派生字段，T-052/T-056）。`targets[].ports` 为端口表达式（T-056）：单端口 `22`、连续范围 `1000-10005`、不连续列表 `22,443,5432` 可混合，各端口 1-65535、范围起始≤结束；旧入参 `{host_id, port: 数字}` 仍兼容（ports 为空时回退），响应仅返回 `ports`；「主机+表达式+协议」重复配对自动去重
- `targets[].protocol` 选填（`tcp`/`udp`，缺省 `tcp`，T-065）：**每目标行单协议**，作用于该行整个 `ports` 表达式；非法值返回 `40001`
- `account_name_zh` 选填（VARCHAR 64，T-063），为「账号姓名」，是 `account_name`（账户名）的补充说明
- `contact` 选填（VARCHAR 64），为**账户联系方式**（T-064，非申请人联系方式）
- `applicant` 选填（VARCHAR 64），与 `account_name`（账户名）区分，为独立的申请人姓名
- **账户名 + 接入地区合并（T-058/T-077）**：`POST` 新增时若库中已有**同账户名且同接入地区**的记录（账户名 Trim 后精确匹配；地区按规整去重后**排序**整串比较，`南宁,青年` 与 `青年,南宁` 等价、空地区自成一组；存量同键多条时取 id 最小一条），则合并进该记录而非新建——接入目标取**并集去重**，`account_name_zh`/`applicant`/`contact`/`system_name`/`remark` 已有非空保留、新值仅补空位，`apply_time` 保留已有；**不同接入地区各自独立成记录，不合并**；无实际变化时仍返回原 id 且不写库；`PUT` 编辑不触发合并
- `targets[].host_id` 对应 `hosts.id`，任一主机不存在：`POST /api/zero-trusts` 返回 `40001`，`PUT /api/zero-trusts/:id` 返回 `40401`；非法端口表达式返回 `40001`
- `exit_location` 选填（TEXT，T-061/T-065），为**独立存储**的**多值**接入地区（逗号分隔，如 `南宁,青年`；逗号/顿号/中文逗号混用会被规整去重）；不关联公网IP、不强制校验取值，前端下拉候选值经 `GET /api/zero-trusts/exit-locations` 取 `public_ips.exit_location` 去重获得
- `system_name` 选填（VARCHAR 128），列表列与表单位于「申请端口」之后
- `apply_time` 选填（T-065 只记录到日期）：字符串 `YYYY-MM-DD`，缺省为服务器当天；非空但非法返回 `40001`，见[日期时间格式](#日期时间格式)
- 列表 `targets` 返回配对数组（`host_id`/`ports`/`protocol`/`host_name`/`private_ip`），附 `exit_location`；`GET` 详情与主机详情的派生 `targets` 字符串为 `host_id:ports:protocol`（对间以 `;` 分隔，对内表达式保留逗号，T-056/T-065）；删除主机时若被台账引用返回 `40901`

### 台账批量添加请求体

```json
POST /api/zero-trusts/batch
{ "text": "申请单位,账户名,账号姓名,申请人,账户联系方式,内网IP,申请端口,系统名称,申请时间,备注,接入地区\n某某研究院,zhangsan,张三丰,张三,138,192.168.1.10,22,统一门户,2026-10-08,,南宁" }
```

```json
POST /api/port-mappings/batch
{ "text": "公网IP,内网IP,外网端口,内网端口,域名,备注,外网协议,内网协议\n203.0.113.10,192.168.1.10,80,8080,www.example.com,业务,tcp,tcp" }
```

- 零信任列顺序：`申请单位,账户名,账号姓名,申请人,账户联系方式,内网IP,申请端口,系统名称,申请时间,备注,接入地区`（至少前 6 列，账号姓名第 3 列、申请人第 4 列、账户联系方式第 5 列、接入地区第 11 列均选填，T-060/T-061/T-063/T-064）；内网IP与申请端口两列**数量必须一致、按位置配对**，申请端口列每组为一个主机的端口表达式（T-056 支持范围/列表；T-058 起列值含 `;` 时按 `;` 分组表达多IP各自不同端口，如 `22,443;8080,1000-10005`——含逗号列值整体加一层 CSV 引号即可，旧平铺与双层引号写法保留兼容；单IP+整列逗号端口视为该主机多端口，免双层引号；T-059：无 `;` 且仅 1 个端口组时该组广播到全部 IP，多IP相同端口可只写 `22`，相同的多端口列表用重复分组 `22,443;22,443`）；**T-065 每个端口组可带 `/协议` 后缀**（`tcp`/`udp`，缺省 tcp，作用于该组整行，如 `53/udp`、`22,443/udp`；单主机多端口不支持 tcp/udp 混用），主机按内网IP定位，不存在则该行失败；接入地区第 11 列选填，**支持多值**（逗号/顿号分隔，如 `南宁,青年`；含逗号需整体加 CSV 引号，取自公网IP资源池去重，不做强制校验）；申请时间第 9 列为日期 `YYYY-MM-DD`（T-065，选填缺省当天，非法行失败）；账户名与已有记录一致的行合并进该记录（T-058）；合法行全部插入
- 映射列顺序：`公网IP,内网IP,外网端口,内网端口,域名,备注,外网协议,内网协议`（至少前 4 列，协议两列选填缺省 tcp，T-055）；公网IP须在资源池；端口支持单口与 `A-B` 连续范围 token，**展开后**数量不一致该行失败
- 映射的运营商/出口位置由公网IP资源池带出，不在批量行填写；零信任接入地区为独立存储字段（T-061/T-065，可多值，前端下拉候选取自资源池去重）
- 返回 `{success, skipped, errors, line_errors[]}`，与主机批量接口同结构

### 主机的人员关联

`POST /api/hosts`、`PUT /api/hosts/:id` 支持 `person_id` 字段：

```json
{ "person_id": 1 }
```

- `person_id` 为 `null` 表示解除关联（仅 `PUT` 生效）
- 指定的人员不存在时返回 `40001`
- 列表与详情接口返回 `person` 对象（已关联时）
- 前端「申请人」为下拉 + 手输：选择已有人员直接带出联系方式/单位；手输新人员时前端先调用 `POST /api/persons` 写入人员库，再用返回的 `id` 作为 `person_id`

## 数据备份 xlsx 导出/导入

`GET /api/export/all` 返回单文件 8 个工作表（T-063 起为中文名，顺序即导入依赖顺序）的 xlsx 附件（`Content-Disposition: attachment`，文件名 `mcloud_backup_YYYY-MM-DD.xlsx`）；前端拿到字节后用密码加密为 Excel 标准加密 `.xlsx` 再下载（T-074）：

`人员` → `公网IP` → `云资源` → `网段` → `主机` → `主机申请信息` → `零信任` → `端口映射`

- 工作表按中文名匹配：导入要求工作表名与上表一致（**不兼容旧英文名 persons/hosts/… 备份**）
- 不含 `users`，不含自增 id/时间戳列；关联一律用自然键表达：`主机` 用「人员姓名/人员联系方式」列、`主机申请信息` 与 `端口映射` 用宿主「内网IP」、`零信任` 按「账户名 + 内网IP」一行导出，端口分列 `内网IP`/`tcp端口`/`udp端口`（T-073：端口列支持掩码 `*`、逗号列表与 `A-B` 范围并可混用；`*` 表示全端口 1-65535）
- **`零信任` 按「账户名 + 接入地区 + 内网IP 一行」导出（T-073/T-077）**：同一账户名、同一接入地区下同一主机的 tcp/udp 端口分别并集到 `tcp端口`/`udp端口` 两列（同协议多目标合并、区间规整）；不同接入地区或不同内网IP 各占一行、其余字段取首条非空；无接入目标时「内网IP/端口」列为空。列头：`申请单位,账户名,账号姓名,申请人,账户联系方式,接入地区,内网IP,tcp端口,udp端口,系统名称,申请时间,备注`（**T-073 起删除旧「接入目标」列，不兼容旧备份**）。导入按「账户名 + 接入地区」合并、接入目标并集去重；`内网IP` 列沿用 `;` 分组广播语义，不做 CIDR 展开；两端口列均空视为行错误
- `主机` sheet 含「是否云外主机」列（T-078，是/否，旧备份无该列按云内）；「对象存储」列为 `hosts.object_storage`（T-072，原「磁盘」列，旧备份列头「磁盘」导入时兼容读取）；主机列表「对象存储」列同样展示 `hosts.object_storage`（T-080 起申请信息不再有该字段）
- `主机.ip_mapped` 为派生列不导出（导入后由后端重算）；`端口映射` 无运营商/出口位置列（`T-049` 起派生字段，库中不存），从资源池带出；`端口映射` 含「外网协议/内网协议」列（T-055，旧备份文件缺列时导入默认 tcp）；`零信任` 的「接入地区」列为独立存储字段（T-061，替代原「公网IP」列；T-065 支持多值，逗号分隔）
- 导入按各表复合自然键 upsert：`人员`=姓名+联系方式、`公网IP`=公网IP、`云资源`=区域、`网段`=网段CIDR、`主机`=内网IP、`主机申请信息`=宿主内网IP、`零信任`=账户名+接入地区（T-058/T-077 合并语义：同键才合并，接入目标并集、空字段补录、非空保留；同文件多行同键同样合并，不同地区各自独立）、`端口映射`=公网IP+宿主内网IP+外网端口+外网协议
- 语义：**仅新增与更新，不删除**已有数据；sheet 与列头按名匹配（列顺序可变），全空行跳过
- `POST /api/import/all`：`multipart/form-data`，字段 `file`，仅 `.xlsx`，≤ 20MB；**加密 Excel 由前端凭密码解密后再上传**（T-074，后端只接收明文 xlsx；上传未解密的加密文件返回 `40001`「文件不是有效的 .xlsx（若为加密 Excel，请在导入时填写密码）」）

加密方式（T-074）：前端 `officecrypto-tool` 生成 ECMA-376 标准 Office 加密 xlsx，用户可用 Excel/WPS 输入密码打开、另存修改后重新导入；T-072 的自定义 `.mcbak` 加密包已移除。

导入响应 `data`（导入报告）：

```json
{
  "committed": true,
  "sheets": [{ "sheet": "人员", "rows": 29, "created": 0, "updated": 0, "skipped": 29 }],
  "errors": []
}
```

- `sheets[]` 各表统计：`rows` 为数据行数，`created + updated + skipped = rows`
- `committed=false`：存在行级错误，**单事务已整体回滚、库数据不变**；`errors[]` 逐行给出 `sheet`（工作表名）、`row`（Excel 行号，表头为第 1 行）、`message`；此时 `sheets[]` 仅为事务内已执行但未落库的统计
- 错误码：非 `.xlsx` / 缺工作表 / 缺列 / 表头重复 / 超限 → `40001`；导入执行失败 → `50001`；未认证 → `40101`

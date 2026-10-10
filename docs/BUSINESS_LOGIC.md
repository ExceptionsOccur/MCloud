# 关键业务逻辑

> 本文档描述 MCloud 的核心业务实现逻辑。文档入口见 [AGENTS.md](../AGENTS.md)，代码位置见 [CODE_INDEX.md](./CODE_INDEX.md)。

## 目录

- [认证流程](#认证流程)
- [JWT 中间件](#jwt-中间件)
- [审计日志](#审计日志)
- [主机搜索/筛选](#主机搜索筛选)
- [CSV 导入导出](#csv-导入导出)
- [数据备份 xlsx 导出/导入](#数据备份-xlsx-导出导入)
- [IP 连通性探测（颜色状态机）](#ip-连通性探测颜色状态机)
- [资源统计（区分运行状态）](#资源统计区分运行状态)
- [业务统计聚合](#业务统计聚合)
- [云资源总览录入](#云资源总览录入)
- [人员管理](#人员管理)
- [前端表格与表单](#前端表格与表单)

---

## 认证流程

1. 前端提交 `{ username, password }` 到 `POST /api/auth/login`
2. 后端验证用户名存在 → 检查是否锁定 → 验证密码
3. 密码验证成功：重置 `failed_attempts`，更新 `last_login`，签发 JWT
4. 密码验证失败：`failed_attempts++`，达到 5 次锁定 15 分钟
5. 返回 `{ code: 0, data: { access_token, expires_in, user: { id, username } } }`

密码存储格式为 `salt$hash`，其中 `hash = SHA-256(salt + password)`。

## JWT 中间件

```go
// Authorization: Bearer <token>
// 验证通过 → c.Set("user_id", userID)
// 验证失败 → controllers.Error(c, 40101, ...)（HTTP 映射 401 统一响应）+ c.Abort()
func JWTAuth() gin.HandlerFunc { ... }
```

WebSocket 通道无法携带 Header，token 通过 query 参数传递（`/api/ws/probe?token=xxx`）。前端通过 `encodeURIComponent` 编码。

## 审计日志

- **写入时机**：service 层 CUD 成功后调用 `AuditService.Record`；覆盖常规 CRUD、批量增改、CSV 导入、xlsx 数据备份导入
- **不覆盖**：登录/改密事件、级联写（`refreshHostIPMapped`/登录计数）、WS 探测
- **操作人**：controller 经 `auditContext(c)` 从 JWT `user_id` 取快照（`operator_id` + `operator_name`）
- **request_id**：同一 HTTP 请求内多条审计共用（批量/CSV 导入逐行共用一个 ID）
- **detail**：`{"before":{...},"after":{...}}` 前后值 diff；不含 password_hash 等敏感明文
- **只增不删**：无 UNIQUE 约束、不设 users 外键（操作人以快照落库）
- **查询**：`GET /api/audit-logs`（JWT、分页筛选）；写入失败不阻断业务（审计不改变 CUD 结果语义）


## 主机搜索/筛选

`Filter` 使用 GORM 动态构建条件：

```go
db := database.DB.Model(&models.Host{})

if req.Keyword != "" {
    like := "%" + req.Keyword + "%"
    subQuery := database.DB.Model(&models.HostApplication{}).
        Select("host_id").
        Where("(apply_unit ILIKE ? OR applicant ILIKE ? OR project ILIKE ? OR remark ILIKE ?)", ...)
    db = db.Where(`(region ILIKE ? OR name ILIKE ? OR ... OR hosts.id IN (?))`, ..., subQuery)
}

// ... 更多筛选条件

var total int64
db.Count(&total)

var hosts []models.Host
db.Preload("Application").Offset(offset).Limit(pageSize).Order("id ASC").Find(&hosts)
```

`page_size` 上限为 100（超出会被重置为默认值）。

## CSV 导入导出

### 多编码检测

依次尝试 UTF-8 → GBK → GB18030 解码：

```go
// utils/csv.go
func DetectAndDecode(data []byte) (string, error) {
    if utf8Valid(data) {
        return string(data), nil
    }
    if decoded, err := decodeWith(data, simplifiedchinese.GBK.NewDecoder()); err == nil {
        return decoded, nil
    }
    if decoded, err := decodeWith(data, simplifiedchinese.GB18030.NewDecoder()); err == nil {
        return decoded, nil
    }
    return string(data), fmt.Errorf("unable to detect encoding")
}
```

### 表头映射

```go
var CSVHeaders = []string{
    "区域", "实例ID", "主机名称", "内网IP", "是否映射公网",
    "资产类型", "操作系统", "CPU核数", "CPU架构", "内存(GB)",
    "系统盘(GB)", "数据盘(GB)", "环境类型", "是否数据库服务器",
    "状态", "开放端口", "标签", "申请单位", "申请人", "申请人联系方式",
    "所属项目", "申请理由", "申请配置", "申请时间", "对象存储大小", "备注", "磁盘(GB)",
}
```

导出 CSV 需带 BOM（`\ufeff`）以便 Excel 正确识别 UTF-8。

### 申请时间列校验

`申请时间` 列为纯日期，严格 `YYYY-MM-DD`（空值合法）。非法值该行导入失败：CSV 文件导入计入 `errors` 且不入库，纯文本批量返回行级 `line_errors`，错误信息含 `YYYY-MM-DD`（接口与校验规则见 [API.md · 日期时间格式](./API.md#日期时间格式)）。

## 数据备份 xlsx 导出/导入

与「CSV 导入导出」（仅主机单表）互补，设置菜单「数据备份」进入 `/data-backup`，提供 8 个业务 sheet 的整库备份。

### 导出

- `GET /api/export/all` 返回单个 `.xlsx`（excelize），8 个 sheet 按导入依赖序排列：`persons` → `public_ips` → `cloud_resources` → `ip_subnets` → `hosts` → `host_applications` → `zero_trusts` → `port_mappings`
- 列头为中文、按名匹配；不导出 `users`、自增 id 与时间戳；关联用自然键表达：`hosts` 人员列（姓名+联系方式，由 `person_id` 解析）、宿主一律用「内网IP」、零信任「接入目标」由存储值 `host_id:port` 解析为 `内网IP:端口`
- 派生列不导出：`hosts.ip_mapped`、`port_mappings` 的运营商/出口位置（由公网IP资源池带出）

### 导入（upsert）

- 结构校验先行：非 `.xlsx`、缺 sheet、缺列、表头重复直接返回结构错误（`40001`），不进入事务
- 逐行解析复用既有规范化：`normalizeCIDR`（/24）、`parsePortList`（单端口 1-65535 校验，不做等长校验；端口列表等长校验在 `applyPortMappings` 与 `normalizePortMapping`）、`normalizeTargetPairs`（配对去重）、`ValidateDateOnly`（`host_applications.申请时间` 严格 `YYYY-MM-DD`）；人员按姓名(+联系方式)解析 `person_id`、主机按内网IP解析 `host_id`、填写公网IP时必须在 `public_ips` 资源池（空值合法放行）
- 申请时间两套口径：`host_applications` 走严格 `YYYY-MM-DD`（`utils.ValidateDateOnly`）；`zero_trusts` 走宽松 `parseTimeCell`（依次尝试 `YYYY-MM-DD HH:MM:SS` / `YYYY-MM-DDTHH:MM:SS` / RFC3339 / `YYYY-MM-DD HH:MM` / `YYYY-MM-DD` / `YYYY/MM/DD HH:MM:SS` / `YYYY/MM/DD`，空值保持零值）
- 每个 sheet 按复合自然键定位已有记录：值一致 → `skipped`，不一致 → `updated`，不存在 → `created`；**仅新增与更新，不执行删除**
- 单事务：任一行错误累计到 `errors` 并回滚（`errImportRolledBack`），返回 `committed=false` 与逐行明细，库数据不变；全部通过后重算 `hosts.ip_mapped`
- 详见 [API.md · 数据备份 xlsx 导出/导入](./API.md#数据备份-xlsx-导出导入)

## IP 连通性探测（颜色状态机）

### 颜色语义

| 颜色 | 含义 |
|------|------|
| 🟢 绿色 | IP 未使用 + 表中无记录 |
| 🔴 红色 | IP 已使用 + 表中有非空记录（`cpu > 0`） |
| 🟡 黄色 | IP 已使用/待确认 + 表中无记录或记录为空（`cpu = 0`） |

### 探测方式

ICMP（`ping -c 1`，context 超时 100ms）+ TCP 22/3389（超时 100ms），任一有回包即视为 alive。

> 注意：Linux `ping -W` 单位为**秒**，故超时改用 `exec.CommandContext` + 100ms context 控制。

### 结果处理逻辑（后端 `Probe`）

探测**只计算颜色并返回，不写入 `hosts` 表**（不再新增空记录）。

**探测通（alive）**：

| DB 状态 | 新颜色 |
|---------|--------|
| 无记录 | 🟡（仅会话内标记，不新增记录，刷新后恢复 🟢） |
| 有记录 + 有数据 | 🔴 |
| 有记录 + 空数据 | 🟡 |

**探测不通（dead）**：

| 当前颜色 / DB 状态 | 新颜色 |
|--------------------|--------|
| 🟢 + 无记录 | 🟢 |
| 非🟢 + 无记录 | 🟡 |
| 有记录 + 有数据 | 🔴（**保持红色，兼容关机/停止主机**） |
| 有记录 + 空数据 | 🟡 |

### 前端交互

- **单击格子**：发送探测请求，格子显示脉冲动画，收到结果后按颜色更新并弹提示
- **全量测试**：并发 32 探测整段 256 个 IP，进度条 + 已测高亮边框
- **WebSocket**：结果实时推送，未连接时降级 HTTP，断线 3 秒自动重连
- 提示框颜色与格子状态一致：未使用绿 / 已用红 / 待确认黄

## 资源统计（区分运行状态）

`cloud_resources` 存各区域总资源，`hosts` 按区域聚合已用量：

- 运行中 = `status === '运行中'`
- 已停止 = `status === '已停止' || status === '已关机'`（合并）
- 待确认 → 忽略不计
- 资源统计（vCPU、内存、存储）**均排除裸金属服务器**（`asset_type !== '裸金属服务器'`）
- 未使用 = 总量 − 已用（`Math.max(0, ...)` 防负）
- 饼图分三段：运行中 / 已停止 / 未使用

## 业务统计聚合

`GET /api/stats/business` 一次性返回三个维度聚合（`by_project` / `by_company` / `by_applicant`）：

- 每个分组统计：主机数、vCPU、内存、存储、运行中、已停止
- 同时记录维度间关联（如项目→公司列表、公司→项目列表）
- 空值统一归入 `(未填写)` 分组；概览计数不含该分组
- 每个分组附带 `hosts` 数组（主机名/IP/项目），供柱状图点击下钻
- 前端柱状图 TOP10，支持维度（项目/公司/人员）与指标（主机数/vCPU/内存/存储）切换

## 云资源总览录入

- 后端 `PUT /api/cloud-resources` 按 `region` upsert（存在则更新，否则创建）
- 前端入口在右上角设置菜单 →「云资源录入」，按区域 Tab 分别录入
- 资源统计页监听 `saved` 事件，保存后自动刷新

## 人员管理

- 字段：`name` / `contact` / `unit`，`normalizePerson` 统一 `TrimSpace`，`name` 必填
- **重复判定**（`services/person_service.go` → `Create`）：`name + contact + unit` 三字段**全等**才判重，命中返回「该人员已存在」；`Update` 不判重
- **错误码**：`person.Create/Update` 业务错误（含重复）→ `40001`；`person.Delete` 被主机引用 → `40901`（`ErrPersonReferenced`，文案含「N 台主机正在使用」）；人员不存在 → `40401`
- **主机侧关联校验**：`host` 创建/更新/批量编辑传 `person_id` 时经 `validatePersonID` 校验人员存在，不存在 → `40001`；`person_id: null` 解除关联；host 其余冲突类错误 → `40901`
- **手输自动建人员**（前端 `HostFormDialog.vue` → `resolvePersonId()`，提交主机时执行）：
  1. 表单已选人员且姓名一致 → 直接用其 `person_id`
  2. 否则按姓名在已加载人员列表中命中 → 复用既有 id
  3. 均未命中 → 调 `POST /api/persons`，用表单的申请人/联系方式/申请单位**新建人员**，取新 id 关联；接口失败返回 `false`，主机提交随即中止（错误提示由 axios 拦截器统一弹出）
- 列表 `keyword` 对姓名/联系方式/单位 `ILIKE` 模糊搜索，每条附 `host_count`（关联主机数）

## 前端表格与表单

### 主机列表表格列

| 列 | 宽度 | 说明 |
|----|------|------|
| 选择 | 50px | 复选框 |
| # | 60px | 序号（基于当前页） |
| 区域 | 100px | |
| 主机名称 | min 140px | 含环境标签 + DB标签 |
| 内网IP | 140px | |
| 是否映射公网 | 120px | 布尔标签 是/否 |
| 资产类型 | min 120px | 含 CPU 架构标签 |
| 规格 | min 200px | `4vCPU/8G/100G/50G`，缺失显示 `-` |
| 状态 | 90px | 颜色徽章 |
| 项目 | min 120px | 申请信息表，tooltip |
| 申请人 | min 100px | 申请信息表，tooltip |
| 操作 | 140px | 固定右侧 |

### 表单约定

- CPU核数、内存、系统盘、数据盘：普通 `el-input`，数字校验正则 `/^\d*$/`，提交时 `Number()` 转换

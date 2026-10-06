# 关键业务逻辑

> 本文档描述 MCloud 的核心业务实现逻辑。文档入口见 [AGENTS.md](../AGENTS.md)，代码位置见 [CODE_INDEX.md](./CODE_INDEX.md)。

## 目录

- [认证流程](#认证流程)
- [JWT 中间件](#jwt-中间件)
- [主机搜索/筛选](#主机搜索筛选)
- [CSV 导入导出](#csv-导入导出)
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
5. 返回 `{ code: 0, data: { access_token, expires_in } }`

密码存储格式为 `salt$hash`，其中 `hash = SHA-256(salt + password)`。

## JWT 中间件

```go
// Authorization: Bearer <token>
// 验证通过 → c.Set("user_id", userID)
// 验证失败 → c.AbortWithStatusJSON(401, ...)
func JWTAuth() gin.HandlerFunc { ... }
```

WebSocket 通道无法携带 Header，token 通过 query 参数传递（`/api/ws/probe?token=xxx`）。前端通过 `encodeURIComponent` 编码。

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
func detectAndDecode(data []byte) (string, error) {
    encodings := []encoding.Encoding{
        unicode.UTF8, charmap.Windows1252,
        simplifiedchinese.GBK, simplifiedchinese.GB18030,
    }
    // ...
}
```

### 表头映射

```go
var CSVHeaders = []string{
    "区域", "实例ID", "主机名称", "内网IP", "公网IP",
    "资产类型", "操作系统", "CPU核数", "CPU架构", "内存(GB)",
    "系统盘(GB)", "数据盘(GB)", "环境类型", "是否数据库服务器",
    "状态", "开放端口", "标签", "申请单位", "申请人", "申请人联系方式",
    "所属项目", "申请理由", "申请配置", "申请时间", "对象存储大小", "备注",
}
```

导出 CSV 需带 BOM（`\ufeff`）以便 Excel 正确识别 UTF-8。

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
| 公网IP | 140px | |
| 资产类型 | min 120px | 含 CPU 架构标签 |
| 规格 | min 200px | `4vCPU/8G/100G/50G`，缺失显示 `-` |
| 状态 | 90px | 颜色徽章 |
| 项目 | min 120px | 申请信息表，tooltip |
| 申请人 | min 100px | 申请信息表，tooltip |
| 操作 | 140px | 固定右侧 |

### 表单约定

- CPU核数、内存、系统盘、数据盘：普通 `el-input`，数字校验正则 `/^\d*$/`，提交时 `Number()` 转换

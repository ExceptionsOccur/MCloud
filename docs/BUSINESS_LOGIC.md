# 关键业务逻辑

> 本文档描述 MCloud 的核心业务实现逻辑。文档入口见 [AGENTS.md](../AGENTS.md)，代码位置见 [CODE_INDEX.md](./CODE_INDEX.md)。

## 目录

- [认证流程](#认证流程)
- [JWT 中间件](#jwt-中间件)
- [审计日志](#审计日志)
- [AI 接入配置与密钥加密](#ai-接入配置与密钥加密)
- [AI 申请表单提取与文件归档](#ai-申请表单提取与文件归档)
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
3. 密码验证成功：重置 `failed_attempts`，更新 `last_login`，签发 JWT；存储为旧 SHA-256 格式时**平滑升级为 bcrypt** 后落库（T-009）
4. 密码验证失败：`failed_attempts++`，达到 5 次锁定 15 分钟
5. 返回 `{ code: 0, data: { access_token, expires_in, user: { id, username } } }`

密码存储格式为 **bcrypt**（`$2a$…`，`golang.org/x/crypto/bcrypt`，cost 10）；存量旧格式 `salt$hash`（`hash = SHA-256(salt + password)`）登录校验兼容，成功后自动升级为 bcrypt（T-009）。`password_hash VARCHAR(256)` 两种格式均放得下，无 schema 变更。

## JWT 中间件

```go
// Authorization: Bearer <token>
// 验证通过 → c.Set("user_id", userID)
// 验证失败 → controllers.Error(c, 40101, ...)（HTTP 映射 401 统一响应）+ c.Abort()
func JWTAuth() gin.HandlerFunc { ... }
```

WebSocket 通道无法携带 Header，token 通过 query 参数传递（`/api/ws/probe?token=xxx`）。前端通过 `encodeURIComponent` 编码；升级前经共享的 `middleware.ValidateToken` 校验（与 `Authorization` 同源），缺失/无效/过期/伪造返回 `40101` 拒绝升级（T-022）。

## 审计日志

> 审计用于记录**资产变更记录**：不再有独立审计日志页，改为在主机/零信任/映射详情弹窗的「变更记录」时间线按资源维度展示（T-066）。

- **写入时机**：service 层 CUD 成功后调用 `AuditService.Record`；覆盖常规 CRUD、批量增改、CSV 导入、xlsx 数据备份导入
- **不覆盖**：登录/改密事件、级联写（`refreshHostIPMapped`/登录计数）、WS 探测
- **操作人**：controller 经 `auditContext(c)` 从 JWT `user_id` 取快照（`operator_id` + `operator_name`）
- **request_id**：同一 HTTP 请求内多条审计共用（批量/CSV 导入逐行共用一个 ID）
- **detail**：`{"before":{...},"after":{...}}` 前后值 diff；不含 password_hash 等敏感明文
- **只增不删**：无 UNIQUE 约束、不设 users 外键（操作人以快照落库）
- **查询**：`GET /api/audit-logs`（分页筛选，通用查询）与 `GET /api/audit-logs/resource`（按 `resource_type`+`resource_id` 取单资产变更记录，T-066）；写入失败不阻断业务（审计不改变 CUD 结果语义）
- **展示**：资产详情弹窗「变更记录」tab（`AuditTimeline` 组件）按时间倒序渲染操作人/动作/diff；主机点击行打开详情（`view` 模式），零信任/映射点击行打开只读详情弹窗


## AI 接入配置与密钥加密

> `T-067` 后端地基：供后续「上传申请表单 → 大模型提取 → 确认入账」（T-068~T-070）复用的配置与密钥管理。

- **存储**：`ai_settings` 单条全局配置（主键恒为 1，全站共用），字段 `api_url` / `api_key_enc` / `model`
- **加密**：API Key 以 **AES-256-GCM** 加密后落库（`utils/crypto.go`），密文格式 `base64(nonce||ciphertext)`；密钥由环境变量 `AI_SECRET_KEY` 经 SHA-256 派生为 32 字节
- **降级**：未配置 `AI_SECRET_KEY` 时 `Encrypt/Decrypt` 按明文原样返回并记录一次告警（不阻断功能）；构建文件（docker-compose）是否透传该变量由人类决定
- **回显**：`GET /api/ai/settings` 只返回脱敏值 `api_key_masked`（保留末四位，如 `****1234`）与 `api_key_set` 布尔，**绝不返回明文**
- **保存**：`PUT /api/ai/settings` 中 `api_key` 传空表示**保留原值**（因前端只拿到脱敏值，无法回传明文），非空则加密覆盖；`api_url`/`model` 必填
- **审计**：保存变更写 `audit_logs`（`resource_type=ai_setting`），detail 仅记录 `api_url`/`model`/`api_key_set`，不落 API Key 明文
- **内部使用**：后端调用大模型经 `AISettingService.GetDecrypted` 取明文 Key，不下发前端

## AI 申请表单提取与文件归档

> `T-068`：上传申请表单（docx/xlsx）→ 大模型识别类型并提取字段 → 返回前端预览；源文件暂存，确认入账成功后归档。

1. **解析**：`AIDocParser.Parse` 按扩展名分派——`.docx` 用标准库 `archive/zip` + `encoding/xml` 读 `word/document.xml`，提取段落与表格（表格转 markdown 行）；`.xlsx` 复用 `excelize` 逐 sheet `GetRows` 转文本。文本上限 24000 字符（超出截断）
2. **提取**：`AIExtractService.Extract` 读 `ai_settings` 取明文 Key → 注入 `buildExtractionPrompt`（含两类模板字段清单）→ `AILLMClient.Complete` 调 `{api_url}/chat/completions`（Bearer Key、`response_format=json_object`、temperature 0、60s 超时）→ `parseLLMOutput` 归一化（容忍 ```` ```json ```` 代码块包裹）
3. **类型识别**：`doc_type` ∈ `host_apply`（主机申请）/ `zero_trust_apply`（零信任申请）；未匹配返回 `unknown` 并在 `warnings` 追加提示。字段名严格对齐各模块 `CreateRequest`，前端确认后直接调现有 Create 入库
4. **文件暂存与归档**（`AIFileStore`，根目录 `AI_UPLOAD_DIR`，默认 `./data/ai_uploads`）：
   - `extract` 时原文件写入 `<root>/incoming/<token>__<清洗后文件名>`，返回随机 hex `file_token`
   - 确认入库成功后前端调 `POST /api/ai/archive`，文件移入 `<root>/archived/<YYYY-MM-DD>/<时间戳>_<文件名>`，token 随即失效
   - 安全：token 仅允许 32 位小写 hex；文件名经 `sanitizeFilename` 去路径分隔与 `..`，定位文件按 token 前缀匹配，杜绝路径穿越
5. **不写业务库**：提取阶段仅返回预览，实际入账复用 `POST /api/hosts` / `POST /api/zero-trusts` 的既有校验与审计


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
db.Preload("Application").Preload("Person").Offset(offset).Limit(pageSize).Order("id DESC").Find(&hosts)
```

`page_size` 上限为 100（超出会被重置为默认值）。

`is_external`（T-078）：`1` 仅云外主机、`0` 仅云内主机、留空不过滤；主机管理页默认传 `0`（只看云内），工具栏「主机位置」可切换。响应 `internal_count`/`external_count` 为**全库**拆分计数（不受筛选影响，T-084），页脚显示「全库共 (云内+云外) 台（云内 X 台 / 云外 Y 台）」；分页 `total` 仍为命中筛选的条数。

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
    "状态", "标签", "申请单位", "申请人", "申请人联系方式",
    "所属项目", "申请理由", "申请配置", "申请时间", "备注", "对象存储(GB)", "是否云外主机",
}
```

导出 CSV 需带 BOM（`\ufeff`）以便 Excel 正确识别 UTF-8。

### 申请时间列校验

`申请时间` 列为纯日期，严格 `YYYY-MM-DD`（空值合法）。非法值该行导入失败：CSV 文件导入计入 `errors` 且不入库，纯文本批量返回行级 `line_errors`，错误信息含 `YYYY-MM-DD`（接口与校验规则见 [API.md · 日期时间格式](./API.md#日期时间格式)）。

## 数据备份 xlsx 导出/导入

与「CSV 导入导出」（仅主机单表）互补，设置菜单「数据备份」进入 `/data-backup`，提供 8 个业务 sheet 的整库备份。

### 导出

- `GET /api/export/all` 返回单个 `.xlsx`（excelize，8 个 sheet 按导入依赖序排列）；加密在**前端**完成（T-074，见下「导出加密」）：`人员` → `公网IP` → `云资源` → `网段` → `主机` → `主机申请信息` → `零信任` → `端口映射`（T-063 起 sheet 名为中文，导入按中文名匹配、不再兼容旧英文名备份）
- 列头为中文、按名匹配；不导出 `users`、自增 id 与时间戳；关联用自然键表达：`hosts` 人员列（姓名+联系方式，由 `person_id` 解析）、宿主一律用「内网IP」；**零信任按「账户名 + 接入地区 + 内网IP 一行」导出（T-073/T-077）**：`zero_trust_targets` 子表（`host_id` → 内网IP）按账户名、接入地区与内网IP 聚合，tcp/udp 端口分别并集后写入 `tcp端口`/`udp端口` 两列（端口列支持掩码 `*`＝全端口 1-65535、逗号列表、`A-B` 范围，可混用，导出按区间升序合并规整）；不同接入地区或不同内网IP 各占一行、其余字段取首条非空；导入按「账户名 + 接入地区」合并、接入目标并集去重，`内网IP` 列沿用 `;` 分组广播（不做 CIDR 展开），两端口列均空为行错误；**T-073 起不再兼容旧「接入目标」列备份**
- 派生列不导出：`hosts.ip_mapped`、`port_mappings` 的运营商/出口位置（由公网IP资源池带出）；`port_mappings` 导出「外网协议/内网协议」列（与端口等长并行，T-055）；`zero_trusts` 导出「接入地区」列（T-061 独立存储，替代原「公网IP」列）

### 导出加密（T-074）

- 导出**必须设置密码**（前端校验，缺省/少于 6 位不允许下载）
- 加密方式：前端 `utils/xlsxCrypto.js` 用 `officecrypto-tool` 把后端返回的 xlsx 字节按 **ECMA-376 标准 Office 加密**（Agile）封装，下载文件名 `mcloud_backup_YYYY-MM-DD.xlsx`；该文件可用 Excel/WPS 输入密码打开、另存修改后重新导入
- 导入：若填写了打开密码，前端先 `isEncryptedXlsx`（OLE magic + `officeCrypto.isEncrypted`）判断并按需 `decryptXlsx` 解出明文 xlsx 再上传；密码错误直接提示且不上传；未加密 xlsx 可留空密码直接导入。后端只接收明文 xlsx，收到未解密的加密文件返回 `40001`

### 导入（upsert）

- 结构校验先行：非 xlsx（zip `PK` 头，含未解密的加密 Excel）、缺 sheet、缺列、表头重复直接返回结构错误（`40001`），不进入事务；`主机` sheet「对象存储」列兼容旧备份列头「磁盘」（T-072）
- 逐行解析复用既有规范化：`normalizeCIDR`（/24）、`parsePortExpr`（单端口或 `A-B` 连续范围 token 规范化，T-055/T-056；`expandPortCount` 展开后数量等长校验在 `applyPortMappings` 与 `normalizePortMapping`，范围可与单口混搭）、`parseProtocols`（协议列缺省 tcp、仅 tcp/udp、与端口 **token** 等长，T-055）、`normalizeTargetPairs`（配对去重，端口为表达式并兼容旧 `port` 数字入参，协议缺省 tcp，T-056/T-065）、`splitPortGroups`（零信任批量申请端口列分组：列值含 `;` 按 `;` 分组表达多IP各自多端口，无 `;` 时顶层逗号+引号分组；T-059 无 `;` 且仅 1 个端口组时广播到全部 IP——多IP相同端口写 `22` 即可、相同多端口列表用重复分组 `22,443;22,443`；T-065 端口组可带 `/协议` 后缀，单主机多端口协议须一致）、`parsePortSet`（零信任 `tcp端口`/`udp端口`：`*`、列表、`A-B` 范围可混用并合并规整，T-073）、`ValidateDateOnly`（`host_applications.申请时间` 严格 `YYYY-MM-DD`）；人员按姓名(+联系方式)解析 `person_id`、主机按内网IP解析 `host_id`、映射行填写公网IP时必须在 `public_ips` 资源池（空值合法放行）；零信任「接入地区」为独立字段（T-061，可多值逗号分隔、不做强制校验，T-065）
- 申请时间口径统一只到日期（T-065）：`host_applications` 与 `zero_trusts` 均存 `YYYY-MM-DD`（API/批量/表单严格 `utils.ValidateDateOnly`，`zero_trusts` 空值缺省服务器当天）；**xlsx 导入宽容归一**（`utils.NormalizeDateLenient`：接受 `2026/9/10`、`2026-9-1`、`2026.9.10`、`2026年9月10日`、`20260910` 等，归一化为 `YYYY-MM-DD`，两位年份与带时刻报错）；审计类 `created_at`/`updated_at` 等保留完整时间戳
- 每个 sheet 按自然键定位已有记录：值一致 → `skipped`，不一致 → `updated`，不存在 → `created`；`zero_trusts` 自然键为**账户名 + 接入地区**（T-077：地区规整去重后排序整串比较，`南宁,青年` 与 `青年,南宁` 等价、空地区自成一组；不同地区各自独立成记录；T-058 合并语义：同键下接入目标 `unionTargetPairs` 并集去重、`account_name_zh`/`applicant`/`contact`/`system_name`/`remark` 空补录非空保留、`apply_time` 保留已有，同文件多行同账户名同样合并，无变化 → `skipped`；`POST /api/zero-trusts` 新增与批量行走同一合并口径，`PUT` 编辑不合并，存量同名多条不迁移、合并目标取 id 最小一条）；**仅新增与更新，不执行删除**
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

纯 Go ICMP Echo（T-011，`golang.org/x/net/icmp`）+ TCP 22/3389（超时均 100ms），任一有回包即视为 alive。

- ICMP 建连顺序：非特权 `udp4` ping socket → `ip4:icmp` 原始 socket（需 `CAP_NET_RAW`，容器内 root 默认具备）；环境均拒绝时 ICMP 判不通，由 TCP 22/3389 兜底
- 回包校验：`Echo Reply` + 来源 IP 等于探测目标，在 100ms 超时窗口内循环读取过滤（原始 socket 可能混入他进程回包）
- **不再依赖宿主机 `ping` 命令**，消除平台差异（Linux `-W` 单位问题随之消除）

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
| # | 50px | 序号（基于当前页连续计数，与数据库 id 无关） |
| 区域 | 100px | |
| 主机名称 | min 140px | 含环境标签 + DB标签 |
| 内网IP | 140px | |
| 是否映射公网 | 120px | 布尔标签 是/否 |
| 资产类型 | min 120px | 含 CPU 架构标签 |
| 主机位置 | 名称列标签 | `云外` 标签（T-078，`is_external=true` 时显示） |
| 规格 | min 90px | `4vCPU/8G/100G/50G`（CPU/内存/系统盘/数据盘），缺失显示 `-` |
| 对象存储 | 90px | 技术信息 `hosts.object_storage`（GB），空显示 `-`（T-072/T-080） |
| 状态 | 90px | 颜色徽章 |
| 项目 | min 120px | 申请信息表，tooltip |
| 申请人 | min 100px | 申请信息表，tooltip |
| 操作 | 140px | 固定右侧 |

### 申请时间默认值（T-082）

- 主机（新增/编辑/复制/详情）与零信任（新增/编辑/复制）弹窗打开后，若「申请时间」为空（新增记录或历史数据缺值）则默认填充**当天**（`utils.formatDate()`，`YYYY-MM-DD`）
- 已有值一律保留原值，不做覆盖；后端仍按既有口径校验（空值缺省服务器当天）

### 区域输入口径（T-079）

- 区域**不做硬编码**：表单「区域」为可创建下拉（`filterable` + `allow-create`），候选取 `GET /api/hosts/regions`（数据库已有区域去重），也可直接输入新区域后保存
- 主机管理「区域」筛选与批量编辑「区域」均为下拉选择（同样取数据库已有区域），仅支持筛选已存在的区域
- 打开表单/批量编辑时拉取一次区域候选，创建或修改成功后再次刷新，保证新区域立即可选

### 资源统计口径

- 资源统计（`cloudResource` store 聚合）与业务统计（`GetBusinessStats`）**排除云外主机**，与裸金属排除口径一致（T-078）；主机管理页页脚按云内/云外分别计数
- ECharts 图表注册（`composables/useChartOption`）内置 `CanvasRenderer` 与 `GraphicComponent`，`graphic`（如饼图「暂无数据」占位文字）无需各页额外注册（T-084）
- **裸金属服务器与物理机**（T-079）不计入 CPU/内存/存储 汇总（避免与云资源总量重复），主机台数仍计入

### 列表排序

- 主机列表后端固定 `ORDER BY id DESC`（最新创建在前），分页跨页顺序一致；前端序号列仅按当前页位置连续计数（T-072）

### 表单约定

- CPU核数、内存、对象存储(GB)、系统盘、数据盘：普通 `el-input`，数字校验正则 `/^\d*$/`，提交时 `Number()` 转换

# 任务队列（ROADMAP）

> **本文件是任务状态的唯一事实来源（source of truth）。** [AGENTS.md](../AGENTS.md) 中的「状态快照」只是它的缓存，两者不一致时**以本文件为准**，并修复 AGENTS.md。
> 执行协议（如何领任务、如何定位代码、如何回写）见 [AGENTS.md · 会话协议](../AGENTS.md#会话协议)。

## 字段约定（AI 必读）

| 字段 | 取值规则 |
|------|----------|
| ID | `T-001` 形式，全局唯一，**永不复用**（done 后也不再启用） |
| 状态 | 仅允许 `todo` / `in_progress` / `blocked` / `done` 四个值，状态变化只在本文件发生 |
| 负责 | 进入 `in_progress` 时必须填执行者标识（AI 会话名或人名）+ 认领日期 |
| 优先级 | `P0` 阻塞他人 > `P1` 当前迭代 > `P2` 可延后；同级按 ID 顺序 |
| 内容 | 任务目标与做法概述（一段式，可带子项） |
| 验收标准 | 客观可判定（命令通过 / 文件存在 / 行为可复现），禁止主观描述 |
| 定位 | 涉及文件清单，来源 `docs/CODE_INDEX.md`「功能→代码映射」；**领任务后必须核对/补齐**。优先写**函数名/路由路径**等稳定锚点；行号仅为快照（以 grep 复核为准，可省略），补不出先声明、不盲写 |
| 回写 | 完成后必须更新的文档（通常由 AGENTS.md 责任矩阵推导，此处列出例外） |
| 分支 | 功能分支名（`feat/fix/docs/refactor/<描述>`）；未写则按 AGENTS 红线 8 推导 |
| 文风 | done 备注只写验收结论（≤ 80 字）；PROJECT_STATUS 变更记录每条一行摘要（≤ 120 字），**禁止** ①②③ 枚举子改动与粘贴验收全文（细则见 [DEVELOPMENT.md](./DEVELOPMENT.md) 文档回写文风） |

**认领规则**：一次只领 **1 条**优先级最高（`P0` > `P1` > `P2`）的任务；领用 = 把该条移到 `in_progress` 并填写负责字段（在写任何代码之前先改本文件）。

**回顾锚点**：`上次回顾至 T-053`。`done` 表自该锚点起新增 ≥ 5 条时，须先执行 [AGENTS.md · 会话协议 E](../AGENTS.md#会话协议) 的周期回顾（按各任务「回写」字段定向核查文档↔代码符合度、补回填写漏，纯文档类任务跳过，偏差先报告人类再修），回顾完成并经人类确认后把锚点推进至该批最后一条 `done` ID。

---

## in_progress

## blocked

（无。填写格式：`- [ ] **T-xxx** … ｜ blocked 原因：具体条件` —— 泛泛的"困难"不算 blocked）

## 已取消

| ID | 任务 | 取消时间 | 原因 |
|----|------|----------|------|
| T-037 | 零信任 `apply_time` 无时区格式解析失败 | 2026-10-09 | 人类裁定并入 T-040（全仓日期格式统一）一次执行 |
| T-007 | CI 流水线落地 | 2026-10-06 | 项目暂不引入 CI/CD；所有集成与检验由 agent 完成任务后本地执行「验证命令」（见 [AGENTS.md](../AGENTS.md#验证命令门禁)） |

## done

| ID | 任务 | 完成时间 | 提交 | 备注 |
|----|------|----------|------|------|
| T-073 | 数据备份「零信任」sheet 改为「账户名 × 内网IP 一行」＋ tcp端口/udp端口 双列 | 2026-10-20 | `20b8d0d` | 验收：导出按账户名+内网IP 分组、tcp/udp 端口并集规整；`*`/列表/范围解析与掩码展开单测；往返不丢+二次导入幂等；旧「接入目标」列报结构错误；dev 库实测；门禁全绿 |
| T-072 | 主机「磁盘」改名对象存储（含 API 字段改名）＋主机列表对象存储列／创建时间倒序／数据备份导出强制密码加密（导入可选密码） | 2026-10-20 | `4810e48` | 验收：`hosts.disk`→`object_storage` 迁移（dev 实测）+ 全链路改名；列表「对象存储」列取申请信息、`id DESC` 倒序（单测）；加密导出 `.mcbak`（无密码 40001/错密码 40001/明文可无密码，单测+dev 实测）；三端门禁全绿 |
| T-071 | AI 接入文档回写与门禁 | 2026-10-16 | `5f12eed` | 验收：check_docs 0 errors；ROADMAP done 表 T-067~T-070 提交号已回填；AGENTS 快照全量对齐（表数 12、页面 11、todo 清空）；三端门禁（build/lint/test/lint:check/build）全绿 |
| T-070 | 前端 AI 申请入账页 | 2026-10-16 | `a6b2c72` | 验收：`/ai-import` 上传 docx/xlsx 带真实上传进度；按识别类型动态渲染可编辑表单（主机/零信任）含 warnings 高亮；确认复用现有 Create 入账（零信任接入目标内网IP→host_id）、成功后归档源文件；11 页齿轮菜单入口；页面数 10→11；lint:check + build 全绿 |
| T-069 | 前端 AI 接入设置页 | 2026-10-16 | `ae983cd` | 验收：`/ai-settings` 可读写、Key 回显脱敏（`****1234`）且留空保留；9 页齿轮菜单均有「AI 接入设置」入口；页面数 9→10；lint:check + build 全绿 |
| T-068 | 后端 AI 表单解析、提取与文件暂存归档 | 2026-10-16 | `12881ef` | 验收：docx 段落/表格、xlsx 多 sheet 转文本单测过；extract 编排（fake LLM）识别 host_apply/zero_trust_apply 并归一化字段、unknown 加告警；incoming 暂存/archived 归档/token 失效/路径穿越拒绝单测过；未配置 AI/非法类型/解析失败返回 40001（dev 实测）；门禁全绿 |
| T-067 | 后端 AI 接入地基：`ai_settings` 单条全局配置表 + AES-GCM 加密 Key + `GET/PUT /api/ai/settings` | 2026-10-16 | `b7adc68` | 验收：迁移幂等且 dev 库实测应用；GET 返回 key 仅末四位（`****1234`）、PUT 空 key 保留原值、AES-GCM 加解密往返（单测覆盖）；settings 脱敏/保留原值/校验单测 + 加密降级/错误密钥单测；门禁全绿 |
| T-066 | 审计改为资产变更记录：移除独立审计页，主机/零信任/映射新增/修改写入字段快照，详情弹窗「变更记录」时间线按旧值→新值逐字段展示，点击资产行打开只读详情 | 2026-10-16 | `c3aa75d` | 验收：`/audit-logs` 页/路由/Tab 移除、`AuditLog.vue` 删；新增 `GET /api/audit-logs/resource`（按 resource_type+resource_id）；主机 create/update/delete/batch 与零信任/映射 create/update 均写统一字段快照（before/after）；主机 view 模式加「变更记录」tab、零信任/映射新增只读详情弹窗含该 tab；三页点击行打开详情、去「详情」按钮；单测覆盖按资源查询过滤与主机 update diff；门禁全绿 |
| T-065 | 零信任接入目标增加 TCP/UDP 协议（每行单协议，缺省 tcp）＋接入地区改多值＋主机下拉拉全量＋导出按接入目标分行＋导入报告层级修复＋申请时间只到日期 | 2026-10-16 | `bd96b7c` | 验收：protocol 迁移加列幂等+回填 tcp、CRUD/合并/批量 `/协议`/xlsx 往返不丢；exit_location `VARCHAR→TEXT` 多值迁移幂等、逗号/顿号规整去重、合并并集、前端多选+标签、批量/xlsx 多值往返；主机下拉分页拉全量（>100 台可选全）；xlsx 零信任按每个接入目标一行导出、导入合并往返不丢；数据备份导入报告取 `res.data`（修「后端成功却提示回滚」）；`apply_time` TIMESTAMPTZ→VARCHAR(10) 纯日期迁移幂等、API/批量/导入导出/前端统一 YYYY-MM-DD、xlsx 导入宽容归一（2026/9/10 等）；列表「申请资源」列按主机聚合（≤3 主机 tag + 可滚动 popover 端口明细，不显示端口数量、行高恒定、气泡限高且 auto 定位+flip 自动换方向防出界）；非法协议/非法日期报错；单测+门禁全绿 |
| T-064 | 零信任 `contact` 标签改「账户联系方式」＋列表「申请人」列移至「申请时间」之后 | 2026-10-16 | `c88f913` | 验收：列表列改名「账户联系方式」且申请人在申请时间之后；表单/批量/xlsx 列头同步；导入按新列头匹配；门禁全绿 |
| T-063 | 零信任新增「账号姓名」字段（`account_name_zh`，账户名补充）＋列表列调整；数据备份 sheet 名改中文 | 2026-10-16 | `db0762f` | 验收：迁移加列幂等；CRUD/账户名合并补录/批量第 3 列/xlsx 往返/搜索命中均有单测；列表账户名/账号姓名/申请人；sheet 名中文（不兼容旧英文名备份）；门禁全绿 |
| T-062 | 主机管理/零信任/映射三页行「复制」按钮（预填所选行、唯一字段置空的新增窗） | 2026-10-16 | `5d31568` | 验收：三页操作列有「复制」且进新增模式；预填行信息、唯一字段（主机内网IP+实例ID/零信任账户名/映射公网IP+主机+端口）为空；提交走 create；lint:check+build 全绿 |
| T-061 | 零信任去「公网IP」改独立「接入地区」（地区从公网IP资源池去重筛选）+ 新增弹框加宽 | 2026-10-16 | `2964e18` | 验收：迁移 DROP public_ip+ADD exit_location 幂等；`/exit-locations` 去重；CRUD/批量/xlsx 接入地区全链路；public_ips 引用不再含零信任；弹框 520→550px；单测+门禁全绿 |
| T-060 | 零信任台账新增「申请人」字段（与账户名区分） | 2026-10-15 | `69bdbd0` | 验收：`applicant` 迁移幂等+dev 实测；申请人独立存储/回显、账户名合并空补录非空保留、批量第 3 列、xlsx 往返、搜索命中均有单测；门禁全绿 |
| T-059 | 零信任批量「多IP相同端口」单组广播 | 2026-10-12 | `dea43d5` | 验收：单组广播/重复分组/`;` 按位/`;` 单组+多IP 与组数不匹配报错均有单测；旧数量用例改真不匹配；门禁全绿 |
| T-058 | 零信任批量导入支持多IP多端口 + 新增/导入按账户名合并 | 2026-10-12 | `22a6fcf` | 验收：`;` 分组/单IP整列多端口/旧格式回归、账户名合并并集+补录、xlsx 自然键改账户名单键均有单测；门禁全绿 |
| T-053 | `models/audit_log.go` NOT NULL 口径对齐 | 2026-10-12 | `39b8539` | 验收：六字段 tag 与迁移逐字段一致（`resource_id`/`detail` 保持可空）；ARCHITECTURE 已是 NOT NULL 口径无需改；门禁全绿 |
| T-057 | 修复零信任 ports 迁移 goose DO 块切分失败 | 2026-10-12 | `81b0eb0` | 验收：StatementBegin/End 包裹；fresh 全量回放+down/up 往返+dev 后端重启三路实测恢复；门禁全绿 |
| T-056 | 零信任接入目标改单主机多端口表达式 | 2026-10-10 | `28d829c` | 验收：`ports` 列迁移 dev 库幂等实测；targets 结构+旧 port 兼容；批量引号分组、xlsx 续接；搜索/删引用回归+门禁全绿 |
| T-055 | 映射端口区分 TCP/UDP 协议并支持 A-B 连续范围 | 2026-10-10 | `afada3d` | 验收：并列协议列缺省 tcp、范围展开后配对、唯一索引含外网协议列、判重含协议维度；xlsx/批量全链路；门禁全绿 |
| T-008 | 数据导出报表（按筛选导出） | 2026-10-10 | `c71d422` | 验收：ExportCSV 复用列表筛选（applyHostFilter 共享），导出行数=列表 total 单测覆盖；前端传当前 filters；API.md 同步；门禁全绿 |
| T-009 | 密码哈希升级 bcrypt | 2026-10-10 | `c27e306` | 验收：旧 `salt$hash` 用户登录成功后自动升级 bcrypt（单测覆盖）；bcrypt cost 10，`seedAdmin`/`resetpw` 同步新格式；**无需迁移 SQL**（`password_hash VARCHAR(256)` 兼容两种格式，平滑升级在登录时完成，技术选型推荐结论）；门禁全绿 |
| T-011 | IP 探测去宿主机 ping 依赖 | 2026-10-10 | `dfd2ac7` | 验收：纯 Go ICMP（`x/net/icmp`，udp4 ping socket → ip4:icmp 原始 socket → TCP 22/3389 兜底，100ms 超时）替代 `exec ping`；颜色状态机单测覆盖 dead 分支 + alive 分支（环境无 ICMP socket 时跳过）；BUSINESS_LOGIC「探测方式」回写；门禁全绿 |
| T-012 | 文档一致性巡检 | 2026-10-10 | `aba8246` | 验收：check_docs 0/0；人工抽查路由 47/47 全匹配（gin Routes() dump 对照）、错误码 5/5、CSV 27 列三方一致、筛选参数 10 项、T-id 引用 53/53、页面路由 11/11、表/模型 10/10、关键常量（5次/15分/100ms/32并发/3s重连/10×26）；修复 2 处口径偏差（AGENTS 位图术语、T-019 定位处数） |
| T-019 | 前端抽 Layout/AppNav | 2026-10-10 | `481477e` | 验收：`components/AppNav.vue` 承载 7 Tab 导航，10 页重复结构仅剩一处；`variant` prop（default/audit/backup）保留三套历史样式实现视觉零变化；导航项 10 页本就同为 7 项（统计三页硬编码 active 改为动态等价）；`lint:check`+`build`+`vitest 6/6` 全绿 |
| T-020 | config ServerPort 硬编码消除 | 2026-10-10 | `8535979` | 验收：`SERVER_PORT` 环境变量可配（`getEnv` 模式，保留默认 `5677`——推荐口径，与 DB_* 等其余字段一致、开发零配置；docker-compose.dev 已传该变量、`.env.example` 已同步）；技术债表该行删除；`go build`+`golangci-lint`+`check_docs` 全绿 |
| T-022 | WS 探测通道 token 有效性校验 | 2026-10-10 | `a0e6477` | 验收：升级前复用 `middleware.ValidateToken`（与 Authorization 同源），缺失/无效/过期/伪造返回 `40101` 拒绝升级；有效 token 建连回归（middleware+routes 单测）；API/BUSINESS_LOGIC/CODE_INDEX/ARCHITECTURE 同步；门禁全绿 |
| T-043 | CORS Origin 白名单收紧 | 2026-10-10 | `63bda34` | 验收：`CORS_ALLOWED_ORIGINS`（逗号分隔精确匹配，默认空）白名单化，非白名单 Origin 不回显 CORS 头；白名单预检/带凭证正常（单测覆盖）；DEVELOPMENT/.env.example/CODE_INDEX 同步；门禁全绿 |
| T-044 | 后端 CSV 上传大小限制 | 2026-10-10 | `176563c` | 验收：`Import` 加 `MaxBytesReader` ≤16MB（对齐前端提示），超限返回 `40001`「文件大小超过 16MB 限制」；未超限穿透业务校验（3 例单测）；DEVELOPMENT/API.md 同步；门禁全绿 |
| T-054 | docker-compose 透传 CORS 变量 | 2026-10-10 | `d487b0b` | 验收：dev/prod backend `environment:` 各加 `CORS_ALLOWED_ORIGINS` 透传；`docker compose config` 双文件解析通过；DEVELOPMENT 注明 compose 已透传；check_docs 0/0 |
| T-052 | `zero_trusts.targets` 拆子表 | 2026-10-10 | `d12f354` | 验收：API `targets` 结构零变化；子表双 FK RESTRICT、回填迁移落地；40901/按主机搜索/空库回灌均有单测；门禁全绿 |
| T-006 | 测试骨架 | 2026-10-10 | `894e93b` | 验收：`go test` services 表驱动（主机CRUD/登录锁定）+ utils 密码全绿（sqlite 内存库注入）；vitest 冒烟 6 例；门禁全绿 |
| T-005 | `IpStatistics.vue` 拆分 | 2026-10-10 | `baa8e2e` | 验收：778→300 行（≤400）；拆 components/ip/（ProbePanel/SubnetCard/SubnetManageDialog 移入）+ composables/useProbeGrid；lint+build+check 全绿 |
| T-004 | 前端 composables 抽取 | 2026-10-10 | `c1d4618` | 验收：抽 useProbeWebSocket/useChartOption/usePagedTable 三件，IpStatistics 841→778 行；lint:check + build + check_docs 全绿，页面行为不变 |
| T-050 | `hosts.disk` 补录入口 | 2026-10-10 | `6b62e8d` | 验收：表单磁盘输入框+校验回显；CSV 27 列导出/导入与 xlsx hosts sheet 含磁盘列，27 列往返值不变、旧 26 列兼容 disk=0；门禁全绿 |
| T-051 | 公网 IP 引用保护 + 端口映射防重 | 2026-10-10 | `a3a19ec` | 验收：被引用公网IP删/改IP返回 40901；重复整组映射 Create 返回 40901 且唯一索引存在；dev 库迁移幂等实测；门禁全绿 |
| T-049 | 删除 `port_mappings` 冗余列 `isp`/`exit_location` | 2026-10-10 | `f325d4d` | 验收：dev 库 DROP COLUMN 重跑幂等实测；List 响应 isp/exit_location 恒等于资源池；模型 gorm:"-" 派生字段；回填与双表 LIKE 已移除；门禁全绿 |
| T-048 | 数据库迁移 SQL ↔ 实际库漂移修复 | 2026-10-10 | `266bf66` | 验收：对齐迁移 dev 库执行+重跑幂等实测；主键全 bigint、created_at 全有默认、唯一索引全 idx_ 前缀、FK 全 NO ACTION；模型/ARCHITECTURE 口径同步；门禁全绿 |
| T-047 | 全资源增删改审计日志 | 2026-10-10 | `76d4e8a` | 验收：audit_logs 只增不删（普通索引、无 UNIQUE/users 外键）；26 个 CUD 方法经 operator+request_id 写入、detail 前后 diff 无敏感明文；GET /api/audit-logs 分页筛选 + /audit-logs 页；门禁全绿 |
| T-042 | 批量编辑 `apply_time` 错误码与空值语义修复 | 2026-10-10 | `8a3940e` | 验收：null 与省略同义放行；非字符串/非法日期统一 40001；API.md 同步；门禁全绿 |
| T-010 | `routes/routes.go` 按域拆分 | 2026-10-10 | `1f0b6a3` | 验收：12 文件域拆分、JWT 收口受保护组；路由全表 A/B 67/67 一致；ARCHITECTURE/CODE_INDEX/T-022 定位已回写 |
| T-003 | `controllers/csv.go` 业务逻辑下沉 | 2026-10-10 | `f0643f8` | 验收：后缀/解析/行循环/去重/裸查全下沉 service，controller 仅绑定+错误码映射；A/B 8 场景字节一致；门禁全绿 |
| T-002 | 后端 `services/host_service.go` 拆分 | 2026-10-10 | `e039190` | 验收：拆为 CRUD/批量/CSV 三文件（9/5/4 函数）；签名零变化、API 行为不变；build+lint+check_docs 全绿 |
| T-046 | 零信任列表「申请主机+申请端口」合并为「申请资源」标签列 | 2026-10-09 | `9c1cbfa` | 验收：两列并一列标签 `主机名(ip:port)` 按主机名排序；前端门禁+check_docs 全绿；API/DB 零改动 |
| T-045 | 协议补强：周期回顾偏差核查跳过纯文档类任务 | 2026-10-09 | `5fbb249` | 验收：AGENTS E.2/DEVELOPMENT/ROADMAP 三处口径同步纯文档类任务跳过；check_docs 0 errors |
| T-041 | 协议补强：每完成 5 条任务的周期回顾 | 2026-10-09 | `4ccd241` | 验收：协议 E 触发/范围/处理三要素落地；锚点 T-040 两处一致；DEVELOPMENT 同步；check_docs 0 errors |
| T-040 | 全仓日期时间格式统一（吸收 T-037） | 2026-10-09 | `9522789` | 验收：零信任时间双格式兼容；主机申请时间全链路严格 `YYYY-MM-DD`；展示与文件名口径回写 API.md；门禁全绿 |
| T-039 | 空库启动迁移失败修复（UNIQUE 约束对齐 GORM uniqueIndex 口径） | 2026-10-09 | `c196f37` | 验收：空库启动+登录通过；4 表 `*_key` 转 `idx_`；外键/重复索引去重；dev 迁移 no-op；门禁全绿 |
| T-038 | 数据备份：8 个业务 sheet 统一导出(xlsx)/导入(upsert) | 2026-10-09 | `6644d5a` | 验收：8 sheet 列=设计清单；空库回灌关联完整重建；重复导入 0 增删；行级错误回滚库不变；门禁全绿 |
| T-036 | 零信任台账多组主机:端口配对 + 公网IP带出地区 | 2026-10-08 | `ab7e3d0` | 验收：targets 配对 CRUD+迁移；公网IP池校验带出地区；批量等长配对；引用40901；门禁+冒烟全绿 |
| T-035 | 映射台账公网IP改为资源池选择 | 2026-10-08 | `627ee8f` | 验收：公网IP下拉选资源池；表单无运营商/出口位置；列表从资源池带出 |
| T-034 | 域名台账改造为端口映射台账 | 2026-10-08 | `d26d6ad` | 验收：port_mappings 多端口等长；ip_mapped 自动；/mapping-ledger；迁 domains 数据 |
| T-033 | 域名/公网IP出口位置字段 | 2026-10-08 | `aac7fea` | 验收：exit_location；运营商之后；域名+公网IP+批量列；迁移+验证全绿 |
| T-032 | 零信任/域名台账批量添加 | 2026-10-08 | `4e96c09` | 验收：两台账 /batch 文本接口；主机按内网IP；域名跳过零信任全插；前端工具栏可用 |
| T-031 | 公网 IP 资源录入 | 2026-10-08 | `e0d84ac` | 验收：public_ips CRUD+IP唯一/格式校验；设置菜单+独立页；迁移+验证全绿 |
| T-015 | 公网 IP 字段改造 + 双台账关联 | 2026-10-08 | `130388a` | 验收：ip_mapped 布尔；详情展示两台账关联；CSV/表单/筛选同步；迁移+验证全绿 |
| T-030 | 域名台账字段改造 | 2026-10-08 | `e83db90` | 验收：isp/host_id/host_port；无 provider/expires_at；FK 校验+主机删除 40901；迁移+验证全绿 |
| T-029 | 零信任台账增加系统名称字段 | 2026-10-08 | `1f70851` | 验收：API 含 system_name；列/表单位于申请端口后；迁移+前后端验证全绿 |
| T-016 | Goose 迁移执行器实装 | 2026-10-08 | `82e0e07` | 验收：存量 dev 库首跑 9 迁移+幂等；build/lint/check_docs 过；口径反转已回写 |
| T-028 | 前端导航修正：台账入口与 Tab 一致性 | 2026-10-08 | `2e32a5b` | 验收：设置菜单无台账入口；各页 Tab 含零信任+域名；lint/build 通过 |
| T-027 | 前端 ESLint 警告清零 | 2026-10-08 | `ebb52de` | 验收：lint:check 0 problems + build 通过 |
| T-026 | 协议补强：多任务串行合并 | 2026-10-08 | `c5208c6` | 验收：AGENTS A.2/A.4 + 分支协议 + DEVELOPMENT 口径一致；check_docs 0 errors |
| T-014 | 域名台账 | 2026-10-08 | `4e0190c` | 验收：5 字段 CRUD + 域名唯一；前后端验证通过 |
| T-013 | 零信任台账 | 2026-10-08 | `6a6e760` | 验收：7 字段 CRUD + 主机 FK 校验；前后端验证通过 |
| T-025 | 协议补强：临时任务入队 + 提交权限时序 + 分支/接管规则 | 2026-10-06 | `1837874` | 验收：A.0/A.4/D 与红线 5 例外落地；提交号时序兼容 shas；check_docs.sh 0 errors |
| T-024 | 文档回写文风上限（防膨胀） | 2026-10-06 | `bcfb410` | 验收：`check_docs.sh` 0 errors；三处口径一致；超长条目已压缩 |
| T-023 | 文档卫生清理 + 移除 CI/CD 规划内容 | 2026-10-06 | `42d4c10` | 验收：`check_docs.sh` 0 errors；规划语境无 T-007/CI 残留；协议改动经人类授权 |
| T-017 | 文档事实修正与协议补强 | 2026-10-06 | `6187228` | 验收：`check_docs.sh` 0 errors；协议条文改动仅限任务列明项 |
| T-021 | 文档一致性自动校验脚本 `scripts/check_docs.sh` | 2026-10-06 | `81f442a` | 验收：5 项检查可用；注入坏数据可正确报错退出 |
| T-018 | 后端 lint 清零 + golangci-lint v2 配置迁移 | 2026-10-06 | `ec2bed6` | 验收：`go build` + `golangci-lint run ./...` exit 0 |
| T-001 | 协作文档体系改造（AGENTS/ROADMAP/删除 ONBOARDING） | 2026-10-06 | `0237d40` | 验收：无 ONBOARDING 残留；md 链接锚点全通 |
| — | 2026-09 及之前的交付项 | — | 见 `git log` | 详细记录见 [PROJECT_STATUS.md](./PROJECT_STATUS.md#近期变更记录) |

> 归档规则：done 条目保留 3 个月后可移入上方汇总行；**ID 不复用**。备注只写验收结论（≤ 80 字），见 [DEVELOPMENT.md](./DEVELOPMENT.md) 文档回写文风。

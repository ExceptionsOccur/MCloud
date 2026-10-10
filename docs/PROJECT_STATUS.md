# 项目状态与技术债

> 本文档记录 MCloud 的**功能完成度、变更记录与已知问题**；任务队列与排期见 [ROADMAP.md](./ROADMAP.md)，执行协议见 [AGENTS.md](../AGENTS.md)。
>
> **维护规则**：功能合并进 `main` 时，由该 PR 作者当日更新「近期变更记录」与「功能完成度」；技术债发现即记，不集中补。**变更记录每条 ≤ 120 字一行摘要，禁止 ①②③ 枚举**（细则见 [DEVELOPMENT.md](./DEVELOPMENT.md) 文档回写文风）。文档入口见 [AGENTS.md](../AGENTS.md)。

## 功能完成度

> 计数（2026-10-10）：功能模块 **19** 项（**17** 完成 / **2** 待开发）；技术债 **7** 项。

| 模块 | 状态 | 说明 |
|------|------|------|
| 用户认证 | ✅ 完成 | JWT 登录/登出/改密码、失败 5 次锁定 15 分钟 |
| 主机资产管理 | ✅ 完成 | CRUD、批量增改、CSV 导入导出（多编码）、多条件筛选 |
| 云资源总览 | ✅ 完成 | 按区域维护总资源，设置菜单「云资源录入」手动录入 |
| 资源统计 | ✅ 完成 | 饼图展示 运行中/已停止/未使用 占比，表格含总量/已用/剩余/运行中/已停止 |
| IP 统计 | ✅ 完成 | /24 位图（10×26 坐标）、单点/全量探测、WebSocket 实时返回、网段在线增删改 |
| 业务统计 | ✅ 完成 | 项目/公司/人员三维聚合，柱状图 TOP10 + 点击下钻主机明细 |
| 人员管理 | ✅ 完成 | 人员录入与主机资产关联（`persons` 表 + `hosts.person_id`） |
| 零信任台账 | ✅ 完成 | `T-013`/`T-029`/`T-036`/`T-046` 申请单位/账户名/联系方式/公网IP入口(接入地区带出)/多组主机:端口配对(列表合并为「申请资源」标签列)/系统名称/时间/备注，主机被引用禁止删除 |
| 密码重置工具 | ✅ 完成 | `cmd/resetpw` 支持随机高熵密码或指定密码 |
| 容器化部署 | ✅ 完成 | Dockerfile（多阶段）+ docker-compose（dev/prod）+ Nginx 反代 |
| 协作文档体系 | ✅ 完成 | `T-001` AGENTS/ROADMAP 改造为 AI 协作协议（状态快照、任务队列、回写协议） |
| 数据导出报表 | ⏳ 待开发 | 预留 Excel/CSV 报表导出 |
| 域名台账 | ✅ 完成 | `T-034` 已升级为**端口映射台账** `/mapping-ledger`（公网IP↔内网多端口，域名可选）；`hosts.ip_mapped` 自动重算 |
| 公网 IP 关联 | ✅ 完成 | `T-015`/`T-034` `hosts.ip_mapped` 由映射台账自动重算；主机详情展示映射/零信任关联 |
| 自动化测试 | ⏳ 待开发 | 后端 Service/Controller、前端组件测试均未编写 |
| Goose 迁移执行器 | ✅ 完成 | `T-016` 启动时 goose 执行 `migrations/*.sql`（embed），AutoMigrate 兜底；存量 dev 库首跑+幂等已实测 |
| 公网 IP 资源录入 | ✅ 完成 | `T-031` `public_ips` 资源池（IP/运营商/备注），设置菜单 + `/public-ip` 独立页 |
| 数据备份 | ✅ 完成 | `T-038` 8 个业务 sheet 导出为单个 xlsx / 导入自然键 upsert（单事务失败整体回滚），`/data-backup` 页 + 设置菜单入口 |
| 审计日志 | ✅ 完成 | `T-047` `audit_logs` 全资源 CUD/导入审计（操作人快照/diff/request_id）+ `GET /api/audit-logs` + `/audit-logs` 页 |

## 近期变更记录

> 按日期倒序，保留最近 3 个月；每条注明提交号，完整历史用 `git log --oneline`。文风：每条 ≤ 120 字一行摘要，禁止枚举（见 [DEVELOPMENT.md](./DEVELOPMENT.md) 文档回写文风）。

### 2026-10-10

- `c1d4618` `refactor:` **T-004** 前端抽 composables（探测 WS / ECharts 注册 / 分页表），`stores/host` 组合改造对外 API 不变，IpStatistics 841→778 行
- `6b62e8d` `feat:` **T-050** `hosts.disk` 补录入口：表单磁盘输入框+校验，CSV 尾列/`xlsx hosts` sheet 补磁盘列（旧 26 列兼容）
- `a3a19ec` `feat:` **T-051** 公网IP引用保护（删/改IP 被映射或零信任引用 40901）+ `port_mappings` 整组唯一防重（迁移清重+唯一索引）
- `f325d4d` `refactor:` **T-049** 删 `port_mappings` 冗余列 `isp`/`exit_location`（派生字段恒从资源池带出，API 响应不变）
- `266bf66` `refactor:` **T-048** 迁移 SQL↔实库漂移对齐（主键 bigint、created_at 默认值、idx_ 唯一索引、FK NO ACTION；对齐迁移幂等实测）
- `76d4e8a` `feat:` **T-047** 全资源增删改审计日志（`audit_logs` 快照+diff+request_id，26 个 CUD 方法写入，`GET /api/audit-logs` + `/audit-logs` 页；表数 9→10、页面 9→10）
- `8a3940e` `fix:` **T-042** 批量编辑 `apply_time` null 与省略同义放行，非字符串/非法日期统一 40001（对齐单条口径；API.md 日期章节同步）
- `e039190` `refactor:` **T-002** 后端 `host_service.go` 拆分为 CRUD/批量/CSV 三文件（同包零签名变化，对外 API 行为不变；技术债「单文件过大」关闭）
- `f0643f8` `refactor:` **T-003** `csv.go` 业务逻辑下沉至 `host_csv_service.go`（controller 仅绑定+错误码映射，A/B 8 场景响应字节一致；技术债「controller 混入业务逻辑」关闭）
- `1f0b6a3` `refactor:` **T-010** 路由按域拆为 12 文件，JWT 收口受保护组（路由全表 A/B 67/67 一致，路径/方法/鉴权零变化）

### 2026-10-09

- `9c1cbfa` `feat:` **T-046** 零信任列表「申请主机/申请端口」两列合并为「申请资源」标签列，每组标签 `主机名(ip:port)` 按主机名排序（纯前端展示，API/DB 不变）
- `5fbb249` `docs:` **T-045** 协议补强：周期回顾偏差核查跳过纯文档类任务（AGENTS E.2 + DEVELOPMENT + ROADMAP 三处口径同步；纯文档任务核查文档同步属重复劳动）
- `f9314e0` `docs:` 全量回顾偏差修复（T-036~T-041 回写缺口与工作流/业务文档语义偏差 40+ 项；计数重算 16/2、技术债 10→9、锚点推进 T-041；入队 T-042~044）
- `4ccd241` `docs:` **T-041** 周期回顾协议：AGENTS 会话协议 E（触发/范围/处理）+ ROADMAP 回顾锚点 T-040 + 快照/DEVELOPMENT 同步；偏差先报告人类再修
- `9522789` `fix:` **T-040** 全仓日期时间格式统一（RFC3339 主+无时区兼容修 T-037；主机申请时间严格 `YYYY-MM-DD`；新增 `utils/datetime.go`；口径回写 API）
- `c196f37` `fix:` **T-039** 空库启动迁移失败修复（4 表内联 UNIQUE 转 GORM `idx_` 口径；修 050001 Down 块；索引/外键去重；补 ip_subnets 迁移）
- `6644d5a` `feat:` **T-038** 数据备份：8 个业务 sheet 单文件 xlsx 导出/导入（自然键 upsert 仅增改不删、单事务回滚逐行明细；`/data-backup` 页 + 8 视图设置菜单入口）

### 2026-10-08

- `ab7e3d0` `feat:` **T-036** 零信任台账多组主机:端口配对 + 公网IP入口带出接入地区（迁移 targets/public_ip；配对行表单；批量等长配对）
- `627ee8f` `fix:` **T-035** 映射公网IP改资源池选择（表单移除运营商/出口位置；列表从 public_ips 带出）
- `d26d6ad` `feat:` **T-034** 域名台账改造为端口映射台账（`port_mappings` 多端口等长校验；`ip_mapped` 自动重算；`/mapping-ledger`）
- `aac7fea` `feat:` **T-033** 域名/公网IP增加出口位置字段（`exit_location`；运营商之后；域名+公网IP+批量列同步）
- `4e96c09` `feat:` **T-032** 零信任/域名台账批量添加（文本CSV弹窗；主机按内网IP；域名重复跳过；`/batch` 接口）
- `e0d84ac` `feat:` **T-031** 公网 IP 资源录入（`public_ips` 表 + `/public-ip` 页 + 设置菜单；IP 唯一/格式校验；表数 8→9）
- `130388a` `feat:` **T-015** 公网 IP 字段改映射布尔 + 主机详情展示双台账关联（`ip_mapped`；CSV/表单/筛选同步；迁移 DROP public_ip）
- `e83db90` `feat:` **T-030** 域名台账字段改造（`isp`/`host_id`/`host_port`；移除 `provider`/`expires_at`；主机删除前校验域名引用）
- `1f70851` `feat:` **T-029** 零信任台账新增系统名称字段（`system_name` 选填；列表列/表单位于申请端口后；迁移 SQL + API/ARCHITECTURE 同步）
- `82e0e07` `feat:` **T-016** Goose 迁移执行器实装（`goose/v3` + embed 执行迁移 SQL，AutoMigrate 兜底；修复 cloud_resources 幂等；口径反转回写）
- `2e32a5b` `fix:` **T-028** 前端导航修正（设置菜单去掉台账入口；ZeroTrustLedger 补域名 Tab）
- `ebb52de` `fix:` **T-027** 前端 ESLint 警告清零（~724→0；`defineExpose` 修复改密弹窗；移除 IpStatistics console）
- `c5208c6` `docs:` **T-026** 协议补强：多任务会话串行合并（AGENTS A.2/A.4 + 分支协议 + DEVELOPMENT 协作流程）
- `4e0190c` `feat:` **T-014** 域名台账上线（models/services/controllers/routes + 迁移 SQL + 前端 `/domain-ledger` 页；域名唯一约束）
- `6a6e760` `feat:` **T-013** 零信任台账上线（models/services/controllers/routes + 迁移 + `/zero-trust` 页；引用主机校验；修前端 lint error）

### 2026-10-06

- `1837874` `docs:` **T-025** 协议补强：临时任务入队 A.0 + 分支步骤 + 红线 5 提交例外 + 提交号时序 + 会话接管 D
- `bcfb410` `docs:` **T-024** 文档回写文风上限（字段「文风」+ DEVELOPMENT 章节 + 压缩超长条目）
- `42d4c10` `docs:` **T-023** 文档卫生 + 移除 CI/CD 规划（T-007 入已取消、门禁改 agent 本地验证、完成度计数）
- `6187228` `docs:` **T-017** 文档事实修正与协议补强（快照协议、门禁前置、业务文档补缺）
- `81f442a` `feat:` **T-021** 文档一致性校验脚本 `check_docs.sh`（links/tree/snapshot/counts/shas）
- `ec2bed6` `chore:` **T-018** 后端 lint 清零 + golangci-lint v2 配置迁移

> 本日文档改造整体提交于 `0237d40`（T-001 协作文档体系 + 红线 7 按文件类别授权 + 入队 T-013/T-014/T-015；细分见该提交 message）。
- `docs:` 会话协议新增第 4 步「定位代码」（读 CODE_INDEX 补齐任务「定位」字段，补不出先声明不盲写）；ROADMAP 字段约定新增「定位」，14 条 todo 任务已回填涉及文件清单
- `docs:` 补齐 CODE_INDEX 缺失的**人员管理模块**（功能映射/`persons` 等共 8 处）；修正 README 页面数与 `/personnel` 入口
- `docs:` **统一迁移口径**：迁移 SQL 归档口径（不执行）确立、运行时 AutoMigrate 兜底——后经 T-016 反转为 goose 主路径（红线 9 / DEVELOPMENT / 注意事项 8 同步）
- `docs:` 排查确认 **goose 从未实装**（存量 `-- +goose` SQL 从未执行、`go.mod`/`go.sum` 无依赖、全仓库无调用）；按人类决定暂不实装，新增 `T-016` 入队并标注"口径反转回写"要求

### 2026-09-30

- `d385dc4` feat: 申请/技术信息 tab 顺序对调，手动输入人员提交时确保新增
- `55fba78` fix: IP 探测不再写入空记录，仅返回颜色状态
- `a8b2d23` feat: **人员录入与主机资产关联**（新增 `persons` 表、`hosts.person_id`、人员管理页）
- `8b0f18e` fix: SearchToolbar CPU 架构筛选补充 X86 选项
- `34a0aef` feat: 动态区域列表 + 资源统计排除裸金属 + dev 环境前后端分离
- `7b2ac44` fix: 资源统计全面排除裸金属服务器

### 2026-09（初始化）

- `903b521` init: MCloud 资产管理系统（主机管理 / 资源统计 / IP 统计 / 业务统计 / 容器化部署）
- `de40235` chore: dev docker-compose 前后端分离架构

<details>
<summary>更早的结构性变更（2026-09 之前，无提交号可考）</summary>

- 新增**业务统计**页面（项目/公司/人员三维聚合 + 柱状图下钻）
- **IP 统计**：网段从硬编码改为数据库管理（增删改、限 /24、自动规范化）；探测改为 WebSocket（并发 + 断线重连）
- **资源统计**：区分运行中/已停止；vCPU/内存/存储统计均排除裸金属；空记录与有数据记录分色
- **状态枚举**：新增「已关机」，统计中与「已停止」合并
- **CPU 架构**：新增 X86 选项（橙色标签）
- 新增**云资源录入**弹窗，入口在设置菜单
- 新增 `cmd/resetpw` 密码重置工具
- 默认管理员密码调整为 `Pass4MCloud`

</details>

## 已知问题 / 技术债

> 修复后请移除对应条目；排期中的技术债同步到 [ROADMAP.md](./ROADMAP.md) 的 `todo` 队列。

| 问题 | 说明 |
|------|------|
| 无自动化测试 | 现有测试约定为规划，尚无实际测试代码 |
| 前端大组件 | `IpStatistics.vue` 835 行等，待抽 composables |
| GORM 列名 | 缩写字段（如 `CIDR`）默认命名异常（`c_id_r`），须显式 `gorm:"column:xxx"` |
| 探测依赖 ping | 依赖宿主机 `ping` 命令；Linux `-W` 单位为秒，已改用 context 控制 100ms 超时 |
| 密码哈希强度 | SHA-256 + 盐，非慢哈希，生产环境建议升级 bcrypt |
| ServerPort 硬编码 | `config.go:43` 默认值 `"5677"` 违反红线 3（配置外置），待改环境变量，`T-020` 已入队 |
| WS 通道无 JWT 校验 | `/api/ws/probe` 仅校验 `token` 非空（`routes/websocket.go`），任意非空字符串即可建连并发起探测；`T-022` 已入队 |

## 注意事项

1. **密码格式**：SHA-256 + 盐，格式 `salt$hash`。默认管理员密码 `Pass4MCloud`（`database/postgres.go` 的 `seedAdmin()`）
2. **is_db_server 字段**：BOOLEAN 类型，API 响应统一返回 `true/false`
3. **CSV 编码**：导出需带 BOM（UTF-8 前缀 `\uFEFF`）以便 Excel 正确识别
4. **关联删除**：删除主机必须级联删除 `host_applications` 记录
5. **事务操作**：创建主机（hosts + host_applications）必须在同一事务完成
6. **ID 返回**：创建成功后返回新记录 ID（`data.id`）
7. **GORM 列名**：缩写字段必须显式指定 `gorm:"column:xxx"`
8. **数据库迁移（goose 主路径）**：启动时 `Migrate()` 先执行 goose（embed 的 `migrations/*.sql`），`AutoMigrate` 仅兜底；SQL 必须与模型定义一致；统一口径见 [DEVELOPMENT.md](./DEVELOPMENT.md)「数据库迁移规范」
9. **IP 网段**：不预置默认网段，由用户自行维护；`cidr` 仅接受 /24
10. **开发端口**：前端 5173（Vite）、后端 5677（容器）、生产 5678（Nginx+Go 单容器）

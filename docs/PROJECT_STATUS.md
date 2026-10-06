# 项目状态与技术债

> 本文档记录 MCloud 的**功能完成度、变更记录与已知问题**；任务队列与排期见 [ROADMAP.md](./ROADMAP.md)，执行协议见 [AGENTS.md](../AGENTS.md)。
>
> **维护规则**：功能合并进 `main` 时，由该 PR 作者当日更新「近期变更记录」与「功能完成度」；技术债发现即记，不集中补。文档入口见 [AGENTS.md](../AGENTS.md)。

## 功能完成度

| 模块 | 状态 | 说明 |
|------|------|------|
| 用户认证 | ✅ 完成 | JWT 登录/登出/改密码、失败 5 次锁定 15 分钟 |
| 主机资产管理 | ✅ 完成 | CRUD、批量增改、CSV 导入导出（多编码）、多条件筛选 |
| 云资源总览 | ✅ 完成 | 按区域维护总资源，设置菜单「云资源录入」手动录入 |
| 资源统计 | ✅ 完成 | 饼图展示 运行中/已停止/未使用 占比，表格含总量/已用/剩余/运行中/已停止 |
| IP 统计 | ✅ 完成 | /24 位图（10×26 坐标）、单点/全量探测、WebSocket 实时返回、网段在线增删改 |
| 业务统计 | ✅ 完成 | 项目/公司/人员三维聚合，柱状图 TOP10 + 点击下钻主机明细 |
| 人员管理 | ✅ 完成 | 人员录入与主机资产关联（`persons` 表 + `hosts.person_id`） |
| 密码重置工具 | ✅ 完成 | `cmd/resetpw` 支持随机高熵密码或指定密码 |
| 容器化部署 | ✅ 完成 | Dockerfile（多阶段）+ docker-compose（dev/prod）+ Nginx 反代 |
| 协作文档体系 | ✅ 完成 | `T-001` AGENTS/ROADMAP 改造为 AI 协作协议（状态快照、任务队列、回写协议） |
| 数据导出报表 | ⏳ 待开发 | 预留 Excel/CSV 报表导出 |
| 零信任台账 | ⏳ 待开发 | `T-013` 申请单位/账户名/联系方式/申请主机/端口/时间/密码(可选)/备注 |
| 域名台账 | ⏳ 待开发 | `T-014` 字段需求待细化 |
| 公网 IP 关联 | ⏳ 待开发 | `T-015` `hosts.public_ip` 改布尔「是否做了映射」，关联上述两台账 |
| 自动化测试 | ⏳ 待开发 | 后端 Service/Controller、前端组件测试均未编写 |
| CI 流水线 | ⏳ 待开发 | 尚无 `.github/workflows`，当前仅为本地 linter 自检 |
| Goose 迁移执行器 | ⏳ 待开发 | `T-016` **从未实装**（`go.mod` 无依赖、无调用）；现状 SQL 仅归档，运行时靠 AutoMigrate |

## 近期变更记录

> 按日期倒序，保留最近 3 个月；每条注明提交号，完整历史用 `git log --oneline`。

### 2026-10-06

> 本日文档改造整体提交于 `0237d40`（下述各条为该提交内的细分变更）。

- `docs:` **T-001 协作文档体系改造**：AGENTS.md 重写为 AI 协作入口（状态快照 + 会话/回写协议 + 术语表 + 验证命令）；`ROADMAP.md` 改造为任务队列（T-xxx、验收标准、状态流转，唯一事实来源）；删除 `docs/ONBOARDING.md` 内容并入 AGENTS.md；分支策略对齐为 `main + feature`；修正 README 端口、ARCHITECTURE 目录树、DEVELOPMENT 中的 `develop` 残留
- `docs:` 红线 7 改为**按文件类别的修改权限**（状态同步类必须回写 / 协议条文默认禁止、仅人类发起的 `docs:` 任务可改 / 构建文件始终禁止），消除与回写协议的冲突；DEVELOPMENT.md 核心原则第 5 条同步
- `docs:` 新增待开发任务 `T-013` 零信任台账（字段已确认）、`T-014` 域名台账（字段待细化）、`T-015` `hosts.public_ip` 改布尔「是否做了映射」+ 双台账关联；均为 P2 暂缓，依赖与待确认项见 ROADMAP
- `docs:` 会话协议新增第 4 步「定位代码」（读 CODE_INDEX 补齐任务「定位」字段，补不出先声明不盲写）；ROADMAP 字段约定新增「定位」，14 条 todo 任务已回填涉及文件清单
- `docs:` 补齐 CODE_INDEX 缺失的**人员管理模块**（models/controllers/services/api/views/功能映射/`persons` 表 共 8 处）；修正 README「4 个功能页面」为 5 个并补 `/personnel` 与设置菜单入口
- `docs:` **统一迁移口径**（消除红线 9 与运行时的矛盾）：迁移 SQL = 归档/评审要求（不被执行），运行时由启动时 `AutoMigrate` 兜底——DEVELOPMENT.md 迁移规范加口径表、AGENTS 红线 9 改写、PROJECT_STATUS 注意事项 8 与 ARCHITECTURE 目录树同步（代码事实：`go.mod` 无 goose，`Migrate()` 仅 AutoMigrate+seedAdmin）
- `docs:` 排查确认 **goose 从未实装**（8 个 `-- +goose` SQL 从未执行、`go.mod`/`go.sum` 无依赖、全仓库无调用）；按人类决定暂不实装，新增 `T-016` 入队并标注"口径反转回写"要求

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
| 无 CI 流水线 | 文档承诺 PR 门禁，实际 `.github/` 不存在，目前仅本地 linter |
| 无自动化测试 | 现有测试约定为规划，尚无实际测试代码 |
| 后端单文件过大 | `services/host_service.go` 22KB 混杂 CRUD/批量/CSV，待拆分 |
| controller 混入业务逻辑 | `controllers/csv.go` 含行校验/去重，违反红线 1，待下沉 |
| 前端大组件 | `IpStatistics.vue` 746 行等，待抽 composables |
| 数据库端口映射 | 数据库容器若未映射 5432 端口，后端会连不上（需 `-p 5432:5432`） |
| GORM 列名 | 缩写字段（如 `CIDR`）默认命名异常（`c_id_r`），须显式 `gorm:"column:xxx"` |
| 探测依赖 ping | 依赖宿主机 `ping` 命令；Linux `-W` 单位为秒，已改用 context 控制 100ms 超时 |
| 密码哈希强度 | SHA-256 + 盐，非慢哈希，生产环境建议升级 bcrypt |

## 注意事项

1. **密码格式**：SHA-256 + 盐，格式 `salt$hash`。默认管理员密码 `Pass4MCloud`（`database/postgres.go` 的 `seedAdmin()`）
2. **is_db_server 字段**：BOOLEAN 类型，API 响应统一返回 `true/false`
3. **CSV 编码**：导出需带 BOM（UTF-8 前缀 `\uFEFF`）以便 Excel 正确识别
4. **关联删除**：删除主机必须级联删除 `host_applications` 记录
5. **事务操作**：创建主机（hosts + host_applications）必须在同一事务完成
6. **ID 返回**：创建成功后返回新记录 ID（`data.id`）
7. **GORM 列名**：缩写字段必须显式指定 `gorm:"column:xxx"`
8. **数据库迁移（双轨口径）**：迁移 SQL 是红线 9 的**归档要求**（必须随模型变更提交，不被执行）；运行时结构由启动时 `AutoMigrate` 兜底应用，`migrations/*.sql` 不参与执行；统一口径见 [DEVELOPMENT.md](./DEVELOPMENT.md)「数据库迁移规范」
9. **IP 网段**：不预置默认网段，由用户自行维护；`cidr` 仅接受 /24
10. **开发端口**：前端 5173（Vite）、后端 5677（容器）、生产 5678（Nginx+Go 单容器）

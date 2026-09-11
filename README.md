# GoCron SQLite

基于 [ouqiang/gocron](https://github.com/ouqiang/gocron) 改造的定时任务管理系统，使用 Go 开发，通过 Web 页面集中管理任务、执行节点和运行日志。

本项目将数据库改为 **SQLite**，保留原有管理页面的布局、操作流程和任务功能，并升级前端框架、认证方式及关键 Go 依赖。部署时不需要安装 MySQL 或 PostgreSQL，也不需要配置数据库服务、端口和账号。

## 主要功能

- **任务管理**：支持精确到秒的 cron 表达式、手动执行、启用、停用和删除任务。
- **HTTP 任务**：支持 GET、POST 请求，由调度器直接执行，无需额外执行节点。
- **Shell 任务**：通过执行节点运行命令，支持配置多个节点。
- **执行控制**：支持超时、失败重试、单实例运行和父子任务依赖。
- **日志管理**：查看执行状态、输出、耗时及登录记录。
- **用户权限**：支持管理员和普通用户，通过账号密码登录。
- **结果通知**：保留邮件、Slack、Webhook，以及通知模板和关键字匹配功能。

## 相比原项目的修改

### 1. 数据库存储改为 SQLite

- 移除应用对 MySQL、PostgreSQL 驱动和数据库服务的依赖，默认使用本地 SQLite 文件。
- 首次安装自动创建数据库目录、数据库文件、数据表和索引，无需提前建库或执行 SQL。
- 安装页面去掉数据库主机、端口、用户名和密码等配置，只需创建管理员。
- 启用 WAL 日志模式、10 秒锁等待和 FULL 同步，并使用单连接串行访问数据库，减少并发写入时的锁竞争。
- 建表、默认通知配置和管理员创建放在同一事务中；主机关联替换、通知配置更新也使用事务，并返回写入错误。

### 2. 升级前端，保留管理页面与操作流程

- Vue 2 升级为 **Vue 3**，Element UI 替换为 **Element Plus**。
- Vue Router、Vuex 升级到适配 Vue 3 的版本，更新 Axios 等依赖。
- 使用 **Vite** 替换旧 webpack/Babel 构建链，并通过 `package-lock.json` 固定构建依赖。
- 适配路由、插槽、弹窗、时间格式化和通知设置等页面逻辑，保留原有功能入口。
- 管理页面资源嵌入 Go 程序，运行时无需单独部署前端服务。

### 3. 调整账号登录与密码存储

- 支持普通账号密码登录，**邮箱为选填项**，多个账号可以不填写邮箱。
- 设置了邮箱的用户仍可使用邮箱登录；登录时优先匹配账号。
- 密码仅要求非空，不设置最小或最大长度，也不会自动删除密码首尾的空格。
- 新密码使用带随机盐的 **Argon2id** 散列存储，替换原有 MD5 方案。
- 保留旧 MD5 密码的校验兼容逻辑，旧账号成功登录后自动升级为 Argon2id。
- 使用 **JWT v5**，校验签名算法、签发者和过期时间；请求认证时重新读取账号状态和权限，使禁用、删除及权限修改生效。

### 4. 更新 Go 依赖与修复运行问题

- 升级 gRPC、SQLite 驱动、xorm、Macaron、`golang.org/x/crypto` 等关键依赖。
- 单实例任务通过原子操作获取执行权，修复并发触发时可能重复执行的问题。
- 将任务执行中的 panic 明确记录为失败，避免异常被当作成功。
- 修复 Shell 任务超时取消时的空指针风险和 goroutine 无法退出的问题，终止子进程组并等待回收。
- 修正任务计数初始化顺序、主机端口校验，以及数据库会话和静态资源文件的关闭。

## SQLite 为什么更方便

| 操作 | 本项目的处理方式 |
| --- | --- |
| 安装数据库 | 无需额外数据库服务，SQLite 已集成到程序中 |
| 配置数据库连接 | 无需主机、端口、数据库账号或密码 |
| 创建数据库与表 | 首次在安装页面提交管理员信息时自动完成 |
| 保存任务与账号 | 持久化到本地 `data/gocron.db` 文件 |
| 指定存储位置 | 可设置数据根目录，也可修改数据库文件路径 |
| 备份与迁移 | 正常停止程序后复制数据目录，无需维护独立数据库服务 |

SQLite 适合单机调度场景。一个调度器仍可管理多个远程执行节点；数据库保存在调度器所在机器上。

## 快速开始

### 从源码构建

构建后端需要 Go 1.26+、C 编译器和 CGO。SQLite 驱动会随程序一起编译，运行时不需要安装数据库服务。

```bash
git clone https://github.com/cocoajin/gocron.git
cd gocron
CGO_ENABLED=1 go build -o bin/gocron ./cmd/gocron
CGO_ENABLED=1 go build -o bin/gocron-node ./cmd/node
```

仓库包含已嵌入的管理页面资源，直接构建 Go 程序即可使用。只有修改前端页面时，才需要重新构建前端，步骤见下文。

### 启动与初始化

启动调度器：

```bash
./bin/gocron web --host 127.0.0.1 -p 5920
```

浏览器访问 [http://127.0.0.1:5920](http://127.0.0.1:5920)，在首次安装页面设置管理员账号和密码。邮箱可留空，数据库会自动初始化，没有预设管理员账号。

需要运行 Shell 任务时，在另一个终端启动执行节点：

```bash
./bin/gocron-node -s 127.0.0.1:5921
```

然后在「任务节点」中添加 `127.0.0.1`，端口填写 `5921`，创建 Shell 任务时选择该节点。HTTP 任务无需启动执行节点。

以上示例仅监听本机地址。如需远程访问，请按实际部署环境设置监听地址；调度器与远程执行节点之间支持 TLS 配置。

## SQLite 配置与数据目录

### 默认配置

首次安装后，程序自动生成 `conf/app.ini`。与 SQLite 相关的核心配置如下，**无需手动填写**：

```ini
[default]
db.engine = sqlite3
db.database = data/gocron.db
```

这是配置文件中的数据库片段，完整文件还包含应用名称、认证密钥及其他设置；修改数据库路径时请保留其余配置。

默认的数据根目录为可执行文件所在目录；当可执行文件位于 `bin/` 下时，使用其上一级目录。上述源码构建方式对应项目根目录：

```text
conf/
  app.ini          # 应用配置与认证密钥
  install.lock     # 已完成安装的标记
  .version         # 数据版本标记
data/
  gocron.db        # SQLite 数据库
```

### 自定义存储位置

通过 `GOCRON_DATA_DIR` 指定配置和数据库的数据根目录：

```bash
GOCRON_DATA_DIR=/srv/gocron ./bin/gocron web --host 127.0.0.1 -p 5920
```

首次安装后，配置保存在 `/srv/gocron/conf/app.ini`，数据库保存在 `/srv/gocron/data/gocron.db`。目录必须允许运行程序的用户写入，后续启动应使用同一数据根目录。

如需单独调整数据库位置，可在停止程序后修改 `conf/app.ini` 中的 `db.database`：

```ini
[default]
db.engine = sqlite3
db.database = /srv/gocron-storage/tasks.db
```

相对路径以数据根目录为基准，绝对路径直接使用指定位置。修改配置不会自动搬迁已有数据库；已有安装需要先将数据库迁移到目标位置，再启动程序。

### 备份与恢复

1. 正常停止调度器，等待正在执行的任务结束。
2. 备份数据根目录中的 `conf/` 和 `data/`；如果自定义了数据库绝对路径，也要备份对应文件。
3. 恢复时复制配置与数据库到目标位置，确认文件权限和 `db.database` 配置后启动。

WAL 模式下，数据库旁可能存在 `gocron.db-wal` 和 `gocron.db-shm`。运行时只复制 `.db` 文件可能遗漏尚未合并的数据；需要在线备份时，应使用 SQLite 备份 API。

## 前端开发

需要 Node.js 22.12+ 和 npm。

```bash
cd web/vue
npm ci --ignore-scripts
npm run dev
```

开发服务将 `/api` 请求代理到 `127.0.0.1:5920`，需同时启动调度器。

修改前端后，在项目根目录重新生成内嵌资源并构建：

```bash
cd web/vue
npm run build
cd ../..
go run github.com/rakyll/statik -src=web/vue/dist -dest=internal -f
gofmt -w internal/statik/statik.go
CGO_ENABLED=1 go build -o bin/gocron ./cmd/gocron
```

## 使用范围

- 当前面向 SQLite 单机部署，不包含旧 MySQL、PostgreSQL 数据的自动导入工具。密码兼容逻辑不等同于跨数据库迁移。
- 不应让多个调度器通过网络共享盘共同使用同一个 SQLite 数据库文件。
- 邮件、Slack 和 Webhook 通知需要自行配置发送服务或接收地址。

## 版本与自动发布

当前版本：**2.0.0**。版本记录在根目录 `VERSION`，Git 标签使用 `2.0.0` 格式，不加 `v` 前缀。

[GitHub Actions 发布流程](.github/workflows/release.yml) 在推送版本标签后自动构建并发布到 [Releases](https://github.com/cocoajin/gocron/releases)：

- Linux amd64、Linux arm64：`.tar.gz` 压缩包。
- Windows amd64：`.zip` 压缩包。
- 每个包包含调度器、执行节点、README、许可证及版本号，附带统一的 `SHA256SUMS` 校验文件。

流程先构建并嵌入管理页面，再启用 CGO 编译 SQLite 驱动。Linux 和 Windows 构建静态链接包；全部平台构建成功后才发布 Release。重复运行同一标签的流程会更新对应 Release 的附件。

发布新版本时，同步更新 `VERSION`、两个命令入口的 `AppVersion` 及前端包版本，提交后创建并推送同名标签。可在 `docs/releases/版本号.md` 中编写发布说明；没有该文件时自动生成说明。

```bash
git tag -a 2.0.0 -m "Release 2.0.0"
git push origin master
git push origin 2.0.0
```

工作流使用 GitHub 自动提供的 `GITHUB_TOKEN`，发布作业已配置 `contents: write`，无需额外保存个人访问令牌。仓库需要启用 Actions，并允许工作流写入 Release。手动重跑时选择版本标签，不能选择分支。

## 来源与协议

本项目基于 [ouqiang/gocron](https://github.com/ouqiang/gocron)，遵循 [MIT License](LICENSE)。问题和建议可提交至本仓库的 [Issues](https://github.com/cocoajin/gocron/issues)。

---

## 原项目介绍

### gocron — 定时任务管理系统

原项目 [ouqiang/gocron](https://github.com/ouqiang/gocron) 是使用 Go 语言开发的轻量级定时任务集中调度和管理系统，用于替代 Linux crontab，通过 Web 界面统一配置任务、管理执行节点并查看执行结果。

原项目文档：[Wiki](https://github.com/ouqiang/gocron/wiki)。原有的延时任务功能已拆分为独立项目：[延迟队列](https://github.com/ouqiang/delay-queue)。

### 原项目功能特性

- Web 界面管理定时任务，支持精确到秒的 crontab 时间表达式。
- 任务执行失败可重试，任务超时可强制结束。
- 支持任务依赖配置，在 A 任务完成后执行 B 任务。
- 支持多用户及账户权限控制。
- Shell 任务在执行节点上运行，支持同一任务在多个节点执行。
- HTTP 任务由调度器直接访问指定 URL，不依赖执行节点。
- 查看任务执行结果日志。
- 支持邮件、Slack、Webhook 执行结果通知。

### 原项目截图

以下截图来自原项目，用于展示原有调度流程和管理界面。

**调度流程**

![原项目调度流程](assets/screenshot/scheduler.png)

**任务管理**

![原项目任务管理](assets/screenshot/task.png)

**通知配置**

![原项目通知配置](assets/screenshot/notification.png)

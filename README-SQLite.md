# GoCron SQLite 本机测试版

基于 https://github.com/ouqiang/gocron 。新代码位于 `goCorn`，保留原管理页面布局和任务功能，前端升级 Vue 3 + Element Plus。首次安装页改为 SQLite，只需创建管理员，无需 MySQL、PostgreSQL、Node.js 或 Go 运行环境。

## macOS 使用

1. 解压 `dist/gocron-sqlite-darwin-arm64.zip`，双击 `GoCron.app`。
2. 程序打开终端并启动调度器和本地执行节点，随后打开 http://127.0.0.1:5920 。
3. 首次使用在安装页设置管理员账号和密码，邮箱选填；密码仅要求非空，不设置最小或最大长度，没有预设密码。
4. 创建 Shell 任务前，在主机管理添加 `127.0.0.1`，端口 `5921`。HTTP 任务不需要节点。
5. 按终端里的 Control+C 停止，等待正在执行的任务结束后退出；下次双击继续使用。

本包为 Apple Silicon arm64 本机测试包，采用本地 ad-hoc 签名，未做 Apple 开发者公证。若拷贝到其他 Mac 被系统拦截，可在系统设置的「隐私与安全性」中允许打开。

默认仅监听本机 127.0.0.1；5920 是管理端口，5921 是 Shell 执行节点端口。端口占用时启动器会报错，不会终止其他程序。

## 数据与备份

默认数据根目录：`~/Library/Application Support/gocron-sqlite/`

- 数据库：`data/gocron.db`，运行时还会出现 `gocron.db-wal` 和 `gocron.db-shm`。
- 配置和安装标记：`conf/app.ini`、`conf/install.lock`。
- 日志：`log/cron.log`、`log/web.log`、`log/node.log`。

更新应用包不会覆盖数据。备份时先 Control+C 正常停止，再复制整个数据目录。不要在运行时只复制 `.db` 文件，以免遗漏 WAL 中尚未合并的数据。也可使用 SQLite 官方备份 API。

可用 `GOCRON_DATA_DIR` 环境变量指定独立数据目录。直接运行二进制时，默认仍使用可执行文件所在目录保存数据；发布启动器会设置上述用户数据目录。

## 本次修复

- 移除 MySQL/PostgreSQL 驱动和连接配置；默认 SQLite，自动创建目录与文件。
- SQLite 启用 WAL、10 秒 busy timeout、FULL 同步及单连接写入，避免定时任务并发写日志造成锁冲突。
- 安装中的建表、索引、通知默认配置与管理员写入使用事务；重复安装请求串行处理。
- 主机关联替换和通知配置更新使用事务；返回数据库写入错误。
- 单实例任务以原子操作获取执行权，修复先检查后写入导致的重复执行。
- 任务 panic 明确记录为失败，修复异常被当作成功的情况。
- 修复 Shell 超时取消时进程未创建的空指针风险、未释放的 goroutine，终止整个子进程组并等待回收。
- 修正任务计数初始化顺序、主机端口验证、数据库会话和静态资源文件关闭。
- 数据及配置采用较严格文件权限，提供 macOS 启动器、可重复构建脚本和测试。

## 构建与测试

需要 macOS、Go 1.26+、Xcode Command Line Tools（SQLite CGO 编译），重新构建前端还需 Node.js 22.12+ / npm。

```bash
./scripts/package-macos.sh
go test -race ./... -timeout 60s
python3 scripts/smoke_test.py
```

前端使用 Vue 3.5.42、Element Plus 2.14.5、Vue Router 4、Vuex 4、Axios 1.20.0 和 Vite 8.3.0；`package-lock.json` 固定依赖。已移除 Vue 2、Element UI、旧 webpack/Babel 构建链，无需 OpenSSL legacy 选项。发布程序不依赖构建工具，也无需联网加载管理页面。

## 范围与限制

本次面向全新 SQLite 本机安装，不包含旧 MySQL 数据自动迁移。SQLite 适合单机调度，不支持多个调度器共享网络盘上的同一数据库。现有远程节点、HTTP/Shell 任务、依赖任务、通知和权限界面保留。

本次升级了 JWT v5、gRPC 1.83.2、xorm 1.4.1、SQLite 驱动 1.14.52、Go x/crypto 等依赖；新密码采用 Argon2id 随机盐散列，旧 MD5 密码仅用于兼容读取，旧用户成功登录后自动转换。账号优先匹配，设置了邮箱的用户仍可用邮箱登录。空邮箱可以被多个账号使用。

本版按用户要求仅完成前后端编译和打包，不启动程序进行功能测试，也不预创建用户或数据库。早期 SQLite 版本的运行测试结果不代表本次大版本升级已完成运行验证。保留回归测试与 smoke_test.py，供后续需要时手动运行。

邮件、Slack、Webhook 的外部发送需要你自己的配置。尚未进行全量安全审计；部分无新版替代的历史组件暂时保留。默认仅开放本机端口。

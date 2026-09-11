package install

import (
	"fmt"
	"sync"

	macaron "gopkg.in/macaron.v1"

	"github.com/go-macaron/binding"
	"github.com/ouqiang/gocron/internal/models"
	"github.com/ouqiang/gocron/internal/modules/app"
	"github.com/ouqiang/gocron/internal/modules/setting"
	"github.com/ouqiang/gocron/internal/modules/utils"
	"github.com/ouqiang/gocron/internal/service"
)

// 系统安装

type InstallForm struct {
	DbType               string `binding:"In(sqlite,sqlite3)"`
	AdminUsername        string `binding:"Required;MaxSize(32)"`
	AdminPassword        string `binding:"Required"`
	ConfirmAdminPassword string `binding:"Required"`
	AdminEmail           string `binding:"MaxSize(50)"`
}

func (f InstallForm) Error(ctx *macaron.Context, errs binding.Errors) {
	if len(errs) == 0 {
		return
	}
	json := utils.JsonResponse{}
	content := json.CommonFailure("表单验证失败, 请检测输入")
	ctx.Write([]byte(content))
}

var installMu sync.Mutex

// 安装
func Store(ctx *macaron.Context, form InstallForm) string {
	installMu.Lock()
	defer installMu.Unlock()
	json := utils.JsonResponse{}
	if app.Installed {
		return json.CommonFailure("系统已安装!")
	}
	if form.AdminPassword != form.ConfirmAdminPassword {
		return json.CommonFailure("两次输入密码不匹配")
	}
	err := testDbConnection(form)
	if err != nil {
		return json.CommonFailure(err.Error())
	}
	// 写入数据库配置
	err = writeConfig(form)
	if err != nil {
		return json.CommonFailure("数据库配置写入文件失败", err)
	}

	appConfig, err := setting.Read(app.AppConfig)
	if err != nil {
		return json.CommonFailure("读取应用配置失败", err)
	}
	app.Setting = appConfig

	models.Db = models.CreateDb()
	defer func() {
		if !app.Installed {
			models.Db.Close()
		}
	}()
	// 创建数据库表
	migration := new(models.Migration)
	err = migration.InstallAdmin(&models.User{Name: form.AdminUsername, Password: form.AdminPassword, Email: form.AdminEmail, IsAdmin: 1})
	if err != nil {
		return json.CommonFailure(fmt.Sprintf("创建数据库表失败-%s", err.Error()), err)
	}

	// 创建安装锁
	err = app.CreateInstallLock()
	if err != nil {
		return json.CommonFailure("创建文件安装锁失败", err)
	}

	// 更新版本号文件
	app.UpdateVersionFile()

	app.Installed = true
	// 初始化定时任务
	service.ServiceTask.Initialize()

	return json.Success("安装成功", nil)
}

// 配置写入文件
func writeConfig(form InstallForm) error {
	dbConfig := []string{
		"db.engine", "sqlite3",
		"db.database", "data/gocron.db",
		"db.max.idle.conns", "1",
		"db.max.open.conns", "1",
		"allow_ips", "",
		"app.name", "定时任务管理系统", // 应用名称
		"api.key", "",
		"api.secret", "",
		"enable_tls", "false",
		"concurrency.queue", "500",
		"auth_secret", utils.RandAuthToken(),
		"ca_file", "",
		"cert_file", "",
		"key_file", "",
	}

	return setting.Write(dbConfig, app.AppConfig)
}

// 测试数据库连接
func testDbConnection(form InstallForm) error {
	var s setting.Setting
	s.Db.Engine = "sqlite3"
	s.Db.Database = "data/gocron.db"
	db, err := models.CreateTmpDb(&s)
	if err != nil {
		return err
	}
	return db.Close()
}

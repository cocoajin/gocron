package models

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	macaron "gopkg.in/macaron.v1"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ouqiang/gocron/internal/modules/app"
	"github.com/ouqiang/gocron/internal/modules/logger"
	"github.com/ouqiang/gocron/internal/modules/setting"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

type Status int8
type CommonMap map[string]interface{}

var TablePrefix = ""
var Db *xorm.Engine

const (
	Disabled Status = 0 // 禁用
	Failure  Status = 0 // 失败
	Enabled  Status = 1 // 启用
	Running  Status = 1 // 运行中
	Finish   Status = 2 // 完成
	Cancel   Status = 3 // 取消
)

const (
	Page        = 1    // 当前页数
	PageSize    = 20   // 每页多少条数据
	MaxPageSize = 1000 // 每次最多取多少条
)

const DefaultTimeFormat = "2006-01-02 15:04:05"

type BaseModel struct {
	Page     int `xorm:"-"`
	PageSize int `xorm:"-"`
}

func (model *BaseModel) parsePageAndPageSize(params CommonMap) {
	page, ok := params["Page"]
	if ok {
		model.Page = page.(int)
	}
	pageSize, ok := params["PageSize"]
	if ok {
		model.PageSize = pageSize.(int)
	}
	if model.Page <= 0 {
		model.Page = Page
	}
	if model.PageSize <= 0 {
		model.PageSize = MaxPageSize
	}
}

func (model *BaseModel) pageLimitOffset() int {
	return (model.Page - 1) * model.PageSize
}

// CreateDb opens the local database, failing before the scheduler starts on errors.
func CreateDb() *xorm.Engine {
	engine, err := CreateTmpDb(app.Setting)
	if err != nil {
		logger.Fatal("打开 SQLite 数据库失败", err)
	}
	TablePrefix = app.Setting.Db.Prefix
	if macaron.Env == macaron.DEV {
		engine.ShowSQL(true)
		engine.Logger().SetLevel(log.LOG_DEBUG)
	}
	return engine
}

// CreateTmpDb uses a single connection to serialize writes. WAL and busy_timeout
// also allow SQLite readers and backup tools to access the file safely.
func CreateTmpDb(s *setting.Setting) (*xorm.Engine, error) {
	if s.Db.Engine != "sqlite3" && s.Db.Engine != "sqlite" {
		return nil, fmt.Errorf("仅支持 SQLite，请使用 db.engine=sqlite3")
	}
	if strings.ContainsAny(s.Db.Prefix, "`\" ;.-/\\") {
		return nil, fmt.Errorf("表前缀无效")
	}
	path := s.Db.Database
	if path == "" {
		path = "data/gocron.db"
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(app.AppDir, path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: path}
	q := url.Values{"_busy_timeout": {"10000"}, "_journal_mode": {"WAL"}, "_synchronous": {"FULL"}, "_foreign_keys": {"on"}, "_txlock": {"immediate"}}
	u.RawQuery = q.Encode()
	engine, err := xorm.NewEngine("sqlite3", u.String())
	if err != nil {
		return nil, err
	}
	engine.SetMaxOpenConns(1)
	engine.SetMaxIdleConns(1)
	if s.Db.Prefix != "" {
		engine.SetTableMapper(names.NewPrefixMapper(names.SnakeMapper{}, s.Db.Prefix))
	}
	if err = engine.Ping(); err != nil {
		engine.Close()
		return nil, err
	}
	return engine, nil
}

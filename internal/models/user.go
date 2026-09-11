package models

import (
	"crypto/subtle"
	"strings"
	"time"

	"github.com/ouqiang/gocron/internal/modules/utils"
)

const PasswordSaltLength = 6

// 用户model
type User struct {
	Id        int       `json:"id" xorm:"pk autoincr notnull "`
	Name      string    `json:"name" xorm:"varchar(32) notnull unique"`       // 用户名
	Password  string    `json:"-" xorm:"varchar(255) notnull "`               // 密码
	Salt      string    `json:"-" xorm:"char(6) notnull "`                    // 密码盐值
	Email     string    `json:"email" xorm:"varchar(50) notnull default '' "` // 邮箱
	Created   time.Time `json:"created" xorm:"datetime notnull created"`
	Updated   time.Time `json:"updated" xorm:"datetime updated"`
	IsAdmin   int8      `json:"is_admin" xorm:"tinyint notnull default 0"` // 是否是管理员 1:管理员 0:普通用户
	Status    Status    `json:"status" xorm:"tinyint notnull default 1"`   // 1: 正常 0:禁用
	BaseModel `json:"-" xorm:"-"`
}

// 新增
func (user *User) Create() (insertId int, err error) {
	user.Status = Enabled
	user.Salt = ""
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return 0, err
	}

	_, err = Db.Insert(user)
	if err == nil {
		insertId = user.Id
	}

	return
}

// 更新
func (user *User) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(user).ID(id).Update(data)
}

func (user *User) UpdatePassword(id int, password string) (int64, error) {
	safePassword, err := hashPassword(password)
	if err != nil {
		return 0, err
	}
	return user.Update(id, CommonMap{"password": safePassword, "salt": ""})
}

// 删除
func (user *User) Delete(id int) (int64, error) {
	return Db.ID(id).Delete(user)
}

// 禁用
func (user *User) Disable(id int) (int64, error) {
	return user.Update(id, CommonMap{"status": Disabled})
}

// 激活
func (user *User) Enable(id int) (int64, error) {
	return user.Update(id, CommonMap{"status": Enabled})
}

// 验证用户名和密码
func (user *User) Match(username, password string) bool {
	found, err := Db.Where("name = ? AND status = ?", username, Enabled).Get(user)
	if err != nil {
		return false
	}
	if !found && username != "" {
		found, err = Db.Where("email = ? AND status = ?", username, Enabled).Get(user)
	}
	if err != nil || !found {
		return false
	}
	if strings.HasPrefix(user.Password, "$argon2id$") {
		return verifyPassword(password, user.Password)
	}
	// A valid old MD5 login transparently upgrades the stored hash.
	legacy := utils.Md5(password + user.Salt)
	if len(user.Password) != 32 || subtle.ConstantTimeCompare([]byte(legacy), []byte(user.Password)) != 1 {
		return false
	}
	hash, err := hashPassword(password)
	if err != nil {
		return false
	}
	changed, err := Db.Table(user).Where("id = ? AND password = ?", user.Id, user.Password).Update(CommonMap{"password": hash, "salt": ""})
	if err != nil || changed != 1 {
		return false
	}
	user.Password, user.Salt = hash, ""
	return true
}

// 获取用户详情
func (user *User) Find(id int) error {
	_, err := Db.ID(id).Get(user)

	return err
}

// 用户名是否存在
func (user *User) UsernameExists(username string, uid int) (int64, error) {
	if uid > 0 {
		return Db.Where("name = ? AND id != ?", username, uid).Count(user)
	}

	return Db.Where("name = ?", username).Count(user)
}

// 邮箱地址是否存在
func (user *User) EmailExists(email string, uid int) (int64, error) {
	if email == "" {
		return 0, nil
	}
	if uid > 0 {
		return Db.Where("email = ? AND id != ?", email, uid).Count(user)
	}

	return Db.Where("email = ?", email).Count(user)
}

func (user *User) List(params CommonMap) ([]User, error) {
	user.parsePageAndPageSize(params)
	list := make([]User, 0)
	err := Db.Desc("id").Limit(user.PageSize, user.pageLimitOffset()).Find(&list)

	return list, err
}

func (user *User) Total() (int64, error) {
	return Db.Count(user)
}

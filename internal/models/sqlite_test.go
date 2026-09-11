package models

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ouqiang/gocron/internal/modules/setting"
)

func openTestDB(t *testing.T) *setting.Setting {
	t.Helper()
	s := &setting.Setting{}
	s.Db.Engine = "sqlite3"
	s.Db.Database = filepath.Join(t.TempDir(), "nested & 中文", "gocron.db")
	var err error
	Db, err = CreateTmpDb(s)
	if err != nil {
		t.Fatal(err)
	}
	TablePrefix = ""
	t.Cleanup(func() { Db.Close() })
	return s
}

func TestSQLiteLifecycle(t *testing.T) {
	s := openTestDB(t)
	if err := new(Migration).InstallAdmin(&User{Name: "admin", Password: "test123456", Email: "admin@example.test", IsAdmin: 1}); err != nil {
		t.Fatal(err)
	}
	if !new(User).Match("admin", "test123456") {
		t.Fatal("login failed")
	}
	if new(User).Match("admin", "wrong") {
		t.Fatal("wrong password accepted")
	}
	h := &Host{Name: "127.0.0.1", Alias: "local", Port: 5921}
	if _, err := h.Create(); err != nil {
		t.Fatal(err)
	}
	task := &Task{Name: "sqlite test", Spec: "0 * * * * *", Command: "echo ok", Protocol: TaskRPC, Level: TaskLevelParent}
	if _, err := task.Create(); err != nil {
		t.Fatal(err)
	}
	th := new(TaskHost)
	if err := th.Add(task.Id, []int{int(h.Id)}); err != nil {
		t.Fatal(err)
	}
	if _, err := task.Enable(task.Id); err != nil {
		t.Fatal(err)
	}
	list, err := task.ActiveList(1, 20)
	if err != nil || len(list) != 1 || len(list[0].Hosts) != 1 {
		t.Fatalf("active list: %+v %v", list, err)
	}
	if total, err := task.Total(CommonMap{}); err != nil || total != 1 {
		t.Fatalf("total %d %v", total, err)
	}
	if err := th.Add(task.Id, nil); err != nil {
		t.Fatal(err)
	}
	if hosts, err := th.GetHostIdsByTaskId(task.Id); err != nil || len(hosts) != 0 {
		t.Fatalf("empty hosts: %v %v", hosts, err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l := &TaskLog{TaskId: task.Id, Name: fmt.Sprint(i), Protocol: TaskRPC}
			id, err := l.Create()
			if err == nil {
				_, err = l.Update(id, CommonMap{"status": Finish, "result": "ok"})
			}
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := Db.Close(); err != nil {
		t.Fatal(err)
	}
	Db, err = CreateTmpDb(s)
	if err != nil {
		t.Fatal(err)
	}
	if !new(User).Match("admin", "test123456") {
		t.Fatal("user not persisted")
	}
	logs, err := new(TaskLog).List(CommonMap{})
	if err != nil || len(logs) != 40 {
		t.Fatalf("logs %d %v", len(logs), err)
	}
	for _, l := range logs {
		if l.Status != Finish || l.Result != "ok" {
			t.Fatalf("bad log %+v", l)
		}
	}
	if _, err := task.Delete(task.Id); err != nil {
		t.Fatal(err)
	}
	if list, err := task.List(CommonMap{}); err != nil || len(list) != 0 {
		t.Fatalf("soft deletion: %v %v", list, err)
	}
	if _, err := new(TaskLog).Clear(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLiteInstallRollback(t *testing.T) {
	openTestDB(t)
	if err := Db.Sync2(&TaskLog{}); err != nil {
		t.Fatal(err)
	}
	if err := new(Migration).Install("gocron"); err == nil {
		t.Fatal("must reject preexisting table")
	}
	for _, bean := range []interface{}{&User{}, &Task{}} {
		exists, err := Db.IsTableExist(bean)
		if err != nil || exists {
			t.Fatalf("partially created schema survived: %T %v", bean, err)
		}
	}
}

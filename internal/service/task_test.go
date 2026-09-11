package service

import (
	"github.com/ouqiang/gocron/internal/models"
	"github.com/ouqiang/gocron/internal/modules/logger"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSingleInstanceAtomicClaim(t *testing.T) {
	var instance Instance
	var winners int32
	var wg sync.WaitGroup
	for j := 0; j < 100; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if instance.tryStart(42) {
				atomic.AddInt32(&winners, 1)
			}
		}()
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("%d simultaneous winners", winners)
	}
	instance.done(42)
	if !instance.tryStart(42) {
		t.Fatal("claim not released")
	}
}

type panicHandler struct{}

func (panicHandler) Run(models.Task, int64) (string, error) { panic("test panic") }
func TestPanicIsFailure(t *testing.T) {
	logger.InitLogger()
	result := execJob(panicHandler{}, models.Task{}, 1)
	if result.Err == nil {
		t.Fatal("panic reported as success")
	}
}

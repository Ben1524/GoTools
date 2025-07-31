package timewheel

import (
	"sync"
	"testing"
	"time"
)

// 测试单个任务正常执行
func TestSingleTask(t *testing.T) {
	// 创建时间轮：1秒间隔，60个槽
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Fatalf("创建时间轮失败: %v", err)
	}
	defer tw.Stop()

	var (
		execCount int32
		wg        sync.WaitGroup
	)
	wg.Add(1) // 等待1次执行

	// 添加任务：延迟1秒，执行1次
	taskID := "single-task"
	tw.AddTask(taskID, time.Second, func(key string) {
		defer wg.Done()
		execCount++
		if key != taskID {
			t.Errorf("任务ID不匹配，预期 %s, 实际 %s", taskID, key)
		}
	}, 1)

	// 等待任务执行或超时（3秒超时，允许系统误差）
	timeout := time.After(3 * time.Second)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if execCount != 1 {
			t.Errorf("任务执行次数错误，预期1次，实际%d次", execCount)
		}
	case <-timeout:
		t.Fatal("任务执行超时，未在预期时间内执行")
	}
}

// 测试重复任务（有限次数）
func TestRepeatTask(t *testing.T) {
	// 创建时间轮：1秒间隔，60个槽
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Fatalf("创建时间轮失败: %v", err)
	}
	defer tw.Stop()

	var (
		execCount int32
		mu        sync.Mutex
		wg        sync.WaitGroup
	)
	repeatTimes := int64(3)
	wg.Add(int(repeatTimes)) // 等待3次执行

	// 添加任务：延迟1秒，执行3次
	taskID := "repeat-task"
	tw.AddTask(taskID, time.Second, func(key string) {
		defer wg.Done()
		mu.Lock()
		execCount++
		mu.Unlock()
	}, repeatTimes)

	// 等待所有执行完成或超时（5秒超时）
	timeout := time.After(5 * time.Second)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		if execCount != int32(repeatTimes) {
			t.Errorf("重复任务执行次数错误，预期%d次，实际%d次", repeatTimes, execCount)
		}
	case <-timeout:
		t.Fatal("重复任务执行超时")
	}
}

// 测试无限重复任务（times=-1）
func TestInfiniteTask(t *testing.T) {
	// 创建时间轮：1秒间隔，60个槽
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Fatalf("创建时间轮失败: %v", err)
	}
	defer tw.Stop()

	var (
		execCount int32
		mu        sync.Mutex
		wg        sync.WaitGroup
	)
	expectedExec := 2 // 预期执行2次后停止时间轮
	wg.Add(expectedExec)

	// 添加任务：延迟1秒，无限执行
	taskID := "infinite-task"
	tw.AddTask(taskID, time.Second, func(key string) {
		mu.Lock()
		execCount++
		current := execCount
		mu.Unlock()

		if current <= int32(expectedExec) {
			wg.Done()
		}
	}, -1)

	// 等待2次执行后停止时间轮
	timeout := time.After(4 * time.Second)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		tw.Stop() // 执行2次后停止
		close(done)
	}()

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		if execCount != int32(expectedExec) {
			t.Errorf("无限任务执行次数错误，预期%d次，实际%d次", expectedExec, execCount)
		}
	case <-timeout:
		t.Fatal("无限任务执行超时")
	}
}

// 测试任务删除（确保删除后不执行）
func TestRemoveTask(t *testing.T) {
	// 创建时间轮：2秒间隔，60个槽（延迟更长，确保有时间删除）
	tw, err := NewTimeWheel(2*time.Second, 60)
	if err != nil {
		t.Fatalf("创建时间轮失败: %v", err)
	}
	defer tw.Stop()

	var execCount int32

	// 添加任务：延迟4秒（确保有时间删除）
	taskID := "remove-task"
	tw.AddTask(taskID, 4*time.Second, func(key string) {
		execCount++ // 如果执行则计数
	}, 1)

	// 立即删除任务
	time.Sleep(1 * time.Second) // 等待任务被添加到时间轮
	err = tw.RemoveTask(taskID)
	if err != nil {
		t.Fatalf("删除任务失败: %v", err)
	}

	// 等待足够长的时间（超过延迟时间），检查是否执行
	time.Sleep(5 * time.Second)

	if execCount != 0 {
		t.Errorf("已删除的任务不应执行，实际执行了%d次", execCount)
	}
}

// 测试长延迟任务（超过单圈最大时间）
func TestLongDelayTask(t *testing.T) {
	// 时间轮配置：1秒间隔，3个槽（单圈最大延迟3秒）
	tw, err := NewTimeWheel(time.Second, 3)
	if err != nil {
		t.Fatalf("创建时间轮失败: %v", err)
	}
	defer tw.Stop()

	var (
		execCount int32
		wg        sync.WaitGroup
	)
	wg.Add(1)

	// 延迟4秒（超过单圈3秒，需要多转1圈）
	taskID := "long-delay-task"
	tw.AddTask(taskID, 4*time.Second, func(key string) {
		defer wg.Done()
		execCount++
	}, 1)

	// 等待执行或超时（6秒超时）
	timeout := time.After(6 * time.Second)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if execCount != 1 {
			t.Errorf("长延迟任务执行次数错误，预期1次，实际%d次", execCount)
		}
	case <-timeout:
		t.Fatal("长延迟任务执行超时")
	}
}

// 测试时间轮停止后任务不再执行
func TestStopTimeWheel(t *testing.T) {
	tw, err := NewTimeWheel(time.Second, 60)
	if err != nil {
		t.Fatalf("创建时间轮失败: %v", err)
	}

	var execCount int32

	// 添加任务：延迟2秒执行
	taskID := "stop-test-task"
	tw.AddTask(taskID, 2*time.Second, func(key string) {
		execCount++
	}, 1)

	// 1秒后停止时间轮（在任务执行前）
	time.Sleep(1 * time.Second)
	tw.Stop()

	// 再等待2秒，确认任务不执行
	time.Sleep(2 * time.Second)

	if execCount != 0 {
		t.Errorf("时间轮停止后任务不应执行，实际执行了%d次", execCount)
	}
}

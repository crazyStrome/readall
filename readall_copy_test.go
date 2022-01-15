package readall

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testName = "test.data.rar"

func TestGrow(t *testing.T) {
	var buf bytes.Buffer
	t.Logf("before len:%v, cap:%v", buf.Len(), buf.Cap())
	buf.Grow(bytes.MinRead)
	t.Logf("after len:%v, cap:%v", buf.Len(), buf.Cap())
}
func BenchmarkReadAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readAllData(&testing.T{}, testName)
	}
}
func TestIOCopy(t *testing.T) {
	file, err := os.Open(testName)
	if err != nil {
		t.Errorf("open err:%v", err)
		return
	}
	data := make([]byte, 0, 74077894*2)
	buf := bytes.NewBuffer(data)
	_, err = io.Copy(buf, file)
	if err != nil {
		t.Errorf("readall err:%v", err)
		return
	}
}
func TestReadAll(t *testing.T) {
	file, err := os.Open(testName)
	if err != nil {
		t.Errorf("open err:%v", err)
		return
	}
	_, err = ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("readall err:%v", err)
		return
	}

}
func TestHttpGet(t *testing.T) {
	rsp, err := http.Get("http://xxx.com")
	if err != nil {
		t.Errorf("get err:%v", err)
		return
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	t.Logf("body len:%v, read err:%v", len(body), err)
}
func TestReadAllIOCopy(t *testing.T) {
	for i := 0; i < 100; i++ {
		readmax, readtotal := readAllData(t, testName)
		copymax, copytotal := iocopyData(t, testName)
		t.Logf("Max copy/read:%v, total copy/read:%v",
			float64(copymax)/float64(readmax), float64(copytotal)/float64(readtotal))
	}
}
func readAllData(t *testing.T, fileName string) (int64, int64) {
	mu := &sync.Mutex{}
	var max int64
	var total int64
	ctrl := make(chan struct{}, 10)
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		ctrl <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-ctrl
				wg.Done()
			}()
			start := time.Now()
			file, err := os.Open(fileName)
			if err != nil {
				t.Errorf("open err:%v", err)
				return
			}
			_, err = ioutil.ReadAll(file)
			if err != nil {
				t.Errorf("readall err:%v", err)
				return
			}
			cost := time.Since(start).Milliseconds()
			atomic.AddInt64(&total, cost)
			mu.Lock()
			if cost > max {
				max = cost
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return max, total
}

func iocopyData(t *testing.T, fileName string) (int64, int64) {
	mu := &sync.Mutex{}
	var max int64
	var total int64
	wg := &sync.WaitGroup{}
	ctrl := make(chan struct{}, 10)
	for i := 0; i < 100; i++ {
		ctrl <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-ctrl
				wg.Done()
			}()
			start := time.Now()
			file, err := os.Open(fileName)
			if err != nil {
				t.Errorf("open err:%v", err)
				return
			}
			fileInfo, er := os.Stat(fileName)
			if er != nil {
				t.Errorf("state err:%v", err)
				return
			}
			data := make([]byte, 0, fileInfo.Size()*2)
			buf := bytes.NewBuffer(data)
			_, err = io.Copy(buf, file)
			if err != nil {
				t.Errorf("copy err:%v", err)
				return
			}
			cost := time.Since(start).Milliseconds()
			atomic.AddInt64(&total, cost)
			mu.Lock()
			if cost > max {
				max = cost
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return max, total
}
func BenchmarkFib10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fib(10)
	}
}

// 斐波那契数列
func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

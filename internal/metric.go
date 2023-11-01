package internal

import (
	"bytes"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	metric *Metric
	once   sync.Once
)

type Metric struct {
	statsdM      sync.RWMutex
	ready        bool
	clientPool   []*Client
	localBufPool []*bytes.Buffer
	wg           sync.WaitGroup
	local        bool
}

// GetSingleInst 获取单例
func GetSingleInst() *Metric {
	if metric == nil {
		once.Do(func() {
			metric = new(Metric)
		})
	}

	return metric
}

// ReCreate 初始化
func (metric *Metric) ReCreate(prefix string, addresses []string) error {
	return metric.reCreate(prefix, addresses, false)
}

// Init 初始化
func (metric *Metric) Init(prefix string, addresses []string) error {
	return metric.reCreate(prefix, addresses, false)
}

// Stop 暂停
func (metric *Metric) Stop() {
	defer noPanic()

	if metric == nil {
		return
	}

	// 先标记
	metric.statsdM.Lock()
	defer metric.statsdM.Unlock()
	if metric.ready == false {
		return
	}

	metric.ready = false

	// 等全部业务走完
	metric.wg.Wait()

	// 刷新缓存
	for _, client := range metric.clientPool {
		if client != nil {
			_ = client.Flush()
			_ = client.Close()
		}
	}

	return
}

// reCreate 重建
func (metric *Metric) reCreate(prefix string, addresses []string, local bool) error {
	defer noPanic()

	if metric == nil {
		return nil
	}

	// 先标记
	metric.statsdM.Lock()
	defer metric.statsdM.Unlock()
	metric.ready = false
	// 是否只是本地连接
	metric.local = local

	// 等全部业务走完
	metric.wg.Wait()

	// 刷新缓存
	for _, client := range metric.clientPool {
		if client != nil {
			_ = client.Flush()
			if !metric.local {
				_ = client.Close()
			}
		}
	}

	// 释放
	metric.localBufPool = nil
	metric.clientPool = nil
	if len(addresses) == 0 {
		return nil
	}

	// 重建资源
	for _, address := range addresses {
		var client *Client
		var err error
		if local {
			localBuf := new(bytes.Buffer)
			metric.localBufPool = append(metric.localBufPool, localBuf)
			client = NewClient(localBuf)
		} else {
			client, err = DialTimeout(address, 5*time.Second)
			if err != nil {
				return err
			}
		}

		if client == nil {
			continue
		}

		client.Prefix(prefix)
		metric.clientPool = append(metric.clientPool, client)
	}

	// 成功
	metric.ready = true

	return nil
}

// increment 计数增加
func (metric *Metric) increment(name string, count int, rate float64, needFlush bool) (string, error) {
	defer noPanic()

	if metric == nil {
		return "", nil
	}

	// 判断是否要停止
	metric.statsdM.RLock()
	defer metric.statsdM.RUnlock()
	if metric.ready == false {
		return "", nil
	}

	// 记录
	metric.wg.Add(1)
	defer metric.wg.Done()

	index := rand.Intn(len(metric.clientPool))
	client := metric.clientPool[index]
	if client == nil {
		return "", nil
	}

	err := client.Increment(name, count, rate)
	if err != nil {
		return "", err
	}

	// 本地接口
	if metric.local {
		needFlush = true
	}

	if needFlush {
		err = client.Flush()
		if err != nil {
			return "", err
		}
	}

	// 本地接口
	if metric.local {
		ret := metric.localBufPool[index].String()
		metric.localBufPool[index].Reset()
		return ret, nil
	}

	return "", nil
}

// Increment 计数增加
func (metric *Metric) Increment(name string, count int, rate float64, needFlush bool) error {
	_, err := metric.increment(name, count, rate, needFlush)
	return err
}

// Gauge 设置计数
func (metric *Metric) Gauge(name string, count int, needFlush bool) error {
	defer noPanic()

	if metric == nil {
		return nil
	}

	// 判断是否要停止
	metric.statsdM.RLock()
	defer metric.statsdM.RUnlock()
	if metric.ready == false {
		return nil
	}

	// 记录
	metric.wg.Add(1)
	defer metric.wg.Done()

	index := rand.Intn(len(metric.clientPool))
	client := metric.clientPool[index]
	if client == nil {
		return nil
	}

	err := client.Gauge(name, count)
	if err != nil {
		return err
	}

	if needFlush {
		err = client.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

// Histogram 直方图
func (metric *Metric) Histogram(name string, count int, needFlush bool) error {
	defer noPanic()

	if metric == nil {
		return nil
	}

	// 判断是否要停止
	metric.statsdM.RLock()
	defer metric.statsdM.RUnlock()
	if metric.ready == false {
		return nil
	}

	// 记录
	metric.wg.Add(1)
	defer metric.wg.Done()

	index := rand.Intn(len(metric.clientPool))
	client := metric.clientPool[index]
	if client == nil {
		return nil
	}

	err := client.Histogram(name, count)
	if err != nil {
		return err
	}

	if needFlush {
		err = client.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

func noPanic() {
	r := recover()
	if r != nil {
		log.Println(r)
	}
}

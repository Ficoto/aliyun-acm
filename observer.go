package aliacm

import (
	"sync"
)

// AfterUpdateHook 配置更新完毕后的回调函数
type AfterUpdateHook func(Config)

// ObserverRegisterInfo 用来添加观察者需要关心的配置
type ObserverRegisterInfo struct {
	AfterUpdateHook AfterUpdateHook
	Info            Info
}

// Observer observes the config change.
type Observer struct {
	notifyMap        map[Info]AfterUpdateHook
	isFirstUpdateMap map[Info]bool
	updateLockMap    map[Info]sync.RWMutex
	readLockMap      map[Info]sync.WaitGroup
}

// GenerateObserver 用来创建新的Observer
func GenerateObserver(oriList ...ObserverRegisterInfo) *Observer {
	var obs Observer
	obs.isFirstUpdateMap = make(map[Info]bool)
	obs.readLockMap = make(map[Info]sync.WaitGroup)
	obs.notifyMap = make(map[Info]AfterUpdateHook)
	obs.updateLockMap = make(map[Info]sync.RWMutex)
	for _, ori := range oriList {
		obs.isFirstUpdateMap[ori.Info] = true
		obs.readLockMap[ori.Info] = sync.WaitGroup{}
		obs.readLockMap[ori.Info].Add(1)
		obs.notifyMap[ori.Info] = ori.AfterUpdateHook
		obs.updateLockMap[ori.Info] = sync.RWMutex{}
	}
	return &obs
}

// Infos 获取Observer所有的Info
func (o *Observer) Infos() []Info {
	var infos []Info
	for info := range o.notifyMap {
		infos = append(infos, info)
	}
	return infos
}

// OnUpdate ACM配置更新后的回调函数
func (o *Observer) OnUpdate(config Config) {
	if _, ok := o.updateLockMap[config.Info]; !ok {
		return
	}

	o.updateLockMap[config.Info].Lock()
	defer o.updateLockMap[config.Info].Unlock()

	if o.isFirstUpdateMap[config.Info] {
		o.isFirstUpdateMap[config.Info] = false
	} else {
		o.readLockMap[config.Info].Add(1)
	}
	defer o.readLockMap[config.Info].Done()

	o.notifyMap[config.Info](config)
}

// Wait 用于等待所有AfterUpdateHook调用完
func (o *Observer) Wait() {
	for _, lock := range o.readLockMap {
		lock.Wait()
	}
}

// SingleWait 用于等待单个AfterUpdateHook调用完
func (o *Observer) SingleWait(info Info) {
	if _, ok := o.readLockMap[info]; !ok {
		return
	}
	o.readLockMap[info].Wait()
}

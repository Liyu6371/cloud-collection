package vmware

import (
	"cloud-collection/common"
	"cloud-collection/logger"
	"context"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

var (
	wg   sync.WaitGroup
	Name = "VMWare"
)

// NewVMWareTask 实例化 VMCollector
func NewVMWareTask(tasks map[string]interface{}) *VMCollector {
	if len(tasks) == 0 {
		return nil
	}
	var vm VMCollector
	if err := mapstructure.Decode(tasks, &vm); err != nil {
		logger.Errorf("unable to decode tasks %+v into VMCollector, error:%s\n", tasks, err)
		return nil
	}
	return &vm
}

// Run VMCollectorTask 入口，由上层 CollectorService 调用
func (v *VMCollector) Run(ctx context.Context) {
	defer logger.Infoln("VMCollectionTask shutting down.")

	// 判断是否有需要进行采集的云实例
	if len(*v.Clouds) == 0 {
		logger.Infoln("VMCollectionTask: No cloud found, Nothing to do!")
		return
	}

	// 判断需要采集的云实例与并发数量
	if len(*v.Clouds) > v.Concurrency {
		logger.Warnln("VMCollectionTask: Number of cloud foundry is greater than concurrency.")
		logger.Warnln("VMCollectionTask: The rest of cloud foundry will be ignored.")
	}

	context, cancel := context.WithCancel(ctx)
	defer cancel()
	count := 0
	for _, cloud := range *v.Clouds {
		count += 1
		if count > v.Concurrency {
			logger.Warnln("VMCollectionTask: The number of cloud foundry reach the top concurrency.")
			break
		} else {
			wg.Add(1)
			go process(cloud, context)
		}
	}
	<-ctx.Done()
	logger.Infoln("VMCollectionTask: catch ctx.Done signal, waiting for all goroutines to finish.")
	cancel()
	wg.Wait()
	logger.Infoln("VMCollector Run: All goroutines finished.")
}

func process(c Cloud, ctx context.Context) {
	defer func() {
		wg.Done()
	}()

	period := common.DefaultVMPeriod
	if c.Period != "" {
		if p, err := time.ParseDuration(c.Period); err == nil {
			period = p
		}
	}

	// 周期调度
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			NewVMCollectorTask(c, ctx).process()
		}
	}
}

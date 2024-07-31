package collector

import (
	"cloud-collection/collector/vmware"
	"cloud-collection/config"
	"cloud-collection/logger"
	"context"
	"sync"
)

var Name = "CollectorService"

type CloudCollectionTask interface {
	Run(ctx context.Context)
}

// Collector 结构体
type CloudTaskController struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	cloudTaskConf config.CloudCollectTask
}

func NewCollectorService(c config.CloudCollectTask) *CloudTaskController {
	return &CloudTaskController{
		cloudTaskConf: c,
		wg:            sync.WaitGroup{},
	}
}

func (c *CloudTaskController) StartService(ctx context.Context) {
	defer func() {
		logger.Infoln("CollectorService stopped")
	}()
	c.ctx, c.cancel = context.WithCancel(ctx)
	defer c.cancel()
	// 启动 VM 处理
	c.runCloudCollector(vmware.NewVMWareTask(c.cloudTaskConf.VMWare), vmware.Name)
	<-ctx.Done()
	c.cancel()
	c.wg.Wait()
}

func (c *CloudTaskController) runCloudCollector(task CloudCollectionTask, name string) {
	if task == nil {
		logger.Errorf("CloudCollectionTask is nil,  task name is: %s\n", name)
		return
	}
	c.wg.Add(1)
	go func(t CloudCollectionTask) {
		defer c.wg.Done()
		t.Run(c.ctx)
	}(task)
	logger.Infof("collectionTask :%s is running", name)
}

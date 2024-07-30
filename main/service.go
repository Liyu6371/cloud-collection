package main

import (
	"cloud-collection/collector"
	"cloud-collection/config"
	"cloud-collection/logger"
	"cloud-collection/socket"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Service interface {
	StartService(ctx context.Context)
}

// CloudServiceController 控制 Server 的启动
type CloudServiceController struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg               sync.WaitGroup
	socketService    *socket.GseSocket              // socket 服务
	collectorService *collector.CloudTaskController // 云采集控制器
}

// NewCloudServiceController 创建 ServiceController
func NewCloudServiceController(c config.Config) *CloudServiceController {
	ctx, cancel := context.WithCancel(context.Background())
	controller := &CloudServiceController{
		ctx:              ctx,
		cancel:           cancel,
		wg:               sync.WaitGroup{},
		socketService:    socket.NewGseSocketService(c.Socket),
		collectorService: collector.NewCollectorService(c.CloudCollectTask),
	}

	return controller
}

func (svc *CloudServiceController) Run() {
	// 非检测模式状态
	if !*testFlag {
		// 启动 GSEService
		svc.runService(svc.socketService, socket.Name)
	}
	// 启动云监控采集服务
	svc.runService(svc.collectorService, collector.Name)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svc.cancel()
	svc.wg.Wait()
	logger.Infoln("Cloud Collection Service Exit. Bye Bye~")
}

func (svc *CloudServiceController) runService(s Service, name string) {
	if s == nil {
		logger.Errorf("service is nil, process exit, service name is: %s\n", name)
		os.Exit(1)
	}
	svc.wg.Add(1)
	go func(service Service) {
		defer svc.wg.Done()
		service.StartService(svc.ctx)
	}(s)
	logger.Infof("service %s is running\n", name)
}

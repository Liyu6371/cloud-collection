package socket

import (
	"cloud-collection/common"
	"cloud-collection/config"
	"cloud-collection/logger"
	"context"
	"os"
	"sync"

	"github.com/TencentBlueKing/bkmonitor-datalink/pkg/libgse/gse"
)

var (
	defaultWorkerNum   = 1
	defaultGseClient   *gse.GseSimpleClient
	defaultQueueBuffer = 500
	Name               = "GSEService"
	GlobalMsgCh        chan gse.GseMsg
)

type GseSocket struct {
	c      config.Socket
	ch     chan gse.GseMsg
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewGseSocketService(c config.Socket) *GseSocket {
	defaultGseClient = gse.NewGseSimpleClient()
	if c.SocketPath != "" {
		defaultGseClient.SetAgentHost(c.SocketPath)
	} else {
		defaultGseClient.SetAgentHost(common.DefaultSocketPath)
		logger.Warn("use default socket path:", common.DefaultSocketPath)
	}
	// 根据配置进行消息队列缓冲区的设置
	if c.QueueBuffer != 0 {
		GlobalMsgCh = make(chan gse.GseMsg, c.QueueBuffer)
	} else {
		GlobalMsgCh = make(chan gse.GseMsg, defaultQueueBuffer)
	}

	return &GseSocket{
		c:  c,
		ch: GlobalMsgCh,
		wg: sync.WaitGroup{},
	}
}

func (g *GseSocket) StartService(c context.Context) {
	ctx, cancel := context.WithCancel(c)
	g.ctx = ctx
	g.cancel = cancel
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.startGseServer()
	}()
	// 上层传递的 ctx 信息
	<-c.Done()
	g.cancel()
	g.wg.Wait()
}

// startGseServer 启动GSE服务，并且监听 MsgQueue 等待数据上报
func (g *GseSocket) startGseServer() {
	err := defaultGseClient.Start()
	if err != nil {
		logger.Errorf("start gse server error: %v\n", err)
		os.Exit(1)
	}
	logger.Debugln("start gse server success!")

	if g.c.Worker != 0 {
		defaultWorkerNum = g.c.Worker
	}

	for i := 0; i < defaultWorkerNum; i++ {
		g.wg.Add(1)
		go g.monitorMsg(i)
	}
}

func (g *GseSocket) monitorMsg(workerId int) {
	for {
		select {
		case msg := <-g.ch:
			if err := defaultGseClient.Send(msg); err != nil {
				logger.Errorf("send msg error: %v\n", err)
			}
			logger.Debugf("send msg success: %v\n", msg)
		// 监听父进程（主进程的中断信号）
		case <-g.ctx.Done():
			g.wg.Done()
			logger.Infof("gse goroutine:%d exit", workerId)
			return
		}
	}
}

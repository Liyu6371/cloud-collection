package vmware

import (
	"cloud-collection/common"
	"cloud-collection/logger"
	"cloud-collection/socket"
	"context"
	"net/url"
	"sync"

	"github.com/TencentBlueKing/bkmonitor-datalink/pkg/libgse/gse"
	"github.com/vmware/govmomi"
)

type VMCollectorTask struct {
	c      Cloud
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	client *govmomi.Client
}

func NewVMCollectorTask(c Cloud, ctx context.Context) *VMCollectorTask {
	v := &VMCollectorTask{
		c:  c,
		wg: sync.WaitGroup{},
	}
	v.ctx, v.cancel = context.WithCancel(ctx)
	return v
}

// getClient 获取 VMClient
func (v *VMCollectorTask) getClient() *govmomi.Client {
	if v.client != nil {
		return v.client
	}
	// 开始创先 VMClient
	password := common.DecryptPassword(v.c.Password, common.VMAesKey)
	u := &url.URL{
		Scheme: "https",
		Host:   v.c.Server,
		Path:   "/sdk",
	}
	u.User = url.UserPassword(v.c.Account, password)
	client, err := govmomi.NewClient(v.ctx, u, true)
	if err != nil {
		logger.Errorf("unable to create vmware client: %v\n", err)
		return nil
	}
	v.client = client
	return client
}

func (v *VMCollectorTask) hostInstanceConfig() []string {
	if v.c.Host.HostInstances != nil {
		return *v.c.Host.HostInstances
	}
	return nil
}

func (v *VMCollectorTask) hostMetricsConfig() []Metrics {
	if v.c.Host.HostMetrics != nil {
		return *v.c.Host.HostMetrics
	}
	return nil
}

// SendMsg 发送数据到 GSESocket 消费
func (v *VMCollectorTask) sendMsg(dataid int32, data []byte, des string, dataType string) {
	msg := gse.NewGseCommonMsg(data, dataid, 0, 0, 0)
	socket.GlobalMsgCh <- msg
}

// process 拉起 VMCollectorTask 中的所有采集项目
func (v *VMCollectorTask) process() {

	// 并发的启动采集任务
	if len(*v.c.Host.HostInstances) != 0 {
		v.wg.Add(2)
		go v.collectHost()
		go v.collectEventByTime(v.c.Host.HostEventDataId, "host", *v.c.Host.HostInstances, "主机", v.c.Id, v.c.Period)
	}

}

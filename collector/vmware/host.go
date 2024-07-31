// host 收集 VMWare Cloud 中的 host 实例指标
package vmware

import (
	"cloud-collection/common"
	"cloud-collection/logger"
	"encoding/json"
	"strconv"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// collectHost 收集 host 指标
func (v *VMCollectorTask) collectHost() {
	client := v.getClient()
	if client == nil {
		logger.Errorln("unable to get client!")
		return
	}

	kind := []string{"HostSystem"}
	manager := view.NewManager(client.Client)
	view, err := manager.CreateContainerView(v.ctx, client.ServiceContent.RootFolder, kind, true)
	if err != nil {
		logger.Errorf("unable to create container view: %v\n", err)
		return
	}
	// 退出前关闭 ContainerView
	defer func() {
		if err := view.Destroy(v.ctx); err != nil {
			logger.Errorf("unable to destroy container view: %v\n", err)
			return
		}
	}()
	// 获取 所有的 host 主机列表
	var hosts []mo.HostSystem
	if err := view.Retrieve(v.ctx, kind, []string{"summary", "datastore"}, &hosts); err != nil {
		logger.Errorf("unable to retrieve hosts, error:%s\n", err)
		return
	}
	// 判定一下 hosts 是否为空的情况
	if len(hosts) == 0 {
		logger.Warnln("no hosts found, nothing to collect")
		return
	}
	// 用于计算后续的 host 主机内存
	hostMemMap := make(map[string]int64)
	// 筛选出 hosts 里面需要采集的 host 主机用于后续在 reference manager 进行采集使用
	var referenceHosts []types.ManagedObjectReference
	for _, host := range hosts {
		reference := host.Reference()
		if common.IsContain(v.hostInstanceConfig(), reference.Value) {
			hostMemMap[reference.Value] = host.Summary.Hardware.MemorySize
			referenceHosts = append(referenceHosts, reference)
		}
	}

	// perfMetrics 用于存储可以直接使用 performance Manager 获取的指标
	var perfMetrics []string
	// 标志位，判定是否需要采集磁盘信息
	DiskFlag := false
	// 开始处理配置里面需要采集的指标
	cMetrics := v.hostMetricsConfig()
	for _, metric := range cMetrics {
		if metric.Alias == common.DiskUsedAvg {
			DiskFlag = true
		} else if metric.Alias == common.MemTotalCapacityAverag {
			// 源代码中是这样判定 memTotalCapacityAverag
			// 不明白原因是为什么， 如果不想采集 直接不下发这个配置不就好了 ╮(╯▽╰)╭
			continue
		} else {
			perfMetrics = append(perfMetrics, metric.Alias)
		}
	}
	// 获取性能计数器
	perfManager := performance.NewManager(client.Client)
	spec := types.PerfQuerySpec{
		MaxSample:  1,
		MetricId:   []types.PerfMetricId{{Instance: "*"}},
		IntervalId: 20,
	}
	sample, err := perfManager.SampleByName(v.ctx, spec, perfMetrics, referenceHosts)
	if err != nil {
		logger.Errorf("unable to sample perf metrics: %v\n", err)
		return
	}
	metricSeries, err := perfManager.ToMetricSeries(v.ctx, sample)
	if err != nil {
		logger.Errorf("unable to convert perf metrics: %v\n", err)
		return
	}
	counters, err := perfManager.CounterInfoByName(v.ctx)
	if err != nil {
		logger.Errorf("unable to get counters: %v\n", err)
		return
	}

	hostSpaceMap := make(map[string]float64)
	if DiskFlag {
		storeSpaceMap := v.getStoreInfo()
		for _, host := range hosts {
			var sizeTotal int64
			var sizeUsed int64
			for _, s := range host.Datastore {
				sizeTotal += storeSpaceMap[s.Value].Capacity
				sizeUsed += storeSpaceMap[s.Value].Used
			}
			if sizeTotal == 0 {
				hostSpaceMap[host.Reference().Value] = 0
			} else {
				hostSpaceMap[host.Reference().Value] = common.UnitConversion((float64(sizeUsed) / float64(sizeTotal)) * 100)
			}
		}
	}

	var rd []MetricsData
	timestamp := time.Now().Unix()
	for _, item := range metricSeries {
		target := strconv.Itoa(v.c.Id) + "@" + item.Entity.Value
		dimension := Dimension{v.c.Id, item.Entity.Value, "host", ""}

		metrics := make(map[string]float64)
		// 对获取到的 metric 进行单位的转化
		for _, v := range item.Value {
			counter := counters[v.Name]
			units := counter.UnitInfo.GetElementDescription().Label

			if n := common.TransformMetricAlias(v.Name); n != "" {
				metrics[n] = common.ConvertMetricValue(v.ValueCSV(), units)
			} else {
				logger.Errorf("something went wrong, got an empty metric name, dropped.\n")
			}
		}
		// 开始计算 memTotalmbAverage 指标
		if v, ok := hostMemMap[item.Entity.Value]; ok {
			metrics[common.MemTotalMBAverage] = common.UnitConversion(float64(v) / (1024 * 1024 * 1024))
		}

		if DiskFlag {
			if v := common.TransformMetricAlias(common.DiskUsedAvg); v != "" {
				if dv, ok := hostSpaceMap[item.Entity.Value]; ok {
					metrics[v] = dv
				} else {
					metrics[v] = 0
				}
			}
		}
		rd = append(rd, MetricsData{
			metrics,
			target,
			dimension,
			timestamp,
		})
	}

	data, err := json.Marshal(map[string]interface{}{"data": rd})
	if err != nil {
		logger.Errorf("unable to marshal MetricData, error: %s\n", err)
		return
	}
	v.SendMsg(v.c.Host.HostMetricDataId, data, "主机", "Metrics")
}

package common

import "time"

const (
	Version           = "1.0.0"
	NameSpace         = "cloud_collection"
	DefaultSocketPath = "/var/run/gse/ipc.state.report"
)

// VMWare 相关的一些定义
var DefaultVMPeriod = time.Minute * 5 // 默认VM采集任务调度周期
const (
	VMAesKey               = "jski2ksuey4xn8fu"
	DiskUsedAvg            = "disk.used.average"
	MemTotalCapacityAverag = "mem.totalCapacity.average"
	MemTotalMBAverage      = "mem_totalmb_average"
)

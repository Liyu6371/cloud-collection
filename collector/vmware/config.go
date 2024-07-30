package vmware

type VMCollector struct {
	Concurrency int      `mapstructure:"concurrency"`
	Clouds      *[]Cloud `mapstructure:"clouds"`
}

type Cloud struct {
	Id       int      `mapstructure:"id"`
	Server   string   `mapstructure:"server"`
	Account  string   `mapstructure:"account"`
	Password string   `mapstructure:"password"`
	Cluster  *Cluster `mapstructure:"cluster"`
	Host     *Host    `mapstructure:"host"`
	Storage  *Storage `mapstructure:"storage"`
	VM       *VM      `mapstructure:"vm"`
}

type Cluster struct {
	ClusterMetricNamespace string     `mapstructure:"cluster_metric_namespace"`
	ClusterMetricDataId    int32      `mapstructure:"cluster_metric_data_id"`
	ClusterEventDataId     int32      `mapstructure:"cluster_event_data_id"`
	ClusterInstances       *[]string  `mapstructure:"cluster_instances"`
	ClusterMetrics         *[]Metrics `mapstructure:"cluster_metrics"`
}

type Host struct {
	HostMetricNamespace string     `mapstructure:"host_metric_namespace"`
	HostMetricDataId    int32      `mapstructure:"host_metric_data_id"`
	HostEventDataId     int32      `mapstructure:"host_event_data_id"`
	HostInstances       *[]string  `mapstructure:"host_instances"`
	HostMetrics         *[]Metrics `mapstructure:"host_metrics"`
}

type Storage struct {
	StorageMetricNamespace string     `mapstructure:"storage_metric_namespace"`
	StorageMetricDataId    int32      `mapstructure:"storage_metric_data_id"`
	StorageEventDataId     int32      `mapstructure:"storage_event_data_id"`
	StorageInstances       *[]string  `mapstructure:"storage_instances"`
	StorageMetrics         *[]Metrics `mapstructure:"storage_metrics"`
}

type VM struct {
	VMMetricNamespace string     `mapstructure:"vm_metric_namespace"`
	VMMetricDataId    int32      `mapstructure:"vm_metric_data_id"`
	VMEventDataId     int32      `mapstructure:"vm_event_data_id"`
	VMInstances       *[]string  `mapstructure:"vm_instances"`
	VMMetrics         *[]Metrics `mapstructure:"vm_metrics"`
}

type Metrics struct {
	Alias  string `mapstructure:"alias"`
	Metric string `mapstructure:"metric"`
}

type Dimension struct {
	CloudID    int    `json:"cloud_id"`
	InstanceID string `json:"instanceid"`
	Type       string `json:"type"`
	DeviceName string `json:"device_name"`
}

type MemTotalValue struct {
	key           string
	totalMemories int64
}

type storeSpace struct {
	Capacity int64
	Used     int64
}

type MetricsData struct {
	Metrics   map[string]float64 `json:"metrics"`
	Target    string             `json:"target"`
	Dimension Dimension          `json:"dimension"`
	Timestamp int64              `json:"timestamp"`
}

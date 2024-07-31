package vmware

import (
	"cloud-collection/common"
	"cloud-collection/logger"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// collectEventByTime
// 根据给的时间以及 instanceType 来收集不同类型 instance 一定时间范围内的 Event 数据
func (v *VMCollectorTask) collectEventByTime(dataid int32, instanceType string, instances []string, des string, cloudId int, timeRange string) {
	defer v.wg.Done()
	refList := v.getReference(instanceType, instances)
	if len(refList) == 0 {
		logger.Warnln("length of refList is zero, nothing to do!")
		return
	}
	if v.client == nil {
		logger.Errorf("client is nil.")
		return
	}

	// 计算查询 Event 的时间范围
	duration := time.Minute * 5
	if d, err := time.ParseDuration(timeRange); err == nil {
		duration = d
	}
	now := time.Now()
	timeBefore := now.Add(-duration)

	var eventList []EventInstance
	eventManager := event.NewManager(v.client.Client)
	for _, item := range refList {
		filter := types.EventFilterSpec{
			Entity: &types.EventFilterSpecByEntity{
				Entity:    item,
				Recursion: types.EventFilterSpecRecursionOptionAll,
			},
			Time: &types.EventFilterSpecByTime{
				BeginTime: &now,
				EndTime:   &timeBefore,
			},
		}

		events, err := eventManager.QueryEvents(v.ctx, filter)
		if err != nil {
			logger.Errorf("unable to query events, instanceType: %s, instances: %+v, timeRange: %s, error: %s\n", instanceType, instances, timeRange, err)
			return
		}
		for _, e := range events {
			info := e.GetEvent()
			eventList = append(eventList, EventInstance{
				getEventType(e),
				EventContent{info.FullFormattedMessage},
				getEventTarget(info),
				EventDimension{info.UserName, cloudId, item.Value},
				info.CreatedTime.UnixNano() / 1e6,
			})
		}
	}

	if len(eventList) == 0 {
		logger.Infof("%s no event data\n", des)
		return
	}

	eventData := map[string]interface{}{"data": eventList}
	if data, err := json.Marshal(eventData); err == nil {
		v.sendMsg(dataid, data, des, "event")
	} else {
		logger.Errorf("unable to marshal eventData, error: %s\n", err)
		return
	}

}

func (v *VMCollectorTask) getReference(instanceType string, instanceList []string) []types.ManagedObjectReference {
	client := v.getClient()
	if client == nil {
		return nil
	}

	m := view.NewManager(client.Client)
	var kind []string
	var objects interface{}

	switch instanceType {
	case "cluster":
		kind = []string{"ClusterComputeResource"}
		objects = new([]mo.ClusterComputeResource)
	case "host":
		kind = []string{"HostSystem"}
		objects = new([]mo.HostSystem)
	case "vm":
		kind = []string{"VirtualMachine"}
		objects = new([]mo.VirtualMachine)
	case "storage":
		kind = []string{"Datastore"}
		objects = new([]mo.Datastore)
	default:
		logger.Warnf("unsupported type: %s\n", instanceType)
		return nil
	}

	return v.retrieveReferences(m, kind, instanceList, objects)
}

func (v *VMCollectorTask) retrieveReferences(m *view.Manager, kind []string, instanceList []string, objects interface{}) []types.ManagedObjectReference {
	var refList []types.ManagedObjectReference
	client := v.client
	if client == nil {
		return refList
	}
	containerView, err := m.CreateContainerView(v.ctx, client.ServiceContent.RootFolder, kind, true)
	if err != nil {
		logger.Errorf("unable to create container view, error: %s\n", err)
		return refList
	}
	defer containerView.Destroy(v.ctx)

	err = containerView.Retrieve(v.ctx, kind, []string{"summary"}, objects)
	if err != nil {
		logger.Errorf("retrieve error: %s\n", err)
		return refList
	}

	// Use reflection to handle different types in a generic way
	val := reflect.ValueOf(objects).Elem()
	for i := 0; i < val.Len(); i++ {
		rf := val.Index(i).Interface().(mo.ManagedEntity).Reference()
		if common.IsContain(instanceList, rf.Value) {
			refList = append(refList, rf)
		}
	}

	return refList
}

func getEventType(e interface{}) string {
	et := reflect.TypeOf(e).String()
	return strings.Replace(et, "*types.", "", -1)
}

func getEventTarget(e *types.Event) string {
	if e.Vm != nil {
		return fmt.Sprintf("[%s][%s][%s]", e.Vm.Vm.Type, e.Vm.Vm.Value, e.Vm.Name)
	}
	if e.Host != nil {
		return fmt.Sprintf("[%s][%s][%s]", e.Host.Host.Type, e.Host.Host.Value, e.Host.Name)
	}
	if e.ComputeResource != nil {
		return fmt.Sprintf("[%s][%s][%s]", e.ComputeResource.ComputeResource.Type, e.ComputeResource.ComputeResource.Value, e.ComputeResource.Name)
	}
	if e.Datacenter != nil {
		return fmt.Sprintf("[%s][%s][%s]", e.Datacenter.Datacenter.Type, e.Datacenter.Datacenter.Value, e.Datacenter.Name)
	}
	return ""
}

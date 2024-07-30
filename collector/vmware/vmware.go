package vmware

import (
	"cloud-collection/logger"
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

var Name = "VMWare"

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

func (v *VMCollector) Run(ctx context.Context) {
	fmt.Println("hahahaha")
	fmt.Println(v.Concurrency)
	for _, v := range *v.Clouds {
		fmt.Printf("%+v\n", v)
	}
}

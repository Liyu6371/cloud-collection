package vmware

import (
	"cloud-collection/logger"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func (v *VMCollectorTask) getStoreInfo() map[string]storeSpace {
	client := v.getClient()
	m := view.NewManager(client.Client)
	kind := []string{"Datastore"}
	view, err := m.CreateContainerView(v.ctx, client.ServiceContent.RootFolder, kind, true)
	if err != nil {
		logger.Errorf("unable to create container view: %v\n", err)
		return nil
	}
	var stores []mo.Datastore
	err = view.Retrieve(v.ctx, kind, []string{"summary"}, &stores)
	if err != nil {
		logger.Errorf("unable to retrieve stores: %v\n", err)
		return nil
	}
	storeSpaceMap := make(map[string]storeSpace)
	for _, store := range stores {
		storeSpaceMap[store.Reference().Value] = storeSpace{
			store.Summary.Capacity,
			store.Summary.Capacity - store.Summary.FreeSpace}
	}
	return storeSpaceMap
}

package controller

import (
	"context"
	"errors"

	"github.com/oracle/oci-go-sdk/v52/common"
	"github.com/oracle/oci-go-sdk/v52/core"
)

type coreController struct {
	computeClient *core.ComputeClient
	initiated     bool
}

func newCoreController() *coreController {
	return &coreController{
		computeClient: nil,
		initiated:     false,
	}
}

func (controller *coreController) init(ConfigProvider *common.ConfigurationProvider) error {
	if c, err := core.NewComputeClientWithConfigurationProvider(*ConfigProvider); err == nil {
		controller.computeClient = &c
		controller.initiated = true
		return nil
	} else {
		controller.initiated = false
		return err
	}
}

func (controller *coreController) ListInstances(Ctx context.Context,
	CompartmentId string,
	Limit int,
	Page string,
	SortBy core.ListInstancesSortByEnum,
	SortOrder core.ListInstancesSortOrderEnum,
	LifecycleState core.InstanceLifecycleStateEnum) (instances []core.Instance, nextPage string, err error) {

	if !controller.initiated {
		return nil, "", errors.New("core Controller not initiated")
	}

	request := core.ListInstancesRequest{
		CompartmentId:  common.String(CompartmentId),
		Limit:          common.Int(Limit),
		Page:           common.String(Page),
		SortBy:         SortBy,
		SortOrder:      SortOrder,
		LifecycleState: LifecycleState,
	}

	response, err := controller.computeClient.ListInstances(Ctx, request)
	if err != nil {
		return nil, "", err
	}
	// OpcNextPage can be nil
	var p string
	if response.OpcNextPage != nil {
		p = *response.OpcNextPage
	} else {
		p = ""
	}

	return response.Items, p, nil
}

func (controller *coreController) InstanceAction(Ctx context.Context, Ocid *string, Action core.InstanceActionActionEnum) (instance *core.Instance, err error) {
	if !controller.initiated {
		return nil, errors.New("core Controller not initiated")
	}
	reques := core.InstanceActionRequest{
		InstanceId: Ocid,
		Action:     Action,
	}

	response, err := controller.computeClient.InstanceAction(Ctx, reques)
	if err != nil {
		return nil, err
	}
	return &response.Instance, nil
}

func (controller *coreController) GetInstance(Ctx context.Context, OcidId string) (instance *core.Instance, err error) {
	if !controller.initiated {
		return nil, errors.New("core Controller not initiated")
	}
	req := core.GetInstanceRequest{InstanceId: common.String(OcidId)}
	response, err := controller.computeClient.GetInstance(Ctx, req)
	if err != nil {
		return nil, err
	}
	return &response.Instance, nil
}

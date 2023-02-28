package controller

import (
	"context"
	"time"

	"github.com/oracle/oci-go-sdk/v52/common"
	"github.com/oracle/oci-go-sdk/v52/core"
	"github.com/oracle/oci-go-sdk/v52/identity"
)

type OCIController struct {
	configFilePath, configProfile string
	configProvider                *common.ConfigurationProvider
	context                       context.Context
	cancelContext                 func()
	identityCtrl                  *identityController
	coreCtrl                      *coreController
	monitoringCtrl                *monitoringController
}

func NewOCIControllerDefault() *OCIController {
	return NewOCIControler("", "")
}

func NewOCIControler(filePath string, profile string) *OCIController {
	res := OCIController{
		configFilePath: "",
		configProfile:  "",
		identityCtrl:   newIdentityController(),
		coreCtrl:       newCoreController(),
		monitoringCtrl: newMonitoringController(),
		configProvider: nil,
	}
	res.context, res.cancelContext = context.WithCancel(context.Background())
	res.ReloadConfig(filePath, profile)
	return &res
}

func (controller *OCIController) GetInstance(Ocid string) (*core.Instance, error) {
	return controller.coreCtrl.GetInstance(controller.context, Ocid)
}

func (controller *OCIController) CloseContext() {
	controller.cancelContext()
}

func (controller *OCIController) IsChangedConfig(filePath string, profile string) bool {
	return (controller.configFilePath != filePath || controller.configProfile != profile)
}

func (controller *OCIController) ReloadConfig(filePath string, profile string) error {
	if controller.IsChangedConfig(filePath, profile) || controller.configProvider == nil {
		if filePath == "" && profile == "" {

			var conf common.ConfigurationProvider = common.DefaultConfigProvider()
			controller.configProvider = &conf
		} else {

			var conf common.ConfigurationProvider = common.CustomProfileConfigProvider(filePath, profile)
			controller.configProvider = &conf
		}
		controller.configFilePath = filePath
		controller.configProfile = profile
		return controller.reoladControllers()
	}
	return nil
}

func (controller *OCIController) ChangeRegion(region string) {
	controller.identityCtrl.client.SetRegion(region)
	controller.coreCtrl.computeClient.SetRegion(region)
}

func (controller *OCIController) reoladControllers() error {
	if err := controller.identityCtrl.init(controller.configProvider); err != nil {
		return err
	}

	if err := controller.coreCtrl.init(controller.configProvider); err != nil {
		return err
	}

	if err := controller.monitoringCtrl.init(controller.configProvider); err != nil {
		return err
	}
	return nil
}

func (controller *OCIController) ListRegions() (regions []identity.Region, err error) {
	return controller.identityCtrl.ListRegions(controller.context)
}

func (controller *OCIController) ListAllCompartments() (compartments []identity.Compartment, err error) {
	confPrv := *(controller.configProvider)
	tenancyID, err := confPrv.TenancyOCID()
	if err != nil {
		return nil, err
	}
	return controller.identityCtrl.ListAllCompartments(controller.context, tenancyID)
}

func (controller *OCIController) GetConfigurationProvider() common.ConfigurationProvider {
	return *controller.configProvider
}

func (controller *OCIController) ListInstances(compartmentId string,
	limit int,
	sortBy core.ListInstancesSortByEnum,
	sortOrder core.ListInstancesSortOrderEnum,
	lifecycleState core.InstanceLifecycleStateEnum,
	page string) (instances []core.Instance, nextPage string, err error) {
	return controller.coreCtrl.ListInstances(controller.context, compartmentId, limit, page, sortBy, sortOrder, lifecycleState)
}

func (controller *OCIController) ListCompartments(compartmentId string,
	limit int,
	accessLevel identity.ListCompartmentsAccessLevelEnum,
	sortBy identity.ListCompartmentsSortByEnum,
	sortOrder identity.ListCompartmentsSortOrderEnum,
	lifecycleState identity.CompartmentLifecycleStateEnum,
	page string) (compartments []identity.Compartment, nextPage string, err error) {
	return controller.identityCtrl.ListCompartments(controller.context, compartmentId, limit, accessLevel, sortBy, sortOrder, lifecycleState, page)
}

func (controller *OCIController) ExecuteInstanceAction(instanceOCID *string, action core.InstanceActionActionEnum) (instance *core.Instance, err error) {
	return controller.coreCtrl.InstanceAction(controller.context, instanceOCID, action)
}

// monitoring functions
func (controller *OCIController) CpuUtilization10mLast24hMax(compartmentId string,
	instanceId string) (map[float64]float64, error) {
	return controller.monitoringCtrl.getMetrics(
		controller.context, "CpuUtilization", "10m", instanceId, "max", compartmentId, time.Now().AddDate(0, 0, -1), time.Now())
}

func (controller *OCIController) MemoryUtilization10mLast24hMax(compartmentId string,
	instanceId string) (map[float64]float64, error) {
	return controller.monitoringCtrl.getMetrics(
		controller.context, "MemoryUtilization", "10m", instanceId, "max", compartmentId, time.Now().AddDate(0, 0, -1), time.Now())
}

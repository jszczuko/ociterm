package controller

import (
	"context"
	"errors"

	"github.com/oracle/oci-go-sdk/v52/common"
	"github.com/oracle/oci-go-sdk/v52/identity"
)

type identityController struct {
	client    *identity.IdentityClient
	initiated bool
}

func newIdentityController() *identityController {
	return &identityController{
		client:    nil,
		initiated: false,
	}
}

func (controller *identityController) init(configProvider *common.ConfigurationProvider) error {
	if c, err := identity.NewIdentityClientWithConfigurationProvider(*configProvider); err == nil {
		controller.client = &c
		controller.initiated = true
		return nil
	} else {
		controller.initiated = false
		return err
	}
}

func (controller *identityController) ListRegions(ctx context.Context) (regions []identity.Region, err error) {
	if !controller.initiated {
		return nil, errors.New("identity Controller not initiated")
	}
	response, err := controller.client.ListRegions(ctx)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (controller *identityController) ListAllCompartments(ctx context.Context, cmp string) (compartments []identity.Compartment, err error) {
	if !controller.initiated {
		return nil, errors.New("identity Controller not initiated")
	}
	req := identity.ListCompartmentsRequest{AccessLevel: identity.ListCompartmentsAccessLevelAccessible,
		CompartmentId:          common.String(cmp),
		CompartmentIdInSubtree: common.Bool(true),
		SortOrder:              identity.ListCompartmentsSortOrderAsc,
		LifecycleState:         identity.CompartmentLifecycleStateActive,
		SortBy:                 identity.ListCompartmentsSortByName}
	response, err := controller.client.ListCompartments(ctx, req)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (controller *identityController) ListCompartments(ctx context.Context,
	compartmentId string,
	limit int,
	accessLevel identity.ListCompartmentsAccessLevelEnum,
	sortBy identity.ListCompartmentsSortByEnum,
	sortOrder identity.ListCompartmentsSortOrderEnum,
	lifecycleState identity.CompartmentLifecycleStateEnum,
	page string) (compartments []identity.Compartment, nextPage string, err error) {

	if !controller.initiated {
		return nil, "", errors.New("identity Controller not initiated")
	}
	req := identity.ListCompartmentsRequest{
		AccessLevel:            accessLevel,
		CompartmentId:          common.String(compartmentId),
		CompartmentIdInSubtree: common.Bool(false),
		SortOrder:              sortOrder,
		LifecycleState:         lifecycleState,
		SortBy:                 sortBy,
		Limit:                  common.Int(limit),
		Page:                   common.String(page),
	}
	response, err := controller.client.ListCompartments(ctx, req)
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

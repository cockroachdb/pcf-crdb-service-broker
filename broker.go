package main

import (
	"context"
	"errors"

	"github.com/pivotal-cf/brokerapi"
)

type CRDBServiceBroker struct{}

const (
	serviceID   = "64b3f845-7de2-4e95-8c4a-214808e013c6" // an arbitrary GUID
	serviceName = "cockroachdb"
	planID      = "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00" // an arbitrary GUID
	planName    = "default"
)

// Services is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Services(context context.Context) []brokerapi.Service {
	plans := []brokerapi.ServicePlan{
		{
			ID:          planID,
			Name:        planName,
			Description: "This plan is used to connect to an externally managed CockroachDB cluster.",
			Metadata: &brokerapi.ServicePlanMetadata{
				Bullets: []string{
					"Each instance shares the same cluster",
				},
				DisplayName: "default",
			},
		},
	}

	return []brokerapi.Service{
		brokerapi.Service{
			ID:          serviceID,
			Name:        serviceName,
			Description: "This service is used to connect to an externally managed CockroachDB cluster.",
			Bindable:    true,
			Plans:       plans,
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName:      serviceName,
				LongDescription:  "",
				DocumentationUrl: "",
				SupportUrl:       "",
				//ImageUrl:            fmt.Sprintf("data:image/png;base64,%s", redisServiceBroker.Config.RedisConfiguration.IconImage),
				ProviderDisplayName: "CockroachDB",
			},
			Tags: []string{
				"pivotal",
				"cockroachdb",
			},
		},
	}
}

// Provision is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Provision(
	context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool,
) (brokerapi.ProvisionedServiceSpec, error) {
	return brokerapi.ProvisionedServiceSpec{}, errors.New("not implemented")
}

// Deprovision is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Deprovision(
	context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool,
) (brokerapi.DeprovisionServiceSpec, error) {
	return brokerapi.DeprovisionServiceSpec{}, errors.New("not implemented")
}

// Bind is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Bind(
	context context.Context, instanceID, bindingID string, details brokerapi.BindDetails,
) (brokerapi.Binding, error) {
	return brokerapi.Binding{}, errors.New("not implemented")
}

// Unbind is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Unbind(
	context context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails,
) error {
	return errors.New("not implemented")
}

// Update is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Update(
	context context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool,
) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, errors.New("not implemented")
}

// LastOperation is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) LastOperation(
	context context.Context, instanceID, operationData string,
) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/lib/pq" // initialize the postgres sql driver

	"github.com/dchest/uniuri"

	"github.com/pivotal-cf/brokerapi"
)

const (
	serviceID   = "64b3f845-7de2-4e95-8c4a-214808e013c6" // an arbitrary GUID
	serviceName = "cockroachdb"
	planID      = "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00" // an arbitrary GUID
	planName    = "default"
)

type CRDBServiceInstance struct {
	dbName string
}

type CRDBServiceBroker struct {
	crdb *sql.DB

	mu struct {
		sync.Mutex
		instances map[string]CRDBServiceInstance
	}
}

func newCRDBServiceBroker(host, port, user, pass string) (*CRDBServiceBroker, error) {
	if host == "" {
		return nil, errors.New("CockroachDB host not specified")
	}
	if port == "" {
		return nil, errors.New("CockroachDB port not specified")
	}

	var dbUrl string
	if user == "" {
		dbUrl = fmt.Sprintf("postgres://%s:%s/?sslmode=disable", host, port)
	} else if pass == "" {
		dbUrl = fmt.Sprintf("postgres://%s@%s:%s/?sslmode=disable", user, host, port)
	} else {
		dbUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/?sslmode=disable", user, pass, host, port)
	}
	crdb, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	return &CRDBServiceBroker{crdb: crdb}, nil
}

func (sb *CRDBServiceBroker) getInstance(instanceID string) (CRDBServiceInstance, bool) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	info, ok := sb.mu.instances[instanceID]
	return info, ok
}

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
	if details.PlanID != planID {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("unknown plan ID '%s'", details.PlanID)
	}
	if _, exists := sb.getInstance(instanceID); exists {
		return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceAlreadyExists
	}
	dbName := uniuri.New()

	// Create database.
	if _, err := sb.crdb.Exec("CREATE DATABASE $1", dbName); err != nil {
		log.Error("create-database", err)
		return brokerapi.ProvisionedServiceSpec{}, err
	}
	sb.mu.Lock()
	// We could have a parallel call with the same instanceID, but if that happens
	// one of them will fail to create the database.
	sb.mu.instances[instanceID] = CRDBServiceInstance{dbName: dbName}
	sb.mu.Unlock()
	return brokerapi.ProvisionedServiceSpec{}, nil
}

// Deprovision is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) Deprovision(
	context context.Context,
	instanceID string,
	details brokerapi.DeprovisionDetails,
	asyncAllowed bool,
) (brokerapi.DeprovisionServiceSpec, error) {
	instance, exists := sb.getInstance(instanceID)
	if !exists {
		return brokerapi.DeprovisionServiceSpec{}, brokerapi.ErrInstanceDoesNotExist
	}

	// Delete database.
	if _, err := sb.crdb.Exec("DROP DATABASE $1", instance.dbName); err != nil {
		log.Error("drop-database", err)
		return brokerapi.DeprovisionServiceSpec{}, err
	}
	sb.mu.Lock()
	// We could have a parallel call with the same instanceID, but if that happens
	// one of them will fail to delete the database.
	delete(sb.mu.instances, instanceID)
	sb.mu.Unlock()

	return brokerapi.DeprovisionServiceSpec{}, nil
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
	return brokerapi.UpdateServiceSpec{}, nil
}

// LastOperation is part of the brokerapi.ServiceBroker interface.
func (sb *CRDBServiceBroker) LastOperation(
	context context.Context, instanceID, operationData string,
) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

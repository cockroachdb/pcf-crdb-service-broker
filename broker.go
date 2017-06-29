package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

type crdbBinding struct {
	user string
}

type crdbServiceInstance struct {
	host   string
	port   string
	dbName string

	mu struct {
		sync.Mutex
		bindings map[string]*crdbBinding
	}
}

func newCRDBServiceInstance(host, port, dbName string) *crdbServiceInstance {
	si := &crdbServiceInstance{
		host:   host,
		port:   port,
		dbName: dbName,
	}
	si.mu.bindings = make(map[string]*crdbBinding)
	return si
}

func (si *crdbServiceInstance) getBinding(bindingID string) *crdbBinding {
	si.mu.Lock()
	defer si.mu.Unlock()
	return si.mu.bindings[bindingID]
}

type crdbServiceBroker struct {
	host string
	port string
	crdb *sql.DB

	mu struct {
		sync.Mutex
		instances map[string]*crdbServiceInstance
	}
}

// dbURI creates a URI that can be used to connect to CockroachDB;
// user, pass, and db are optional.
func dbURI(host, port, user, pass, db string) string {
	if host == "" || port == "" {
		panic("host/port not passed")
	}
	if user == "" {
		return fmt.Sprintf("postgres://%s:%s/%s?sslmode=disable", host, port, db)
	}
	if pass == "" {
		return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", user, host, port, db)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func newCRDBServiceBroker(host, port, user, pass string) (*crdbServiceBroker, error) {
	if host == "" {
		return nil, errors.New("CockroachDB host not specified")
	}
	if port == "" {
		return nil, errors.New("CockroachDB port not specified")
	}

	crdb, err := sql.Open("postgres", dbURI(host, port, user, pass, "" /* no database */))
	if err != nil {
		return nil, err
	}
	sb := &crdbServiceBroker{
		host: host,
		port: port,
		crdb: crdb,
	}
	sb.mu.instances = make(map[string]*crdbServiceInstance)
	return sb, nil
}

// getInstance retrieves a service instance; returns nil if it doesn't exist.
func (sb *crdbServiceBroker) getInstance(instanceID string) *crdbServiceInstance {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.mu.instances[instanceID]
}

// Services is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Services(context context.Context) []brokerapi.Service {
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

// getDBName creates a database name that is associated with an instance.
// Example:
//   getDBName("c5c7fcbd-618a-4de0-953a-d4e357acc22a")
// returns
//   "cf_c5c7fcbd_618a_4de0_953a_d4e357acc22a"

func getDBName(instanceID string) string {
	// instanceID should be a GUID; verify that it has no special characters just
	// in case.
	for _, s := range instanceID {
		if !(s == '-' || (s >= 'a' && s <= 'z') || (s >= 'A' && s <= 'Z') || (s >= '0' && s <= '9')) {
			return uniuri.New()
		}
	}
	return "cf_" + strings.Replace(instanceID, "-", "_", -1)
}

// Provision is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Provision(
	context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool,
) (brokerapi.ProvisionedServiceSpec, error) {
	if details.PlanID != planID {
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("unknown plan ID '%s'", details.PlanID)
	}
	if sb.getInstance(instanceID) != nil {
		return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceAlreadyExists
	}
	// Generate a random string for the database name.
	// TODO(radu): allow the user to pass the name in the details.
	dbName := getDBName(instanceID)

	// Create database.
	if _, err := sb.crdb.Exec("CREATE DATABASE " + dbName); err != nil {
		log.Error("create-database", err)
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("creating database: %s", err)
	}
	sb.mu.Lock()
	// TODO(radu) We could have a parallel call with the same instanceID;
	// delete the database in this case.
	sb.mu.instances[instanceID] = newCRDBServiceInstance(sb.host, sb.port, dbName)
	sb.mu.Unlock()
	return brokerapi.ProvisionedServiceSpec{}, nil
}

// Deprovision is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Deprovision(
	context context.Context,
	instanceID string,
	details brokerapi.DeprovisionDetails,
	asyncAllowed bool,
) (brokerapi.DeprovisionServiceSpec, error) {
	instance := sb.getInstance(instanceID)
	if instance == nil {
		// Nothing to do.
		return brokerapi.DeprovisionServiceSpec{}, nil
	}

	// Delete database.
	if _, err := sb.crdb.Exec("DROP DATABASE IF EXISTS " + instance.dbName); err != nil {
		log.Error("drop-database", err)
		return brokerapi.DeprovisionServiceSpec{}, fmt.Errorf("dropping database: %s", err)
	}
	sb.mu.Lock()
	// We could have a parallel call with the same instanceID, but that's ok.
	delete(sb.mu.instances, instanceID)
	sb.mu.Unlock()

	return brokerapi.DeprovisionServiceSpec{}, nil
}

// Bind is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Bind(
	context context.Context, instanceID, bindingID string, details brokerapi.BindDetails,
) (brokerapi.Binding, error) {
	instance := sb.getInstance(instanceID)
	if instance == nil {
		return brokerapi.Binding{}, brokerapi.ErrInstanceDoesNotExist
	}
	if instance.getBinding(bindingID) != nil {
		return brokerapi.Binding{}, brokerapi.ErrBindingAlreadyExists
	}
	// TODO(radu): allow user/pass to be passed through BindDetails.RawParameters
	user := uniuri.New()
	pass := uniuri.New()

	if _, err := sb.crdb.Exec(
		fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", user, pass),
	); err != nil {
		log.Error("create-user", err)
		return brokerapi.Binding{}, fmt.Errorf("creating user: %s", err)
	}

	if _, err := sb.crdb.Exec(
		fmt.Sprintf("GRANT ALL ON DATABASE %s TO %s", instance.dbName, user),
	); err != nil {
		log.Error("grant-privileges", err)
		// TODO(radu): delete the user
		return brokerapi.Binding{}, fmt.Errorf("granting privileges: %s", err)
	}

	credMap := map[string]interface{}{
		"host":     instance.host,
		"port":     instance.port,
		"database": instance.dbName,
		"username": user,
		"password": pass,
		"uri":      dbURI(instance.host, instance.port, user, pass, instance.dbName),
	}

	instance.mu.Lock()
	instance.mu.bindings[bindingID] = &crdbBinding{user: user}
	instance.mu.Unlock()

	return brokerapi.Binding{Credentials: credMap}, nil
}

// Unbind is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Unbind(
	context context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails,
) error {
	instance := sb.getInstance(instanceID)
	if instance == nil {
		return brokerapi.ErrInstanceDoesNotExist
	}
	binding := instance.getBinding(bindingID)
	if binding == nil {
		return brokerapi.ErrBindingDoesNotExist
	}

	if _, err := sb.crdb.Exec(fmt.Sprintf("DROP USER IF EXISTS %s", binding.user)); err != nil {
		log.Error("drop-user", err)
		return fmt.Errorf("deleting user: %s", err)
	}

	instance.mu.Lock()
	delete(instance.mu.bindings, bindingID)
	instance.mu.Unlock()
	return nil
}

// Update is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Update(
	context context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool,
) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, nil
}

// LastOperation is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) LastOperation(
	context context.Context, instanceID, operationData string,
) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

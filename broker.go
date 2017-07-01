// Copyright 2017 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"context"
	"fmt"
	"regexp"

	_ "github.com/lib/pq" // initialize the postgres sql driver

	"github.com/dchest/uniuri"

	"github.com/pivotal-cf/brokerapi"
)

type crdbServiceBroker struct {
}

func newCRDBServiceBroker() *crdbServiceBroker {
	return &crdbServiceBroker{}
}

// Services is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Services(context context.Context) []brokerapi.Service {
	services := make([]brokerapi.Service, len(Services))
	for i, s := range Services {
		services[i] = s.Service
		services[i].Plans = make([]brokerapi.ServicePlan, len(s.Plans))
		for j, p := range s.Plans {
			services[i].Plans[j] = p.ServicePlan
		}
	}

	return services
}

var dbExistsErrRegexp = regexp.MustCompile("database .* already exists")
var dbNotFoundErrRegexp = regexp.MustCompile("database .* does not exist")

// Provision is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Provision(
	context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool,
) (brokerapi.ProvisionedServiceSpec, error) {
	plan, err := findPlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	// Generate a database name from the instanceID.
	dbName := dbNameFromInstanceID(instanceID)

	// Create database.
	if _, err := plan.crdb.Exec("CREATE DATABASE " + dbName); err != nil {
		if dbExistsErrRegexp.MatchString(err.Error()) {
			return brokerapi.ProvisionedServiceSpec{}, brokerapi.ErrInstanceAlreadyExists
		}
		log.Error("create-database", err)
		return brokerapi.ProvisionedServiceSpec{}, fmt.Errorf("creating database: %s", err)
	}
	return brokerapi.ProvisionedServiceSpec{}, nil
}

// Deprovision is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Deprovision(
	context context.Context,
	instanceID string,
	details brokerapi.DeprovisionDetails,
	asyncAllowed bool,
) (brokerapi.DeprovisionServiceSpec, error) {
	plan, err := findPlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}
	dbName := dbNameFromInstanceID(instanceID)

	// Delete database.
	if _, err := plan.crdb.Exec("DROP DATABASE IF EXISTS " + dbName); err != nil {
		log.Error("drop-database", err)
		return brokerapi.DeprovisionServiceSpec{}, fmt.Errorf("dropping database: %s", err)
	}

	return brokerapi.DeprovisionServiceSpec{}, nil
}

// Bind is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Bind(
	context context.Context, instanceID, bindingID string, details brokerapi.BindDetails,
) (brokerapi.Binding, error) {
	plan, err := findPlan(details.ServiceID, details.PlanID)
	if err != nil {
		return brokerapi.Binding{}, err
	}
	dbName := dbNameFromInstanceID(instanceID)
	user := userNameFromBinding(instanceID, bindingID)
	pass := uniuri.New()

	if _, err := plan.crdb.Exec(
		fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", user, pass),
	); err != nil {
		log.Error("create-user", err)
		return brokerapi.Binding{}, fmt.Errorf("creating user: %s", err)
	}

	if _, err := plan.crdb.Exec(
		fmt.Sprintf("GRANT ALL ON DATABASE %s TO %s", dbName, user),
	); err != nil {
		_, _ = plan.crdb.Exec("DROP USER %s", user)
		if dbNotFoundErrRegexp.MatchString(err.Error()) {
			return brokerapi.Binding{}, brokerapi.ErrInstanceDoesNotExist
		}
		log.Error("grant-privileges", err)
		// TODO(radu): delete the user
		return brokerapi.Binding{}, fmt.Errorf("granting privileges: %s", err)
	}

	credMap := map[string]interface{}{
		"host":     plan.CRDBHost,
		"port":     plan.CRDBPort,
		"database": dbName,
		"username": user,
		"password": pass,
		"uri":      dbURI(plan.CRDBHost, plan.CRDBPort, user, pass, dbName),
	}

	return brokerapi.Binding{Credentials: credMap}, nil
}

// Unbind is part of the brokerapi.ServiceBroker interface.
func (sb *crdbServiceBroker) Unbind(
	context context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails,
) error {
	plan, err := findPlan(details.ServiceID, details.PlanID)
	if err != nil {
		return err
	}

	user := userNameFromBinding(instanceID, bindingID)

	if _, err := plan.crdb.Exec(fmt.Sprintf("DROP USER IF EXISTS %s", user)); err != nil {
		log.Error("drop-user", err)
		return fmt.Errorf("deleting user: %s", err)
	}
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

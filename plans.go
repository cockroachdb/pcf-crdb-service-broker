package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/pivotal-cf/brokerapi"
)

// Each service plan can connect to a potentially different CockroachDB instance/cluster.
type Plan struct {
	brokerapi.ServicePlan

	ServiceID string `json: serviceID`

	CRDBHost      string `json: crdbHost`
	CRDBPort      string `json: crdbPort`
	CRDBAdminUser string `json: crdbAdminUser`
	CRDBPassword  string `json: crdbPassword`

	crdb *sql.DB `json: -`
}

type Service struct {
	// Note that the Plans field is not populated in this structure.
	brokerapi.Service
	Plans []Plan `json: -`
}

var Services []Service

func findService(serviceID string) (*Service, error) {
	for i := range Services {
		if Services[i].ID == serviceID {
			return &Services[i], nil
		}
	}
	return nil, fmt.Errorf("unknown service ID '%s'", serviceID)
}

func findPlan(serviceID, planID string) (*Plan, error) {
	s, err := findService(serviceID)
	if err != nil {
		return nil, err
	}
	for i := range s.Plans {
		if s.Plans[i].ID == planID {
			return &s.Plans[i], nil
		}
	}
	return nil, fmt.Errorf("unknown plan ID '%s'", planID)
}

func addService(svc Service) {
	if svc.ID == "" {
		log.Fatal("init", errors.New("service id required"))
	}
	for _, s := range Services {
		if s.Name == svc.Name {
			log.Fatal("init", fmt.Errorf("duplicate service name '%s'", svc.Name))
		}
		if s.ID == svc.ID {
			log.Fatal("init", fmt.Errorf("duplicate service id '%s'", svc.ID))
		}
	}
	Services = append(Services, svc)
}

func addPlan(p Plan) {
	if p.ServiceID == "" {
		log.Fatal("init", errors.New("serviceID required in plan"))
	}
	s, err := findService(p.ServiceID)
	if err != nil {
		log.Fatal("init", fmt.Errorf("unknown service ID '%s' in plan", p.ServiceID))
	}

	if p.ID == "" {
		// Generate an ID derived from the service and plan names. This ID is
		// deterministic so we don't need to store it anywhere.
		p.ID = generatePlanID(s.Name, p.Name)
	}
	for _, pl := range s.Plans {
		if pl.Name == p.Name {
			log.Fatal("init", fmt.Errorf("duplicate plan name '%s'", pl.Name))
		}
		if pl.ID == p.ID {
			log.Fatal("init", fmt.Errorf("duplicate plan id '%s'", pl.ID))
		}
	}

	if p.CRDBHost == "" || p.CRDBPort == "" {
		log.Fatal("init", fmt.Errorf("plan '%s' does not specify a CockroachDB host/port", p.Name))
	}
	p.crdb, err = sql.Open(
		"postgres",
		dbURI(p.CRDBHost, p.CRDBPort, p.CRDBAdminUser, p.CRDBPassword, "" /* no database */),
	)
	if err != nil {
		log.Fatal("init-setup-db", err)
	}

	s.Plans = append(s.Plans, p)
}

func init() {
	// Init services.
	var services []Service
	servicesJSON := os.Getenv("SERVICES")
	if servicesJSON == "" {
		log.Fatal("init", errors.New("SERVICES not specified"))
	}
	if err := json.Unmarshal([]byte(servicesJSON), &services); err != nil {
		log.Fatal("init-unmarshal-services", err)
	}
	if len(services) == 0 {
		log.Fatal("init", errors.New("no services"))
	}
	for _, s := range services {
		addService(s)
	}

	// Init static plans.
	var plans []Plan
	planJSON := os.Getenv("PRECONFIGURED_PLANS")
	if planJSON != "" {
		if err := json.Unmarshal([]byte(planJSON), &plans); err != nil {
			log.Fatal("init-unmarshal-preconfigured-plans", err)
		}
	}

	// TODO(radu): add dynamic plans.

	if len(plans) == 0 {
		log.Fatal("init", errors.New("no plans"))
	}
	for _, p := range plans {
		addPlan(p)
	}
}

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
	"os"
	"reflect"
	"testing"

	"github.com/pivotal-cf/brokerapi"
)

func TestInit(t *testing.T) {
	os.Setenv("SERVICES", `[
    {
      "id": "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
      "name": "cockroachdb",
      "description": "desc",
      "bindable": true,
      "plan_updateable": false,
      "metadata": {
        "displayName": "CockroachDB",
        "longDescription": "desc",
        "documentationUrl": "https://www.cockroachlabs.com/docs/",
        "supportUrl": "https://www.cockroachlabs.com/community/",
        "imageUrl": "https://www.cockroachlabs.com/images/CockroachLabs_Logo_Mark-lightbackground.svg"
      },
      "tags": ["cockroachdb", "relational"]
    }
  ]`)

	os.Setenv("PRECONFIGURED_PLANS", `[
    {
      "name": "default",
      "description": "Default",
      "metadata": { "displayName": "Default" },
      "serviceID": "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
      "crdbHost": "13.82.91.246",
      "crdbPort": "26257",
      "crdbAdminUser": "root"
    }
  ]`)

	os.Setenv("CUSTOM_PLANS", `{
    "plan1":{
      "guid":"94c58e75-ec11-470c-a866-ccfc54f24acf",
      "name":"plan1",
      "display_name":"plan1 name",
      "description":"plan1 desc",
      "service":"e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
      "host":"1.2.3.4",
      "port":26257
    },
    "plan2":{
      "guid":"411ad433-b087-4fe5-a5e1-b099c57c83ab",
      "name":"plan2",
      "display_name":"plan2 name",
      "description":"plan2 desc",
      "service":"e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
      "host":"5.6.7.8",
      "port":26257
    }
  }`)

	InitServicesAndPlans()

	expected := []Service{
		{
			Service: brokerapi.Service{
				ID:            "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
				Name:          "cockroachdb",
				Description:   "desc",
				Bindable:      true,
				Tags:          []string{"cockroachdb", "relational"},
				PlanUpdatable: false,
				Metadata: &brokerapi.ServiceMetadata{
					DisplayName:         "CockroachDB",
					ImageUrl:            "https://www.cockroachlabs.com/images/CockroachLabs_Logo_Mark-lightbackground.svg",
					LongDescription:     "desc",
					ProviderDisplayName: "",
					DocumentationUrl:    "https://www.cockroachlabs.com/docs/",
					SupportUrl:          "https://www.cockroachlabs.com/community/",
				},
			},
			Plans: []Plan{
				Plan{
					ServicePlan: brokerapi.ServicePlan{
						ID:          "d9a19cc6-fbae-597e-af9d-2c3fc640e42c",
						Name:        "default",
						Description: "Default",
						Free:        (*bool)(nil),
						Bindable:    (*bool)(nil),
						Metadata:    &brokerapi.ServicePlanMetadata{DisplayName: "Default"},
					},
					ServiceID:     "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
					CRDBHost:      "13.82.91.246",
					CRDBPort:      "26257",
					CRDBAdminUser: "root",
					CRDBPassword:  "",
				},
				Plan{
					ServicePlan: brokerapi.ServicePlan{
						ID:          "94c58e75-ec11-470c-a866-ccfc54f24acf",
						Name:        "plan1",
						Description: "plan1 desc",
						Free:        (*bool)(nil),
						Bindable:    (*bool)(nil),
						Metadata:    &brokerapi.ServicePlanMetadata{DisplayName: "plan1 name"},
					},
					ServiceID:     "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
					CRDBHost:      "1.2.3.4",
					CRDBPort:      "26257",
					CRDBAdminUser: "root",
					CRDBPassword:  "",
				},
				Plan{
					ServicePlan: brokerapi.ServicePlan{
						ID:          "411ad433-b087-4fe5-a5e1-b099c57c83ab",
						Name:        "plan2",
						Description: "plan2 desc",
						Free:        (*bool)(nil),
						Bindable:    (*bool)(nil),
						Metadata:    &brokerapi.ServicePlanMetadata{DisplayName: "plan2 name"},
					},
					ServiceID:     "e2e250b5-73f8-45fd-9a7f-93c8dddc5f00",
					CRDBHost:      "5.6.7.8",
					CRDBPort:      "26257",
					CRDBAdminUser: "root",
					CRDBPassword:  "",
				},
			},
		},
	}

	// Clear out fields we don't want to check.
	for i := range Services {
		for j := range Services[i].Plans {
			Services[i].Plans[j].crdb = nil
		}
	}

	if !reflect.DeepEqual(Services, expected) {
		t.Errorf("Expected\n%+v\ngot\n%+v", expected, Services)
	}
}

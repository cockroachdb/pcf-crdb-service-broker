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
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/pivotal-cf/brokerapi"

	"code.cloudfoundry.org/lager"
)

const defaultCockroachPort = 26257
const brokerPort = 8080

var log = lager.NewLogger("cockroachdb-broker")

func main() {
	log.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	log.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	log.Info("Starting CF CockroachDB broker")

	InitServicesAndPlans()
	defer CleanupPlans()

	serviceBroker := newCRDBServiceBroker()

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: os.Getenv("SECURITY_USER_NAME"),
		Password: os.Getenv("SECURITY_USER_PASSWORD"),
	}
	if brokerCredentials.Username == "" || brokerCredentials.Password == "" {
		log.Fatal("initializing-service", errors.New("SECURITY_USER_NAME/PASSWORD not set"))
	}

	brokerAPI := brokerapi.New(serviceBroker, log, brokerCredentials)
	http.Handle("/", brokerAPI)
	log.Fatal("http-listen", http.ListenAndServe(fmt.Sprintf(":%d", brokerPort), nil))
}

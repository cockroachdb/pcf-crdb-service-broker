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

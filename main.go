package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

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

	crHost := os.Getenv("CRDB_HOST")
	crPort := os.Getenv("CRDB_PORT")
	if crPort == "" {
		crPort = strconv.Itoa(defaultCockroachPort)
	}
	crUser := os.Getenv("CRDB_USER")
	crPass := os.Getenv("CRDN_PASS")

	serviceBroker, err := newCRDBServiceBroker(crHost, crPort, crUser, crPass)
	if err != nil {
		log.Fatal("initialize broker", err)
	}

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

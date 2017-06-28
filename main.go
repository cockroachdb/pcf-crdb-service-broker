package main

import (
	"net/http"
	"os"

	"github.com/pivotal-cf/brokerapi"

	"code.cloudfoundry.org/lager"
)

func main() {
	brokerLogger := lager.NewLogger("cockroachdb-broker")
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	brokerLogger.Info("Starting CF CockroachDB broker")

	serviceBroker := &CRDBServiceBroker{}

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: "user",
		Password: "pass",
	}

	brokerAPI := brokerapi.New(serviceBroker, brokerLogger, brokerCredentials)

	http.Handle("/", brokerAPI)
	brokerLogger.Fatal("http-listen", http.ListenAndServe(":8080", nil))
}

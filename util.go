package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// Namespace uuids used to generate V3/V5 UUIDs in conjunction with an arbitrary
// identifier. These are just some arbitrary GUIDs.
var (
	namespacePlans, _     = uuid.FromString("dab42b46-7339-4bf1-af9b-29dbb728bf52")
	namespaceInstances, _ = uuid.FromString("68751bc4-bd4e-42be-8fe9-dd37e628a186")
	namespaceUsernames, _ = uuid.FromString("4d8d2252-ce16-4c6f-b9a5-422835b89bed")
)

// generatePlanID generates a plan ID derived from the service name and plan
// name.
func generatePlanID(serviceName, planName string) string {
	return uuid.NewV5(namespacePlans, fmt.Sprintf("%s/%s", serviceName, planName)).String()
}

// uuidToChars generates a string of only lowercase characters from a UUID.
func uuidToChars(id uuid.UUID) string {
	res := make([]byte, len(id)*2)
	for i, b := range id {
		res[2*i] = 'a' + (b & 0xF)
		res[2*i+1] = 'a' + ((b >> 4) & 0xF)
	}
	return string(res)
}

func dbNameFromInstanceID(instanceID string) string {
	return "cf_" + uuidToChars(uuid.NewV5(namespaceInstances, instanceID))
}

func userNameFromBinding(instanceID, bindingID string) string {
	return uuidToChars(uuid.NewV5(namespaceUsernames, fmt.Sprintf("%s/%s", instanceID, bindingID)))
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

func jdbcURI(host, port, user, pass, db string) string {
	if host == "" || port == "" {
		panic("host/port not passed")
	}
	uri := fmt.Sprintf("jdbc:postgres://%s:%s/%s?sslmode=disable", host, port, db)
	if user != "" {
		uri = uri + fmt.Sprintf("&user=%s", user)
		if pass != "" {
			uri = uri + fmt.Sprintf("&password=%s", pass)
		}
	}
	return uri
}

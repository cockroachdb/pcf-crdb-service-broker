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
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

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
func dbURI(host, port, user, pass, db string, options url.Values) string {
	if host == "" || port == "" {
		panic("host/port not passed")
	}

	u := url.URL{
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     db,
		RawQuery: options.Encode(),
	}

	if user != "" {
		if pass != "" {
			u.User = url.UserPassword(user, pass)
		} else {
			u.User = url.User(user)
		}
	}
	return u.String()
}

func jdbcURL(host, port, user, pass, db string, options url.Values) string {
	if host == "" || port == "" {
		panic("host/port not passed")
	}
	url := fmt.Sprintf("jdbc:postgresql://%s:%s/%s", host, port, db)
	if user != "" {
		url = url + fmt.Sprintf("&user=%s", user)
		if pass != "" {
			url = url + fmt.Sprintf("&password=%s", pass)
		}
	}
	if len(options) > 0 {
		url = url + "?" + options.Encode()
	}
	return url
}

// createTempFile creates a temporary file and populates it with the given
// contents.
func createTempFile(prefix string, contents []byte) (path string, err error) {
	f, err := ioutil.TempFile("" /* default temp dir */, prefix)
	if err != nil {
		return "", err
	}
	path = f.Name()
	_, err = f.Write(contents)
	if closeErr := f.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

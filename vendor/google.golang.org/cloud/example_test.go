// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloud_test

import (
	"io/ioutil"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/pubsub"
)

func Example_createClientWithApplicationDefaultCredentials(ctx context.Context) *pubsub.Client {
	// Create a pubsub Client to demonstrate using Application Default
	// Credentials for authentication.
	//
	// Application Default Credentials provide a simple way to get
	// authorization credentials for use in calling Google APIs.  They are
	// best suited for cases when the call needs to have the same identity
	// and authorization level for the application independent of the user.
	// This is the recommended approach to authorize calls to Google Cloud
	// APIs, particularly when you're building an application that is
	// deployed to Google App Engine or Google Compute Engine virtual
	// machines.
	// See
	// https://developers.google.com/identity/protocols/application-default-credentials
	// for more information.
	//
	// The same approach may be used to construct Clients from other
	// packages, e.g. bigtable, datastore.
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		log.Fatal("new client:", err)
	}

	return client
}

func Example_createClientWithTokenSource(ctx context.Context) *pubsub.Client {
	jsonKey, err := ioutil.ReadFile("/path/to/json/keyfile.json")
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		pubsub.ScopeCloudPlatform,
		pubsub.ScopePubSub,
	)

	if err != nil {
		log.Fatal(err)
	}
	ts := conf.TokenSource(ctx)

	// Create a pubsub Client to demonstrate using an OAuth2 token source
	// for authentication.  The same approach may be used to construct
	// Clients from other packages, e.g. bigtable, datastore.
	client, err := pubsub.NewClient(ctx, "project-id", cloud.WithTokenSource(ts))
	if err != nil {
		log.Fatal("new client:", err)
	}

	return client
}

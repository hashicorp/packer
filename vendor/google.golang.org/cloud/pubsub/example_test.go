// Copyright 2014 Google Inc. All Rights Reserved.
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

package pubsub_test

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/cloud/pubsub"
)

func Example_createClient(ctx context.Context) *pubsub.Client {
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		log.Fatal("new client:", err)
	}

	// See the other examples to learn how to use the Client.
	return client
}

func ExampleTopicHandle_Publish() {
	ctx := context.Background()
	client := Example_createClient(ctx)

	topic := client.Topic("topicName")
	msgIDs, err := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello world"),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Published a message with a message ID: %s\n", msgIDs[0])
}

func ExampleSubscriptionHandle_Pull() {
	ctx := context.Background()
	client := Example_createClient(ctx)

	sub := client.Subscription("subName")
	it, err := sub.Pull(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()

	// Consume 10 messages.
	for i := 0; i < 10; i++ {
		m, err := it.Next()
		if err == pubsub.Done {
			// There are no more messages.  This will happen if it.Stop is called.
			break
		}
		if err != nil {
			log.Fatalf("advancing iterator: %v", err)
			break
		}
		log.Printf("message %d: %s\n", i, m.Data)

		// Acknowledge the message.
		m.Done(true)
	}
}

// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package ycsdk

import (
	"context"
	"net/url"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
)

type rpcCredentials struct {
	creds     ExchangeableCredentials
	plaintext bool

	// getConn set in Init.
	getConn lazyConn
	// now may be replaced in tests
	now func() time.Time

	// mutex guards conn and currentState, and excludes multiple simultaneous token updates
	mutex        sync.RWMutex
	conn         *grpc.ClientConn // initialized lazily from getConn
	currentState rpcCredentialsState
}

var _ credentials.PerRPCCredentials = &rpcCredentials{}

type rpcCredentialsState struct {
	token        string
	refreshAfter time.Time
	version      int64
}

func newRPCCredentials(creds ExchangeableCredentials, plaintext bool) *rpcCredentials {
	return &rpcCredentials{
		creds:     creds,
		plaintext: plaintext,
		now:       time.Now,
	}
}

func (c *rpcCredentials) Init(lazyConn lazyConn) {
	c.getConn = lazyConn
}

func (c *rpcCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	audienceURL, err := url.Parse(uri[0])
	if err != nil {
		return nil, err
	}
	if audienceURL.Path == "/yandex.cloud.iam.v1.IamTokenService" ||
		audienceURL.Path == "/yandex.cloud.endpoint.ApiEndpointService" {
		return nil, nil
	}

	c.mutex.RLock()
	state := c.currentState
	c.mutex.RUnlock()

	token := state.token
	outdated := state.refreshAfter.Before(c.now())
	if outdated {
		token, err = c.updateToken(ctx, state)
		if err != nil {
			return nil, err
		}
	}

	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

func (c *rpcCredentials) RequireTransportSecurity() bool {
	return !c.plaintext
}

func (c *rpcCredentials) updateToken(ctx context.Context, currentState rpcCredentialsState) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.currentState.version != currentState.version {
		// someone have already updated it
		return c.currentState.token, nil
	}

	if c.conn == nil {
		conn, err := c.getConn(ctx)
		if err != nil {
			return "", err
		}
		c.conn = conn
	}
	tokenClient := iam.NewIamTokenServiceClient(c.conn)

	tokenReq, err := c.creds.IAMTokenRequest()
	if err != nil {
		return "", sdkerrors.WithMessage(err, "failed to create IAM token request from credentials")
	}

	resp, err := tokenClient.Create(ctx, tokenReq)
	if err != nil {
		return "", err
	}

	c.currentState = rpcCredentialsState{
		token:        resp.IamToken,
		refreshAfter: c.now().Add(iamTokenExpiration),
		version:      currentState.version + 1,
	}

	return c.currentState.token, nil
}

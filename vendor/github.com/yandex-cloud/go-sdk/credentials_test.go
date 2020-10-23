// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package ycsdk

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-cloud/go-sdk/iamkey"
)

func TestOAuthToken(t *testing.T) {
	const token = "AAAA00000000000000000000000000000000000"
	creds := OAuthToken(token)
	iamTokenReq, err := creds.(ExchangeableCredentials).IAMTokenRequest()
	require.NoError(t, err)
	assert.Equal(t, token, iamTokenReq.GetYandexPassportOauthToken())
}

func TestIAMToken(t *testing.T) {
	const iamToken = "this-is-iam-token"
	creds := NewIAMTokenCredentials(iamToken)
	iamTokenResp, err := creds.(NonExchangeableCredentials).IAMToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, iamToken, iamTokenResp.GetIamToken())
}

func TestServiceAccountKey(t *testing.T) {
	key := testKey(t)
	creds, err := ServiceAccountKey(key)
	require.NoError(t, err)
	iamTokenReq, err := creds.(ExchangeableCredentials).IAMTokenRequest()
	require.NoError(t, err)

	require.NotEmpty(t, iamTokenReq.GetJwt())

	parser := jwt.Parser{}
	jot, parts, err := parser.ParseUnverified(iamTokenReq.GetJwt(), &jwt.StandardClaims{})
	require.NoError(t, err)

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key.PublicKey))
	require.NoError(t, err)

	// Force salt length: https://github.com/dgrijalva/jwt-go/issues/285
	method := &jwt.SigningMethodRSAPSS{
		SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
		Options: &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
		},
	}
	err = method.Verify(strings.Join(parts[:2], "."), parts[2], publicKey)
	require.NoError(t, err, "token verification failed")

	claims := jot.Claims.(*jwt.StandardClaims)
	assert.Equal(t, key.Id, jot.Header["kid"])
	assert.Equal(t, key.GetServiceAccountId(), claims.Issuer)
	assert.Equal(t, "https://iam.api.cloud.yandex.net/iam/v1/tokens", claims.Audience)
	issuedAt := time.Unix(claims.IssuedAt, 0)
	sinceIssued := time.Since(issuedAt)
	assert.True(t, sinceIssued > 0)
	assert.True(t, sinceIssued < time.Minute)
	assert.Equal(t, time.Hour, time.Unix(claims.ExpiresAt, 0).Sub(issuedAt))
}

func TestInstanceServiceAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		const token = "AAAAAAAAAAAAAAAAAAAAAAAA"
		const expiresIn = 43167
		server := httptest.NewServer(http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				_, err := io.WriteString(rw, fmt.Sprintf(`{
				"access_token": %q,
				"expires_in": %v,
				"token_type":"Bearer"
			}`, token, expiresIn))
				assert.NoError(t, err)

			}))
		defer server.Close()
		creds := newInstanceServiceAccountCredentials(server.Listener.Addr().String())
		iamToken, err := creds.IAMToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, token, iamToken.IamToken)
		expectedExpiresAt := time.Now().Add(expiresIn * time.Second)
		actualExpiresAt, err := ptypes.Timestamp(iamToken.ExpiresAt)
		require.NoError(t, err)
		assert.True(t, expectedExpiresAt.After(actualExpiresAt))
		assert.True(t, expectedExpiresAt.Add(-10*time.Second).Before(actualExpiresAt))
	})

	t.Run("internal error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusInternalServerError)
				_, err := io.WriteString(rw, "ERRRORRRRR")
				assert.NoError(t, err)
			}))
		defer server.Close()
		creds := newInstanceServiceAccountCredentials(server.Listener.Addr().String())
		_, err := creds.IAMToken(context.Background())
		require.Error(t, err)
		t.Log(err)
	})
}

func testKey(t *testing.T) *iamkey.Key {
	key, err := iamkey.ReadFromJSONFile("test_data/service_account_key.json")
	require.NoError(t, err)
	return key
}

// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package ycsdk

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	iampb "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
)

const (
	// iamTokenExpiration is refreshAfter time of IAM token.
	// for now it constant, but in near future token expiration will be returned in
	// See https://cloud.yandex.ru/docs/iam/concepts/authorization/iam-token for details.
	iamTokenExpiration = 12 * time.Hour
)

// Credentials is an abstraction of API authorization credentials.
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/authorization for details.
// Note that functions that return Credentials may return different Credentials implementation
// in next SDK version, and this is not considered breaking change.
type Credentials interface {
	// YandexCloudAPICredentials is a marker method. All compatible Credentials implementations have it
	YandexCloudAPICredentials()
}

// ExchangeableCredentials can be exchanged for IAM Token in IAM Token Service, that can be used
// to authorize API calls.
// For now, this is the only option to authorize API calls, but this may be changed in future.
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/iam-token for details.
type ExchangeableCredentials interface {
	Credentials
	// IAMTokenRequest returns request for fresh IAM token or error.
	IAMTokenRequest() (iamTokenReq *iampb.CreateIamTokenRequest, err error)
}

// OAuthToken returns API credentials for user Yandex Passport OAuth token, that can be received
// on page https://oauth.yandex.ru/authorize?response_type=token&client_id=1a6990aa636648e9b2ef855fa7bec2fb
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/oauth-token for details.
func OAuthToken(token string) Credentials {
	return exchangeableCredentialsFunc(func() (*iampb.CreateIamTokenRequest, error) {
		return &iampb.CreateIamTokenRequest{
			Identity: &iampb.CreateIamTokenRequest_YandexPassportOauthToken{
				YandexPassportOauthToken: token,
			},
		}, nil
	})
}

type exchangeableCredentialsFunc func() (iamTokenReq *iampb.CreateIamTokenRequest, err error)

var _ ExchangeableCredentials = (exchangeableCredentialsFunc)(nil)

func (exchangeableCredentialsFunc) YandexCloudAPICredentials() {}

func (f exchangeableCredentialsFunc) IAMTokenRequest() (iamTokenReq *iampb.CreateIamTokenRequest, err error) {
	return f()
}

// ServiceAccountKey returns credentials for the given IAM Key. The key is used to sign JWT tokens.
// JWT tokens are exchanged for IAM Tokens used to authorize API calls.
// This authorization method is not supported for IAM Keys issued for User Accounts.
func ServiceAccountKey(key *iamkey.Key) (Credentials, error) {
	jwtBuilder, err := newServiceAccountJWTBuilder(key)
	if err != nil {
		return nil, err
	}
	return exchangeableCredentialsFunc(func() (*iampb.CreateIamTokenRequest, error) {
		signedJWT, err := jwtBuilder.SignedToken()
		if err != nil {
			return nil, sdkerrors.WithMessage(err, "JWT sign failed")
		}
		return &iampb.CreateIamTokenRequest{
			Identity: &iampb.CreateIamTokenRequest_Jwt{
				Jwt: signedJWT,
			},
		}, nil
	}), nil
}

func newServiceAccountJWTBuilder(key *iamkey.Key) (*serviceAccountJWTBuilder, error) {
	err := validateServiceAccountKey(key)
	if err != nil {
		return nil, sdkerrors.WithMessage(err, "key validation failed")
	}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
	if err != nil {
		return nil, sdkerrors.WithMessage(err, "private key parsing failed")
	}
	return &serviceAccountJWTBuilder{
		key:           key,
		rsaPrivateKey: rsaPrivateKey,
	}, nil
}

func validateServiceAccountKey(key *iamkey.Key) error {
	if key.Id == "" {
		return errors.New("key id is missing")
	}
	if key.GetServiceAccountId() == "" {
		return fmt.Errorf("key should de issued for service account, but subject is %#v", key.Subject)
	}
	return nil
}

type serviceAccountJWTBuilder struct {
	key           *iamkey.Key
	rsaPrivateKey *rsa.PrivateKey
}

func (b *serviceAccountJWTBuilder) SignedToken() (string, error) {
	return b.issueToken().SignedString(b.rsaPrivateKey)
}

func (b *serviceAccountJWTBuilder) issueToken() *jwt.Token {
	issuedAt := time.Now()
	token := jwt.NewWithClaims(jwtSigningMethodPS256WithSaltLengthEqualsHash, jwt.StandardClaims{
		Issuer:    b.key.GetServiceAccountId(),
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: issuedAt.Add(time.Hour).Unix(),
		Audience:  "https://iam.api.cloud.yandex.net/iam/v1/tokens",
	})
	token.Header["kid"] = b.key.Id
	return token
}

// NOTE(skipor): by default, Go RSA PSS uses PSSSaltLengthAuto, which is not accepted by jwt.io and some python libraries.
// Should be removed after https://github.com/dgrijalva/jwt-go/issues/285 fix.
var jwtSigningMethodPS256WithSaltLengthEqualsHash = &jwt.SigningMethodRSAPSS{
	SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
	Options: &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	},
}

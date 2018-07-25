package oauth

import (
	"crypto"
	"crypto/hmac"
	_ "crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	OAUTH_VERSION         = "1.0"
	SIGNATURE_METHOD_HMAC = "HMAC-"

	HTTP_AUTH_HEADER       = "Authorization"
	OAUTH_HEADER           = "OAuth "
	CONSUMER_KEY_PARAM     = "oauth_consumer_key"
	NONCE_PARAM            = "oauth_nonce"
	SIGNATURE_METHOD_PARAM = "oauth_signature_method"
	SIGNATURE_PARAM        = "oauth_signature"
	TIMESTAMP_PARAM        = "oauth_timestamp"
	TOKEN_PARAM            = "oauth_token"
	TOKEN_SECRET_PARAM     = "oauth_token_secret"
	VERSION_PARAM          = "oauth_version"
)

var HASH_METHOD_MAP = map[crypto.Hash]string{
	crypto.SHA1:   "SHA1",
	crypto.SHA256: "SHA256",
}

// Creates a new Consumer instance, with a HMAC-SHA1 signer
func NewConsumer(consumerKey string, consumerSecret string, requestMethod string, requestURL string) *Consumer {
	clock := &defaultClock{}
	consumer := &Consumer{
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		requestMethod:  requestMethod,
		requestURL:     requestURL,
		clock:          clock,
		nonceGenerator: newLockedNonceGenerator(clock),

		AdditionalParams: make(map[string]string),
	}

	consumer.signer = &HMACSigner{
		consumerSecret: consumerSecret,
		hashFunc:       crypto.SHA1,
	}

	return consumer
}

// lockedNonceGenerator wraps a non-reentrant random number generator with alock
type lockedNonceGenerator struct {
	nonceGenerator nonceGenerator
	lock           sync.Mutex
}

func newLockedNonceGenerator(c clock) *lockedNonceGenerator {
	return &lockedNonceGenerator{
		nonceGenerator: rand.New(rand.NewSource(c.Nanos())),
	}
}

func (n *lockedNonceGenerator) Int63() int64 {
	n.lock.Lock()
	r := n.nonceGenerator.Int63()
	n.lock.Unlock()
	return r
}

type clock interface {
	Seconds() int64
	Nanos() int64
}

type nonceGenerator interface {
	Int63() int64
}

type signer interface {
	Sign(message string) (string, error)
	Verify(message string, signature string) error
	SignatureMethod() string
	HashFunc() crypto.Hash
	Debug(enabled bool)
}

type defaultClock struct{}

func (*defaultClock) Seconds() int64 {
	return time.Now().Unix()
}

func (*defaultClock) Nanos() int64 {
	return time.Now().UnixNano()
}

type Consumer struct {
	AdditionalParams map[string]string

	// The rest of this class is configured via the NewConsumer function.
	consumerKey string

	consumerSecret string

	requestMethod string

	requestURL string

	debug bool

	// Private seams for mocking dependencies when testing
	clock clock
	// Seeded generators are not reentrant
	nonceGenerator nonceGenerator
	signer         signer
}

type HMACSigner struct {
	consumerSecret string
	hashFunc       crypto.Hash
	debug          bool
}

func (s *HMACSigner) Debug(enabled bool) {
	s.debug = enabled
}

func (s *HMACSigner) Sign(message string) (string, error) {
	key := escape(s.consumerSecret)
	if s.debug {
		fmt.Println("Signing:", message)
		fmt.Println("Key:", key)
	}

	h := hmac.New(s.HashFunc().New, []byte(key+"&"))
	h.Write([]byte(message))
	rawSignature := h.Sum(nil)

	base64signature := base64.StdEncoding.EncodeToString(rawSignature)
	if s.debug {
		fmt.Println("Base64 signature:", base64signature)
	}
	return base64signature, nil
}

func (s *HMACSigner) Verify(message string, signature string) error {
	if s.debug {
		fmt.Println("Verifying Base64 signature:", signature)
	}
	validSignature, err := s.Sign(message)
	if err != nil {
		return err
	}

	if validSignature != signature {
		decodedSigniture, _ := url.QueryUnescape(signature)
		if validSignature != decodedSigniture {
			return fmt.Errorf("signature did not match")
		}
	}

	return nil
}

func (s *HMACSigner) SignatureMethod() string {
	return SIGNATURE_METHOD_HMAC + HASH_METHOD_MAP[s.HashFunc()]
}

func (s *HMACSigner) HashFunc() crypto.Hash {
	return s.hashFunc
}

func escape(s string) string {
	t := make([]byte, 0, 3*len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isEscapable(c) {
			t = append(t, '%')
			t = append(t, "0123456789ABCDEF"[c>>4])
			t = append(t, "0123456789ABCDEF"[c&15])
		} else {
			t = append(t, s[i])
		}
	}
	return string(t)
}

func isEscapable(b byte) bool {
	return !('A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || '0' <= b && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~')
}

func (c *Consumer) Debug(enabled bool) {
	c.debug = enabled
	c.signer.Debug(enabled)
}

func (c *Consumer) GetRequestUrl() (loginUrl string, err error) {
	if os.Getenv("GO_ENV") == "development" {
		c.AdditionalParams["approachKey"] = c.consumerKey
		c.AdditionalParams["secretKey"] = c.consumerSecret
	}
	c.AdditionalParams["responseFormatType"] = "xml"

	params := c.baseParams(c.consumerKey, c.AdditionalParams)

	if c.debug {
		fmt.Println("params:", params)
	}

	req := &request{
		method:      c.requestMethod,
		url:         c.requestURL,
		oauthParams: params,
	}

	signature, err := c.signRequest(req)
	if err != nil {
		return "", err
	}

	result := req.url + "?"
	for pos, key := range req.oauthParams.Keys() {
		for innerPos, value := range req.oauthParams.Get(key) {
			if pos+innerPos != 0 {
				result += "&"
			}
			result += fmt.Sprintf("%s=%s", key, value)
		}
	}

	result += fmt.Sprintf("&%s=%s", SIGNATURE_PARAM, escape(signature))

	if c.debug {
		fmt.Println("req: ", result)
	}

	return result, nil
}

func (c *Consumer) baseParams(consumerKey string, additionalParams map[string]string) *OrderedParams {
	params := NewOrderedParams()
	params.Add(VERSION_PARAM, OAUTH_VERSION)
	params.Add(SIGNATURE_METHOD_PARAM, c.signer.SignatureMethod())
	params.Add(TIMESTAMP_PARAM, strconv.FormatInt(c.clock.Seconds(), 10))
	params.Add(NONCE_PARAM, strconv.FormatInt(c.nonceGenerator.Int63(), 10))
	params.Add(CONSUMER_KEY_PARAM, consumerKey)

	for key, value := range additionalParams {
		params.Add(key, value)
	}

	return params
}

func (c *Consumer) signRequest(req *request) (string, error) {
	baseString := c.requestString(req.method, req.url, req.oauthParams)

	if c.debug {
		fmt.Println("baseString: ", baseString)
	}

	signature, err := c.signer.Sign(baseString)
	if err != nil {
		return "", err
	}

	return signature, nil
}

func (c *Consumer) requestString(method string, url string, params *OrderedParams) string {
	result := method + "&" + escape(url)
	for pos, key := range params.Keys() {
		for innerPos, value := range params.Get(key) {
			if pos+innerPos == 0 {
				result += "&"
			} else {
				result += escape("&")
			}
			result += escape(fmt.Sprintf("%s=%s", key, value))
		}
	}
	return result
}

type request struct {
	method      string
	url         string
	oauthParams *OrderedParams
	userParams  map[string]string
}

//
// String Sorting helpers
//

type ByValue []string

func (a ByValue) Len() int {
	return len(a)
}

func (a ByValue) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByValue) Less(i, j int) bool {
	return a[i] < a[j]
}

//
// ORDERED PARAMS
//

type OrderedParams struct {
	allParams   map[string][]string
	keyOrdering []string
}

func NewOrderedParams() *OrderedParams {
	return &OrderedParams{
		allParams:   make(map[string][]string),
		keyOrdering: make([]string, 0),
	}
}

func (o *OrderedParams) Get(key string) []string {
	sort.Sort(ByValue(o.allParams[key]))
	return o.allParams[key]
}

func (o *OrderedParams) Keys() []string {
	sort.Sort(o)
	return o.keyOrdering
}

func (o *OrderedParams) Add(key, value string) {
	o.AddUnescaped(key, escape(value))
}

func (o *OrderedParams) AddUnescaped(key, value string) {
	if _, exists := o.allParams[key]; !exists {
		o.keyOrdering = append(o.keyOrdering, key)
		o.allParams[key] = make([]string, 1)
		o.allParams[key][0] = value
	} else {
		o.allParams[key] = append(o.allParams[key], value)
	}
}

func (o *OrderedParams) Len() int {
	return len(o.keyOrdering)
}

func (o *OrderedParams) Less(i int, j int) bool {
	return o.keyOrdering[i] < o.keyOrdering[j]
}

func (o *OrderedParams) Swap(i int, j int) {
	o.keyOrdering[i], o.keyOrdering[j] = o.keyOrdering[j], o.keyOrdering[i]
}

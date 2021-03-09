package httpmock

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/jarcoal/httpmock/internal"
)

const regexpPrefix = "=~"

// NoResponderFound is returned when no responders are found for a
// given HTTP method and URL.
var NoResponderFound = errors.New("no responder found") // nolint: golint

// ConnectionFailure is a responder that returns a connection failure.
// This is the default responder, and is called when no other matching
// responder is found.
func ConnectionFailure(*http.Request) (*http.Response, error) {
	return nil, NoResponderFound
}

// NewMockTransport creates a new *MockTransport with no responders.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		responders:    make(map[internal.RouteKey]Responder),
		callCountInfo: make(map[internal.RouteKey]int),
	}
}

type regexpResponder struct {
	origRx    string
	method    string
	rx        *regexp.Regexp
	responder Responder
}

// MockTransport implements http.RoundTripper, which fulfills single
// http requests issued by an http.Client.  This implementation
// doesn't actually make the call, instead deferring to the registered
// list of responders.
type MockTransport struct {
	mu               sync.RWMutex
	responders       map[internal.RouteKey]Responder
	regexpResponders []regexpResponder
	noResponder      Responder
	callCountInfo    map[internal.RouteKey]int
	totalCallCount   int
}

// RoundTrip receives HTTP requests and routes them to the appropriate
// responder.  It is required to implement the http.RoundTripper
// interface.  You will not interact with this directly, instead the
// *http.Client you are using will call it for you.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()

	method := req.Method
	if method == "" {
		// http.Request.Method is documented to default to GET:
		method = http.MethodGet
	}

	var (
		responder  Responder
		respKey    internal.RouteKey
		submatches []string
	)
	key := internal.RouteKey{
		Method: method,
	}
	for _, getResponder := range []func(internal.RouteKey) (Responder, internal.RouteKey, []string){
		m.responderForKey,       // Exact match
		m.regexpResponderForKey, // Regexp match
	} {
		// try and get a responder that matches the method and URL with
		// query params untouched: http://z.tld/path?q...
		key.URL = url
		responder, respKey, submatches = getResponder(key)
		if responder != nil {
			break
		}

		// if we weren't able to find a responder, try with the URL *and*
		// sorted query params
		query := sortedQuery(req.URL.Query())
		if query != "" {
			// Replace unsorted query params by sorted ones:
			//   http://z.tld/path?sorted_q...
			key.URL = strings.Replace(url, req.URL.RawQuery, query, 1)
			responder, respKey, submatches = getResponder(key)
			if responder != nil {
				break
			}
		}

		// if we weren't able to find a responder, try without any query params
		strippedURL := *req.URL
		strippedURL.RawQuery = ""
		strippedURL.Fragment = ""

		// go1.6 does not handle URL.ForceQuery, so in case it is set in go>1.6,
		// remove the "?" manually if present.
		surl := strings.TrimSuffix(strippedURL.String(), "?")

		hasQueryString := url != surl

		// if the URL contains a querystring then we strip off the
		// querystring and try again: http://z.tld/path
		if hasQueryString {
			key.URL = surl
			responder, respKey, submatches = getResponder(key)
			if responder != nil {
				break
			}
		}

		// if we weren't able to find a responder for the full URL, try with
		// the path part only
		pathAlone := req.URL.Path

		// First with unsorted querystring: /path?q...
		if hasQueryString {
			key.URL = pathAlone + strings.TrimPrefix(url, surl) // concat after-path part
			responder, respKey, submatches = getResponder(key)
			if responder != nil {
				break
			}

			// Then with sorted querystring: /path?sorted_q...
			key.URL = pathAlone + "?" + sortedQuery(req.URL.Query())
			if req.URL.Fragment != "" {
				key.URL += "#" + req.URL.Fragment
			}
			responder, respKey, submatches = getResponder(key)
			if responder != nil {
				break
			}
		}

		// Then using path alone: /path
		key.URL = pathAlone
		responder, respKey, submatches = getResponder(key)
		if responder != nil {
			break
		}
	}

	m.mu.Lock()
	// if we found a responder, call it
	if responder != nil {
		m.callCountInfo[key]++
		if key != respKey {
			m.callCountInfo[respKey]++
		}
		m.totalCallCount++
	} else if m.noResponder != nil {
		// we didn't find a responder, so fire the 'no responder' responder
		m.callCountInfo[internal.NoResponder]++
		m.totalCallCount++
		responder = m.noResponder
	}
	m.mu.Unlock()

	if responder == nil {
		return ConnectionFailure(req)
	}
	return runCancelable(responder, internal.SetSubmatches(req, submatches))
}

// NumResponders returns the number of responders currently in use.
func (m *MockTransport) NumResponders() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.responders) + len(m.regexpResponders)
}

func runCancelable(responder Responder, req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if req.Cancel == nil && ctx.Done() == nil { // nolint: staticcheck
		resp, err := responder(req)
		return resp, internal.CheckStackTracer(req, err)
	}

	// Set up a goroutine that translates a close(req.Cancel) into a
	// "request canceled" error, and another one that runs the
	// responder. Then race them: first to the result channel wins.

	type result struct {
		response *http.Response
		err      error
	}
	resultch := make(chan result, 1)
	done := make(chan struct{}, 1)

	go func() {
		select {
		case <-req.Cancel: // nolint: staticcheck
			resultch <- result{
				response: nil,
				err:      errors.New("request canceled"),
			}
		case <-ctx.Done():
			resultch <- result{
				response: nil,
				err:      ctx.Err(),
			}
		case <-done:
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				resultch <- result{
					response: nil,
					err:      fmt.Errorf("panic in responder: got %q", err),
				}
			}
		}()

		response, err := responder(req)
		resultch <- result{
			response: response,
			err:      err,
		}
	}()

	r := <-resultch

	// if a cancel() issued from context.WithCancel() or a
	// close(req.Cancel) are never coming, we'll need to unblock the
	// first goroutine.
	done <- struct{}{}

	return r.response, internal.CheckStackTracer(req, r.err)
}

// responderForKey returns a responder for a given key.
func (m *MockTransport) responderForKey(key internal.RouteKey) (Responder, internal.RouteKey, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.responders[key], key, nil
}

// responderForKeyUsingRegexp returns the first responder matching a
// given key using regexps.
func (m *MockTransport) regexpResponderForKey(key internal.RouteKey) (Responder, internal.RouteKey, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, regInfo := range m.regexpResponders {
		if regInfo.method == key.Method {
			if sm := regInfo.rx.FindStringSubmatch(key.URL); sm != nil {
				if len(sm) == 1 {
					sm = nil
				} else {
					sm = sm[1:]
				}
				return regInfo.responder, internal.RouteKey{
					Method: key.Method,
					URL:    regInfo.origRx,
				}, sm
			}
		}
	}
	return nil, key, nil
}

func isRegexpURL(url string) bool {
	return strings.HasPrefix(url, regexpPrefix)
}

// RegisterResponder adds a new responder, associated with a given
// HTTP method and URL (or path).
//
// When a request comes in that matches, the responder is called and
// the response returned to the client.
//
// If url contains query parameters, their order matters as well as
// their content. All following URLs are here considered as different:
//   http://z.tld?a=1&b=1
//   http://z.tld?b=1&a=1
//   http://z.tld?a&b
//   http://z.tld?a=&b=
//
// If url begins with "=~", the following chars are considered as a
// regular expression. If this regexp can not be compiled, it panics.
// Note that the "=~" prefix remains in statistics returned by
// GetCallCountInfo(). As 2 regexps can match the same URL, the regexp
// responders are tested in the order they are registered. Registering
// an already existing regexp responder (same method & same regexp
// string) replaces its responder but does not change its position.
//
// See RegisterRegexpResponder() to directly pass a *regexp.Regexp.
//
// Example:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.DeactivateAndReset()
//
//     httpmock.RegisterResponder("GET", "http://example.com/",
//       httpmock.NewStringResponder(200, "hello world"))
//
//     httpmock.RegisterResponder("GET", "/path/only",
//       httpmock.NewStringResponder("any host hello world", 200))
//
//     httpmock.RegisterResponder("GET", `=~^/item/id/\d+\z`,
//       httpmock.NewStringResponder("any item get", 200))
//
//     // requests to http://example.com/ now return "hello world" and
//     // requests to any host with path /path/only return "any host hello world"
//     // requests to any host with path matching ^/item/id/\d+\z regular expression return "any item get"
//   }
func (m *MockTransport) RegisterResponder(method, url string, responder Responder) {
	if isRegexpURL(url) {
		m.registerRegexpResponder(regexpResponder{
			origRx:    url,
			method:    method,
			rx:        regexp.MustCompile(url[2:]),
			responder: responder,
		})
		return
	}

	key := internal.RouteKey{
		Method: method,
		URL:    url,
	}

	m.mu.Lock()
	m.responders[key] = responder
	m.callCountInfo[key] = 0
	m.mu.Unlock()
}

func (m *MockTransport) registerRegexpResponder(regexpResponder regexpResponder) {
	m.mu.Lock()
	defer m.mu.Unlock()

found:
	for {
		for i, rr := range m.regexpResponders {
			if rr.method == regexpResponder.method && rr.origRx == regexpResponder.origRx {
				m.regexpResponders[i] = regexpResponder
				break found
			}
		}
		m.regexpResponders = append(m.regexpResponders, regexpResponder)
		break // nolint: staticcheck
	}

	m.callCountInfo[internal.RouteKey{
		Method: regexpResponder.method,
		URL:    regexpResponder.origRx,
	}] = 0
}

// RegisterRegexpResponder adds a new responder, associated with a given
// HTTP method and URL (or path) regular expression.
//
// When a request comes in that matches, the responder is called and
// the response returned to the client.
//
// As 2 regexps can match the same URL, the regexp responders are
// tested in the order they are registered. Registering an already
// existing regexp responder (same method & same regexp string)
// replaces its responder but does not change its position.
//
// A "=~" prefix is added to the stringified regexp in the statistics
// returned by GetCallCountInfo().
//
// See RegisterResponder function and the "=~" prefix in its url
// parameter to avoid compiling the regexp by yourself.
func (m *MockTransport) RegisterRegexpResponder(method string, urlRegexp *regexp.Regexp, responder Responder) {
	m.registerRegexpResponder(regexpResponder{
		origRx:    regexpPrefix + urlRegexp.String(),
		method:    method,
		rx:        urlRegexp,
		responder: responder,
	})
}

// RegisterResponderWithQuery is same as RegisterResponder, but it
// doesn't depend on query items order.
//
// If query is non-nil, its type can be:
//   url.Values
//   map[string]string
//   string, a query string like "a=12&a=13&b=z&c" (see net/url.ParseQuery function)
//
// If the query type is not recognized or the string cannot be parsed
// using net/url.ParseQuery, a panic() occurs.
//
// Unlike RegisterResponder, path cannot be prefixed by "=~" to say it
// is a regexp. If it is, a panic occurs.
func (m *MockTransport) RegisterResponderWithQuery(method, path string, query interface{}, responder Responder) {
	if isRegexpURL(path) {
		panic(`path begins with "=~", RegisterResponder should be used instead of RegisterResponderWithQuery`)
	}

	var mapQuery url.Values
	switch q := query.(type) {
	case url.Values:
		mapQuery = q

	case map[string]string:
		mapQuery = make(url.Values, len(q))
		for key, e := range q {
			mapQuery[key] = []string{e}
		}

	case string:
		var err error
		mapQuery, err = url.ParseQuery(q)
		if err != nil {
			panic("RegisterResponderWithQuery bad query string: " + err.Error())
		}

	default:
		if query != nil {
			panic(fmt.Sprintf("RegisterResponderWithQuery bad query type %T. Only url.Values, map[string]string and string are allowed", query))
		}
	}

	if queryString := sortedQuery(mapQuery); queryString != "" {
		path += "?" + queryString
	}
	m.RegisterResponder(method, path, responder)
}

func sortedQuery(m url.Values) string {
	if len(m) == 0 {
		return ""
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b bytes.Buffer
	var values []string // nolint: prealloc

	for _, k := range keys {
		// Do not alter the passed url.Values
		values = append(values, m[k]...)
		sort.Strings(values)

		k = url.QueryEscape(k)

		for _, v := range values {
			if b.Len() > 0 {
				b.WriteByte('&')
			}
			fmt.Fprintf(&b, "%v=%v", k, url.QueryEscape(v))
		}

		values = values[:0]
	}

	return b.String()
}

// RegisterNoResponder is used to register a responder that is called
// if no other responder is found.  The default is httpmock.ConnectionFailure.
func (m *MockTransport) RegisterNoResponder(responder Responder) {
	m.mu.Lock()
	m.noResponder = responder
	m.mu.Unlock()
}

// Reset removes all registered responders (including the no
// responder) from the MockTransport. It zeroes call counters too.
func (m *MockTransport) Reset() {
	m.mu.Lock()
	m.responders = make(map[internal.RouteKey]Responder)
	m.regexpResponders = nil
	m.noResponder = nil
	m.callCountInfo = make(map[internal.RouteKey]int)
	m.totalCallCount = 0
	m.mu.Unlock()
}

// ZeroCallCounters zeroes call counters without touching registered responders.
func (m *MockTransport) ZeroCallCounters() {
	m.mu.Lock()
	for k := range m.callCountInfo {
		m.callCountInfo[k] = 0
	}
	m.totalCallCount = 0
	m.mu.Unlock()
}

// GetCallCountInfo gets the info on all the calls httpmock has caught
// since it was activated or reset. The info is returned as a map of
// the calling keys with the number of calls made to them as their
// value. The key is the method, a space, and the url all concatenated
// together.
//
// As a special case, regexp responders generate 2 entries for each
// call. One for the call caught and the other for the rule that
// matched. For example:
//   RegisterResponder("GET", `=~z\.com\z`, NewStringResponder(200, "body"))
//   http.Get("http://z.com")
//
// will generate the following result:
//   map[string]int{
//     `GET http://z.com`: 1,
//     `GET =~z\.com\z`:   1,
//   }
func (m *MockTransport) GetCallCountInfo() map[string]int {
	m.mu.RLock()
	res := make(map[string]int, len(m.callCountInfo))
	for k, v := range m.callCountInfo {
		res[k.String()] = v
	}
	m.mu.RUnlock()
	return res
}

// GetTotalCallCount returns the totalCallCount.
func (m *MockTransport) GetTotalCallCount() int {
	m.mu.RLock()
	count := m.totalCallCount
	m.mu.RUnlock()
	return count
}

// DefaultTransport is the default mock transport used by Activate,
// Deactivate, Reset, DeactivateAndReset, RegisterResponder, and
// RegisterNoResponder.
var DefaultTransport = NewMockTransport()

// InitialTransport is a cache of the original transport used so we
// can put it back when Deactivate is called.
var InitialTransport = http.DefaultTransport

// oldClients is used to handle custom http clients (i.e clients other
// than http.DefaultClient).
var oldClients = map[*http.Client]http.RoundTripper{}

// Activate starts the mock environment.  This should be called before
// your tests run.  Under the hood this replaces the Transport on the
// http.DefaultClient with httpmock.DefaultTransport.
//
// To enable mocks for a test, simply activate at the beginning of a test:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     // all http requests using http.DefaultTransport will now be intercepted
//   }
//
// If you want all of your tests in a package to be mocked, just call
// Activate from init():
//   func init() {
//     httpmock.Activate()
//   }
//
// or using a TestMain function:
//   func TestMain(m *testing.M) {
//     httpmock.Activate()
//     os.Exit(m.Run())
//   }
func Activate() {
	if Disabled() {
		return
	}

	// make sure that if Activate is called multiple times it doesn't
	// overwrite the InitialTransport with a mock transport.
	if http.DefaultTransport != DefaultTransport {
		InitialTransport = http.DefaultTransport
	}

	http.DefaultTransport = DefaultTransport
}

// ActivateNonDefault starts the mock environment with a non-default
// http.Client.  This emulates the Activate function, but allows for
// custom clients that do not use http.DefaultTransport
//
// To enable mocks for a test using a custom client, activate at the
// beginning of a test:
//   client := &http.Client{Transport: &http.Transport{TLSHandshakeTimeout: 60 * time.Second}}
//   httpmock.ActivateNonDefault(client)
func ActivateNonDefault(client *http.Client) {
	if Disabled() {
		return
	}

	// save the custom client & it's RoundTripper
	if _, ok := oldClients[client]; !ok {
		oldClients[client] = client.Transport
	}
	client.Transport = DefaultTransport
}

// GetCallCountInfo gets the info on all the calls httpmock has caught
// since it was activated or reset. The info is returned as a map of
// the calling keys with the number of calls made to them as their
// value. The key is the method, a space, and the url all concatenated
// together.
//
// As a special case, regexp responders generate 2 entries for each
// call. One for the call caught and the other for the rule that
// matched. For example:
//   RegisterResponder("GET", `=~z\.com\z`, NewStringResponder(200, "body"))
//   http.Get("http://z.com")
//
// will generate the following result:
//   map[string]int{
//     `GET http://z.com`: 1,
//     `GET =~z\.com\z`:   1,
//   }
func GetCallCountInfo() map[string]int {
	return DefaultTransport.GetCallCountInfo()
}

// GetTotalCallCount gets the total number of calls httpmock has taken
// since it was activated or reset.
func GetTotalCallCount() int {
	return DefaultTransport.GetTotalCallCount()
}

// Deactivate shuts down the mock environment.  Any HTTP calls made
// after this will use a live transport.
//
// Usually you'll call it in a defer right after activating the mock
// environment:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.Deactivate()
//
//     // when this test ends, the mock environment will close
//   }
//
// Since go 1.14 you can also use (*testing.T).Cleanup() method as in:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     t.Cleanup(httpmock.Deactivate)
//
//     // when this test ends, the mock environment will close
//   }
//
// useful in test helpers to save your callers from calling defer themselves.
func Deactivate() {
	if Disabled() {
		return
	}
	http.DefaultTransport = InitialTransport

	// reset the custom clients to use their original RoundTripper
	for oldClient, oldTransport := range oldClients {
		oldClient.Transport = oldTransport
		delete(oldClients, oldClient)
	}
}

// Reset removes any registered mocks and returns the mock
// environment to its initial state. It zeroes call counters too.
func Reset() {
	DefaultTransport.Reset()
}

// ZeroCallCounters zeroes call counters without touching registered responders.
func ZeroCallCounters() {
	DefaultTransport.ZeroCallCounters()
}

// DeactivateAndReset is just a convenience method for calling
// Deactivate() and then Reset().
//
// Happy deferring!
func DeactivateAndReset() {
	Deactivate()
	Reset()
}

// RegisterResponder adds a new responder, associated with a given
// HTTP method and URL (or path).
//
// When a request comes in that matches, the responder is called and
// the response returned to the client.
//
// If url contains query parameters, their order matters as well as
// their content. All following URLs are here considered as different:
//   http://z.tld?a=1&b=1
//   http://z.tld?b=1&a=1
//   http://z.tld?a&b
//   http://z.tld?a=&b=
//
// If url begins with "=~", the following chars are considered as a
// regular expression. If this regexp can not be compiled, it panics.
// Note that the "=~" prefix remains in statistics returned by
// GetCallCountInfo(). As 2 regexps can match the same URL, the regexp
// responders are tested in the order they are registered. Registering
// an already existing regexp responder (same method & same regexp
// string) replaces its responder but does not change its position.
//
// See RegisterRegexpResponder() to directly pass a *regexp.Regexp.
//
// Example:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.DeactivateAndReset()
//
//     httpmock.RegisterResponder("GET", "http://example.com/",
//       httpmock.NewStringResponder(200, "hello world"))
//
//     httpmock.RegisterResponder("GET", "/path/only",
//       httpmock.NewStringResponder("any host hello world", 200))
//
//     httpmock.RegisterResponder("GET", `=~^/item/id/\d+\z`,
//       httpmock.NewStringResponder("any item get", 200))
//
//     // requests to http://example.com/ now return "hello world" and
//     // requests to any host with path /path/only return "any host hello world"
//     // requests to any host with path matching ^/item/id/\d+\z regular expression return "any item get"
//   }
func RegisterResponder(method, url string, responder Responder) {
	DefaultTransport.RegisterResponder(method, url, responder)
}

// RegisterRegexpResponder adds a new responder, associated with a given
// HTTP method and URL (or path) regular expression.
//
// When a request comes in that matches, the responder is called and
// the response returned to the client.
//
// As 2 regexps can match the same URL, the regexp responders are
// tested in the order they are registered. Registering an already
// existing regexp responder (same method & same regexp string)
// replaces its responder but does not change its position.
//
// A "=~" prefix is added to the stringified regexp in the statistics
// returned by GetCallCountInfo().
//
// See RegisterResponder function and the "=~" prefix in its url
// parameter to avoid compiling the regexp by yourself.
func RegisterRegexpResponder(method string, urlRegexp *regexp.Regexp, responder Responder) {
	DefaultTransport.RegisterRegexpResponder(method, urlRegexp, responder)
}

// RegisterResponderWithQuery it is same as RegisterResponder, but
// doesn't depends on query items order.
//
// query type can be:
//   url.Values
//   map[string]string
//   string, a query string like "a=12&a=13&b=z&c" (see net/url.ParseQuery function)
//
// If the query type is not recognized or the string cannot be parsed
// using net/url.ParseQuery, a panic() occurs.
//
// Example using a net/url.Values:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.DeactivateAndReset()
//
//     expectedQuery := net.Values{
//       "a": []string{"3", "1", "8"},
//       "b": []string{"4", "2"},
//     }
//     httpmock.RegisterResponderWithQueryValues(
//       "GET", "http://example.com/", expectedQuery,
//       httpmock.NewStringResponder("hello world", 200))
//
//     // requests to http://example.com?a=1&a=3&a=8&b=2&b=4
//     //      and to http://example.com?b=4&a=2&b=2&a=8&a=1
//     // now return 'hello world'
//   }
//
// or using a map[string]string:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.DeactivateAndReset()
//
//     expectedQuery := map[string]string{
//       "a": "1",
//       "b": "2"
//     }
//     httpmock.RegisterResponderWithQuery(
//       "GET", "http://example.com/", expectedQuery,
//       httpmock.NewStringResponder("hello world", 200))
//
//     // requests to http://example.com?a=1&b=2 and http://example.com?b=2&a=1 now return 'hello world'
//   }
//
// or using a query string:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.DeactivateAndReset()
//
//     expectedQuery := "a=3&b=4&b=2&a=1&a=8"
//     httpmock.RegisterResponderWithQueryValues(
//       "GET", "http://example.com/", expectedQuery,
//       httpmock.NewStringResponder("hello world", 200))
//
//     // requests to http://example.com?a=1&a=3&a=8&b=2&b=4
//     //      and to http://example.com?b=4&a=2&b=2&a=8&a=1
//     // now return 'hello world'
//   }
func RegisterResponderWithQuery(method, path string, query interface{}, responder Responder) {
	DefaultTransport.RegisterResponderWithQuery(method, path, query, responder)
}

// RegisterNoResponder adds a mock that is called whenever a request
// for an unregistered URL is received.  The default behavior is to
// return a connection error.
//
// In some cases you may not want all URLs to be mocked, in which case
// you can do this:
//   func TestFetchArticles(t *testing.T) {
//     httpmock.Activate()
//     defer httpmock.DeactivateAndReset()
//     httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
//
//     // any requests that don't have a registered URL will be fetched normally
//   }
func RegisterNoResponder(responder Responder) {
	DefaultTransport.RegisterNoResponder(responder)
}

// ErrSubmatchNotFound is the error returned by GetSubmatch* functions
// when the given submatch index cannot be found.
var ErrSubmatchNotFound = errors.New("submatch not found")

// GetSubmatch has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as a
// string. Example:
//   RegisterResponder("GET", `=~^/item/name/([^/]+)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       name, err := GetSubmatch(req, 1) // 1=first regexp submatch
//       if err != nil {
//         return nil, err
//       }
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":   123,
//         "name": name,
//       })
//     })
//
// It panics if n < 1. See MustGetSubmatch to avoid testing the
// returned error.
func GetSubmatch(req *http.Request, n int) (string, error) {
	if n <= 0 {
		panic(fmt.Sprintf("getting submatches starts at 1, not %d", n))
	}
	n--

	submatches := internal.GetSubmatches(req)
	if n >= len(submatches) {
		return "", ErrSubmatchNotFound
	}
	return submatches[n], nil
}

// GetSubmatchAsInt has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as an
// int64. Example:
//   RegisterResponder("GET", `=~^/item/id/(\d+)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       id, err := GetSubmatchAsInt(req, 1) // 1=first regexp submatch
//       if err != nil {
//         return nil, err
//       }
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":   id,
//         "name": "The beautiful name",
//       })
//     })
//
// It panics if n < 1. See MustGetSubmatchAsInt to avoid testing the
// returned error.
func GetSubmatchAsInt(req *http.Request, n int) (int64, error) {
	sm, err := GetSubmatch(req, n)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(sm, 10, 64)
}

// GetSubmatchAsUint has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as a
// uint64. Example:
//   RegisterResponder("GET", `=~^/item/id/(\d+)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       id, err := GetSubmatchAsUint(req, 1) // 1=first regexp submatch
//       if err != nil {
//         return nil, err
//       }
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":   id,
//         "name": "The beautiful name",
//       })
//     })
//
// It panics if n < 1. See MustGetSubmatchAsUint to avoid testing the
// returned error.
func GetSubmatchAsUint(req *http.Request, n int) (uint64, error) {
	sm, err := GetSubmatch(req, n)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(sm, 10, 64)
}

// GetSubmatchAsFloat has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as a
// float64. Example:
//   RegisterResponder("PATCH", `=~^/item/id/\d+\?height=(\d+(?:\.\d*)?)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       height, err := GetSubmatchAsFloat(req, 1) // 1=first regexp submatch
//       if err != nil {
//         return nil, err
//       }
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":     id,
//         "name":   "The beautiful name",
//         "height": height,
//       })
//     })
//
// It panics if n < 1. See MustGetSubmatchAsFloat to avoid testing the
// returned error.
func GetSubmatchAsFloat(req *http.Request, n int) (float64, error) {
	sm, err := GetSubmatch(req, n)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(sm, 64)
}

// MustGetSubmatch works as GetSubmatch except that it panics in case
// of error (submatch not found). It has to be used in Responders
// installed by RegisterRegexpResponder or RegisterResponder + "=~"
// URL prefix. It allows to retrieve the n-th submatch of the matching
// regexp, as a string. Example:
//   RegisterResponder("GET", `=~^/item/name/([^/]+)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       name := MustGetSubmatch(req, 1) // 1=first regexp submatch
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":   123,
//         "name": name,
//       })
//     })
//
// It panics if n < 1.
func MustGetSubmatch(req *http.Request, n int) string {
	s, err := GetSubmatch(req, n)
	if err != nil {
		panic("GetSubmatch failed: " + err.Error())
	}
	return s
}

// MustGetSubmatchAsInt works as GetSubmatchAsInt except that it
// panics in case of error (submatch not found or invalid int64
// format). It has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as an
// int64. Example:
//   RegisterResponder("GET", `=~^/item/id/(\d+)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       id := MustGetSubmatchAsInt(req, 1) // 1=first regexp submatch
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":   id,
//         "name": "The beautiful name",
//       })
//     })
//
// It panics if n < 1.
func MustGetSubmatchAsInt(req *http.Request, n int) int64 {
	i, err := GetSubmatchAsInt(req, n)
	if err != nil {
		panic("GetSubmatchAsInt failed: " + err.Error())
	}
	return i
}

// MustGetSubmatchAsUint works as GetSubmatchAsUint except that it
// panics in case of error (submatch not found or invalid uint64
// format). It has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as a
// uint64. Example:
//   RegisterResponder("GET", `=~^/item/id/(\d+)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       id, err := MustGetSubmatchAsUint(req, 1) // 1=first regexp submatch
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":   id,
//         "name": "The beautiful name",
//       })
//     })
//
// It panics if n < 1.
func MustGetSubmatchAsUint(req *http.Request, n int) uint64 {
	u, err := GetSubmatchAsUint(req, n)
	if err != nil {
		panic("GetSubmatchAsUint failed: " + err.Error())
	}
	return u
}

// MustGetSubmatchAsFloat works as GetSubmatchAsFloat except that it
// panics in case of error (submatch not found or invalid float64
// format). It has to be used in Responders installed by
// RegisterRegexpResponder or RegisterResponder + "=~" URL prefix. It
// allows to retrieve the n-th submatch of the matching regexp, as a
// float64. Example:
//   RegisterResponder("PATCH", `=~^/item/id/\d+\?height=(\d+(?:\.\d*)?)\z`,
//     func(req *http.Request) (*http.Response, error) {
//       height := MustGetSubmatchAsFloat(req, 1) // 1=first regexp submatch
//       return NewJsonResponse(200, map[string]interface{}{
//         "id":     id,
//         "name":   "The beautiful name",
//         "height": height,
//       })
//     })
//
// It panics if n < 1.
func MustGetSubmatchAsFloat(req *http.Request, n int) float64 {
	f, err := GetSubmatchAsFloat(req, n)
	if err != nil {
		panic("GetSubmatchAsFloat failed: " + err.Error())
	}
	return f
}

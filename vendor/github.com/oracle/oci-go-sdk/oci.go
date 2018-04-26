
/*
This is the official Go SDK for Oracle Cloud Infrastructure

Installation

Refer to https://github.com/oracle/oci-go-sdk/blob/master/README.md#installing for installation instructions.

Configuration

Refer to https://github.com/oracle/oci-go-sdk/blob/master/README.md#configuring for configuration instructions.

Quickstart

The following example shows how to get started with the SDK. The example belows creates an identityClient
struct with the default configuration. It then utilizes the identityClient to list availability domains and prints
them out to stdout

	import (
		"context"
		"fmt"

		"github.com/oracle/oci-go-sdk/common"
		"github.com/oracle/oci-go-sdk/identity"
	)

	func main() {
		c, err := identity.NewIdentityClientWithConfigurationProvider(common.DefaultConfigProvider())
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// The OCID of the tenancy containing the compartment.
		tenancyID, err := common.DefaultConfigProvider().TenancyOCID()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		request := identity.ListAvailabilityDomainsRequest{
			CompartmentId: &tenancyID,
		}

		r, err := c.ListAvailabilityDomains(context.Background(), request)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("List of available domains: %v", r.Items)
		return
	}

More examples can be found in the SDK Github repo: https://github.com/oracle/oci-go-sdk/tree/master/example

Optional fields in the SDK

Optional fields are represented with the `mandatory:"false"` tag on input structs. The SDK will omit all optional fields that are nil when making requests.
In the case of enum-type fields, the SDK will omit fields whose value is an empty string.

Helper functions

The SDK uses pointers for primitive types in many input structs. To aid in the construction of such structs, the SDK provides
functions that return a pointer for a given value. For example:

	// Given the struct
	type CreateVcnDetails struct {

		// Example: `172.16.0.0/16`
		CidrBlock *string `mandatory:"true" json:"cidrBlock"`

		CompartmentId *string `mandatory:"true" json:"compartmentId"`

		DisplayName *string `mandatory:"false" json:"displayName"`

	}

	// We can use the helper functions to build the struct
	details := core.CreateVcnDetails{
		CidrBlock:     common.String("172.16.0.0/16"),
		CompartmentId: common.String("someOcid"),
		DisplayName:   common.String("myVcn"),
	}


Signing custom requests

The SDK exposes a stand-alone signer that can be used to signing custom requests. Related code can be found here:
https://github.com/oracle/oci-go-sdk/blob/master/common/http_signer.go.

The example below shows how to create a default signer.

	client := http.Client{}
	var request http.Request
	request = ... // some custom request

	// Set the Date header
	request.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	// And a provider of cryptographic keys
	provider := common.DefaultConfigProvider()

	// Build the signer
	signer := common.DefaultSigner(provider)

	// Sign the request
	signer.Sign(&request)

	// Execute the request
	client.Do(request)



The signer also allows more granular control on the headers used for signing. For example:

	client := http.Client{}
	var request http.Request
	request = ... // some custom request

	// Set the Date header
	request.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	// Mandatory headers to be used in the sign process
	defaultGenericHeaders    = []string{"date", "(request-target)", "host"}

	// Optional headers
	optionalHeaders = []string{"content-length", "content-type", "x-content-sha256"}

	// A predicate that specifies when to use the optional signing headers
	optionalHeadersPredicate := func (r *http.Request) bool {
		return r.Method == http.MethodPost
	}

	// And a provider of cryptographic keys
	provider := common.DefaultConfigProvider()

	// Build the signer
	signer := common.RequestSigner(provider, defaultGenericHeaders, optionalHeaders, optionalHeadersPredicate)

	// Sign the request
	signer.Sign(&request)

	// Execute the request
	c.Do(request)

For more information on the signing algorithm refer to: https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/signingrequests.htm

Polymorphic json requests and responses

Some operations accept or return polymorphic json objects. The SDK models such objects as interfaces. Further the SDK provides
structs that implement such interfaces. Thus, for all operations that expect interfaces as input, pass the struct in the SDK that satisfies
such interface. For example:

	c, err := identity.NewIdentityClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		panic(err)
	}

	// The CreateIdentityProviderRequest takes a CreateIdentityProviderDetails interface as input
	rCreate := identity.CreateIdentityProviderRequest{}

	// The CreateSaml2IdentityProviderDetails struct implements the CreateIdentityProviderDetails interface
	details := identity.CreateSaml2IdentityProviderDetails{}
	details.CompartmentId = common.String(getTenancyID())
	details.Name = common.String("someName")
	//... more setup if needed
	// Use the above struct
	rCreate.CreateIdentityProviderDetails = details

	// Make the call
	rspCreate, createErr := c.CreateIdentityProvider(context.Background(), rCreate)

In the case of a polymorphic response you can type assert the interface to the expected type. For example:

	rRead := identity.GetIdentityProviderRequest{}
	rRead.IdentityProviderId = common.String("aValidId")
	response, err := c.GetIdentityProvider(context.Background(), rRead)

	provider := response.IdentityProvider.(identity.Saml2IdentityProvider)

An example of polymorphic json request handling can be found here: https://github.com/oracle/oci-go-sdk/blob/master/example/example_core_test.go#L63


Pagination

When calling a list operation, the operation will retrieve a page of results. To retrieve more data, call the list operation again,
passing in the value of the most recent response's OpcNextPage as the value of Page in the next list operation call.
When there is no more data the OpcNextPage field will be nil. An example of pagination using this logic can be found here: https://github.com/oracle/oci-go-sdk/blob/master/example/example_core_test.go#L86

Logging and Debugging

The SDK has a built-in logging mechanism used internally. The internal logging logic is used to record the raw http
requests, responses and potential errors when (un)marshalling request and responses.

To expose debugging logs, set the environment variable "OCI_GO_SDK_DEBUG" to "1", or some other non empty string.


Forward Compatibility

Some response fields are enum-typed. In the future, individual services may return values not covered by existing enums
for that field. To address this possibility, every enum-type response field is a modeled as a type that supports any string.
Thus if a service returns a value that is not recognized by your version of the SDK, then the response field will be set to this value.

When individual services return a polymorphic json response not available as a concrete struct, the SDK will return an implementation that only satisfies
the interface modeling the polymorphic json response.


Contributions

Got a fix for a bug, or a new feature you'd like to contribute? The SDK is open source and accepting pull requests on GitHub
https://github.com/oracle/oci-go-sdk

License

Licensing information available at: https://github.com/oracle/oci-go-sdk/blob/master/LICENSE.txt

Notifications

To be notified when a new version of the Go SDK is released, subscribe to the following feed: https://github.com/oracle/oci-go-sdk/releases.atom

Questions or Feedback

Please refer to this link: https://github.com/oracle/oci-go-sdk#help




 */
package oci

//go:generate go run cmd/genver/main.go cmd/genver/version_template.go --output common/version.go

// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"io"
	"net/http"
)

// PutObjectRequest wrapper for the PutObject operation
type PutObjectRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// The name of the bucket.
	// Example: `my-new-bucket1`
	BucketName *string `mandatory:"true" contributesTo:"path" name:"bucketName"`

	// The name of the object.
	// Example: `test/object1.log`
	ObjectName *string `mandatory:"true" contributesTo:"path" name:"objectName"`

	// The content length of the body.
	ContentLength *int `mandatory:"true" contributesTo:"header" name:"Content-Length"`

	// The object to upload to the object store.
	PutObjectBody io.ReadCloser `mandatory:"true" contributesTo:"body" encoding:"binary"`

	// The entity tag to match. For creating and committing a multipart upload to an object, this is the entity tag of the target object.
	// For uploading a part, this is the entity tag of the target part.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"if-match"`

	// The entity tag to avoid matching. The only valid value is ‘*’, which indicates that the request should fail if the object already exists.
	// For creating and committing a multipart upload, this is the entity tag of the target object. For uploading a part, this is the entity tag
	// of the target part.
	IfNoneMatch *string `mandatory:"false" contributesTo:"header" name:"if-none-match"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`

	// 100-continue
	Expect *string `mandatory:"false" contributesTo:"header" name:"Expect"`

	// The base-64 encoded MD5 hash of the body.
	ContentMD5 *string `mandatory:"false" contributesTo:"header" name:"Content-MD5"`

	// The content type of the object.  Defaults to 'application/octet-stream' if not overridden during the PutObject call.
	ContentType *string `mandatory:"false" contributesTo:"header" name:"Content-Type"`

	// The content language of the object.
	ContentLanguage *string `mandatory:"false" contributesTo:"header" name:"Content-Language"`

	// The content encoding of the object.
	ContentEncoding *string `mandatory:"false" contributesTo:"header" name:"Content-Encoding"`

	// Optional user-defined metadata key and value.
	OpcMeta map[string]string `mandatory:"false" contributesTo:"header-collection" prefix:"opc-meta-"`
}

func (request PutObjectRequest) String() string {
	return common.PointerString(request)
}

// PutObjectResponse wrapper for the PutObject operation
type PutObjectResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The base-64 encoded MD5 hash of the request body as computed by the server.
	OpcContentMd5 *string `presentIn:"header" name:"opc-content-md5"`

	// The entity tag for the object.
	ETag *string `presentIn:"header" name:"etag"`

	// The time the object was modified, as described in RFC 2616 (https://tools.ietf.org/rfc/rfc2616), section 14.29.
	LastModified *common.SDKTime `presentIn:"header" name:"last-modified"`
}

func (response PutObjectResponse) String() string {
	return common.PointerString(response)
}

package request

import (
	"fmt"
	"net/http"
)

func (d *StorageServiceDriver) CreateContainer(containerName string) (resp *http.Response, err error) {

	verb := "PUT"

	uri := fmt.Sprintf("https://%s.blob.core.windows.net/%s?restype=container",  d.account, containerName)

	dateInRfc1123Format := currentTimeRfc1123Formatted()

	headers := map[string]string{
		"x-ms-version":  "2011-08-18",
		"x-ms-date":  dateInRfc1123Format,
	}

	canonicalizedHeaders := d.buildCanonicalizedHeader(headers)
	canonicalizedResource, err := d.buildCanonicalizedResource(uri)

	if err != nil {
		return nil, err
	}

	contentEncoding := ""
	contentLanguage := ""
	contentLength := "0"
	contentMD5 := ""
	contentType := ""
	date := ""
	ifModifiedSince := ""
	ifMatch := ""
	ifNoneMatch := ""
	ifUnmodifiedSince := ""
	Range := ""

	canonicalizedString := d.buildCanonicalizedString(verb, contentEncoding, contentLanguage, contentLength, contentMD5, contentType,
		date, ifModifiedSince, ifMatch, ifNoneMatch, ifUnmodifiedSince, Range, canonicalizedHeaders, canonicalizedResource)

	authHeader, err := d.createAuthorizationHeader(canonicalizedString)
	if err != nil {
		return nil, err
	}

	headers["Authorization"] = authHeader

	resp, err = d.Exec(verb, uri, headers, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

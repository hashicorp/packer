package request

import (
	"fmt"
	"net/http"
)

func (d *StorageServiceDriver) ListContainers() (resp *http.Response, err error) {

	verb := "GET"

	uri := fmt.Sprintf("https://%s.blob.core.windows.net/?comp=list&include=metadata",  d.account)

	dateInRfc1123Format := currentTimeRfc1123Formatted()

	headers := map[string]string{
		"x-ms-version":  "2009-09-19",
		"x-ms-date":  dateInRfc1123Format,
	}

	canonicalizedHeaders := d.buildCanonicalizedHeader(headers)
	canonicalizedResource, err := d.buildCanonicalizedResource(uri)

	if err != nil {
		return nil, err
	}

	contentEncoding := ""
	contentLanguage := ""
	contentLength := ""
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


//	fmt.Printf("--------------------ListContainers headers:\n %v\n", headers)


	resp, err = d.Exec(verb, uri, headers, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

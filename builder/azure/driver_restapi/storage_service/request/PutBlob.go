package request

import (
	"fmt"
	"net/http"
	"bytes"
	"os"
	"path/filepath"
	"log"
)

func (d *StorageServiceDriver) PutBlob(containerName, filePath string) (resp *http.Response, err error) {

	errMsg := "PutBlob error: %s"

	if _, err := os.Stat(filePath); err != nil {
		fmt.Errorf(errMsg, err.Error())
	}

	blobName := filepath.Base(filePath)

	fs, err := os.Open(filePath)
	defer fs.Close()

	var fileBuff bytes.Buffer
	_, err = fileBuff.ReadFrom(fs)
	blobContent := fileBuff.Bytes()

	blobSize := fmt.Sprintf("%d", len(blobContent))

	log.Printf("blobSize = '%s'", blobSize)

	verb := "PUT"

	uri := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", d.account, containerName, blobName)

	dateInRfc1123Format := currentTimeRfc1123Formatted()

	headers := map[string]string{
		"Content-Length":  	blobSize,
		"x-ms-version":  	"2011-08-18",
		"x-ms-date":  		dateInRfc1123Format,
		"x-ms-blob-type":  	"BlockBlob",
	}

	log.Printf("headers: %v\n", headers)

	canonicalizedHeaders := d.buildCanonicalizedHeader(headers)
	canonicalizedResource, err := d.buildCanonicalizedResource(uri)

	if err != nil {
		return nil, err
	}

	contentEncoding := ""
	contentLanguage := ""
	contentLength := blobSize
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

	resp, err = d.Exec(verb, uri, headers, &fileBuff)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

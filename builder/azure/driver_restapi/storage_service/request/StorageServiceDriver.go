package request

import (
	"net/http"
	"io"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"fmt"
	"encoding/xml"
	"io/ioutil"
	"regexp"
	"strings"
	"sort"
	"net/url"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/settings"
	"time"
)

type StorageServiceDriver struct {
	httpClient *http.Client
	account string
	secret string
}

func NewStorageServiceDriver(account, secret string) (*StorageServiceDriver) {
	client := &http.Client {}
	ssd := &StorageServiceDriver{
		httpClient: client,
		account: account,
		secret: secret,
	}
	return ssd
}

func (d *StorageServiceDriver) GetProps() (account, secret string ) {
	return d.account, d.secret	
}

func (d *StorageServiceDriver) Exec(verb, url string, headers map[string]string, body io.Reader) (resp *http.Response, err error) {
	var req *http.Request

	req, err = http.NewRequest(verb, url, body)

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err = d.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode

	if 	statusCode >= 400 && statusCode <= 505 {

		defer resp.Body.Close()
		errXml := new (ErrorXml)

		var respBody []byte
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if settings.LogRawResponseError {
			log.Printf("Raw Error:\n%v\n", string(respBody))
		}

		err = xml.Unmarshal(respBody, errXml)
		if err != nil {
			return nil, err
		}

		err = fmt.Errorf("%s %s", "Remote server returned error:", errXml.Message)

		return nil, err
	}

	if resp != nil {
		log.Printf("Exec resp: %v\n", resp)
	}

	return resp, err
}

func (d *StorageServiceDriver) computeHmac256(message string) (string, error) {
	errMsg := "computeHmac256 error: %s"
	key, err := base64.StdEncoding.DecodeString(d.secret)
	if err != nil {
		return "", fmt.Errorf(errMsg, err.Error())
	}
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func (d *StorageServiceDriver) createAuthorizationHeader (canonicalizedString string ) (string, error) {
	signature, err := d.computeHmac256(canonicalizedString)
	if err != nil {
		return "", err
	}
	authorizationHeader := fmt.Sprintf("%s %s:%s", "SharedKey", d.account, signature )

//	fmt.Printf("---------------------createAuthorizationHeader:\n%s\n", authorizationHeader)

	return authorizationHeader, nil
}

func (d *StorageServiceDriver) buildCanonicalizedHeader( headers map[string]string ) string {

	cm := make(map[string]string)

	for k,v := range(headers) {
		headerName := strings.TrimSpace(strings.ToLower(k))
		match, _ := regexp.MatchString("x-ms-", headerName)
		if match {
			cm[headerName] = v
		}
	}
	
	if len(cm) == 0 {
		return ""
	}

	keys := make([]string, 0, len(cm))
	for key,_ := range cm {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	ch := ""

	for i, key := range keys  {
		if i == len(keys) - 1 {
			ch += fmt.Sprintf("%s:%s", key, cm[key])
		} else {
			ch += fmt.Sprintf("%s:%s\n", key, cm[key])
		}
	}

//	fmt.Printf("-----------------------buildCanonicalizedHeader:\n%s\n", ch)

	return ch
}

func (d *StorageServiceDriver) buildCanonicalizedResource( uri string ) (string, error) {
	errMsg := "buildCanonicalizedResource error: %s"

	u, err := url.Parse(uri)

	if err != nil {
		return "", fmt.Errorf(errMsg, err.Error())
	}

	cr := "/"+d.account

	if len(u.Path) > 0 {
		cr += u.Path
	}

	params, err := url.ParseQuery(u.RawQuery)

	if err != nil {
		return "", fmt.Errorf(errMsg, err.Error())
	}

	if len(params) > 0 {
		cr += "\n"
		keys := make([]string, 0, len(params))
		for key,_ := range params {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for i, key := range(keys) {
			if len(params[key]) > 1 {
				sort.Strings(params[key])
			}

			if i == len(keys) - 1 {
				cr += fmt.Sprintf("%s:%s", key, strings.Join(params[key], ","))
			} else {
				cr += fmt.Sprintf("%s:%s\n", key, strings.Join(params[key], ","))
			}
		}
	}

//	fmt.Printf("--------------------buildCanonicalizedResource:\n%s\n", cr)

	return cr, nil
}

func (d *StorageServiceDriver) buildCanonicalizedString( verb, contentEncoding, contentLanguage, contentLength, contentMD5, contentType,
	date, ifModifiedSince, ifMatch, ifNoneMatch, ifUnmodifiedSince, Range, canonicalizedHeaders, canonicalizedResource string  ) string {

	canonicalizedString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
		verb,
		contentEncoding,
		contentLanguage,
		contentLength,
		contentMD5,
		contentType,
		date, ifModifiedSince, ifMatch, ifNoneMatch, ifUnmodifiedSince, Range,
		canonicalizedHeaders,
		canonicalizedResource)

	if settings.LogCanonicalizedString {
		log.Printf("--------------------buildCanonicalizedString:\n'%s'\n", canonicalizedString)
	}

	return canonicalizedString
}

type ErrorXml struct {
	Code string
	Message string
	AuthenticationErrorDetail string
}

func (d *StorageServiceDriver) buildShareAccessSignature(container, signedstart, signedexpiry string) (string, error) {

//	StringToSign = r + \n
//               2009-02-09 + \n
//               2009-02-10 + \n
//               /myaccount/pictures + \n
//               YWJjZGVmZw== + \n
//               2012-02-12



	stringToSign := fmt.Sprintf("%s\n%s\n%s\n/%s/%s\n%s", "r", signedstart, signedexpiry, d.account, container, "2012-02-12" )
	fmt.Println("stringToSign:\n " + stringToSign)

	urlDec, err := url.QueryUnescape(stringToSign)
	if err != nil {
		return "", err
	}

	fmt.Println("urlDec: " + urlDec)


	signature, err := d.computeHmac256(urlDec)
	if err != nil {
		return "", err
	}

	fmt.Println("signature: " + signature)

	return signature, nil
}

func (d *StorageServiceDriver) GetContainerSAS(container string) (string, error) {

	ts := time.Now().UTC()
	signedstart := ts.Format(time.RFC3339)
	te := ts.Add(time.Hour*24)
	signedexpiry := te.Format(time.RFC3339)

	sas, err := d.buildShareAccessSignature(container, signedstart, signedexpiry)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("?sv=2012-02-12&st=%s&se=%s&sr=c&sp=r&sig=%s", signedstart, signedexpiry, url.QueryEscape(sas)), nil
}

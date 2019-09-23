package vminstance

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

func SignZStack(accessKey, keySecret, method, date, path string) string {
	t := getSign(keySecret, method, date, path)
	return "ZStack " + accessKey + ":" + t
}

func getSign(keySecret, method, date, path string) string {
	signature := getStringToSign(method, date, path)
	hashed := hmac.New(sha1.New, []byte(keySecret))
	hashed.Write([]byte(signature))

	return base64.StdEncoding.EncodeToString(hashed.Sum(nil))
}

func getStringToSign(method, date, path string) string {
	return method + "\n" + date + "\n" + path
}

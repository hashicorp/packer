This is a fork of the from the original PKCS#12 parsing code
published in the Azure repository [go-pkcs12](https://github.com/Azure/go-pkcs12).
This fork adds serializing a x509 certificate and private key as PKCS#12 binary blob
(aka .PFX file).  Due to the specific nature of this code it was not accepted for
inclusion in the official Go crypto repository.

The methods used for decoding PKCS#12 have been moved to the test files to further
discourage the use of this library for decoding.  Please use the official
[pkcs12](https://godoc.org/golang.org/x/crypto/pkcs12) library for decode support.

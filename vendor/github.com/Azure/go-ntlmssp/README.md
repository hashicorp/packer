# go-ntlmssp
Golang package that provides NTLM/Negotiate authentication over HTTP

[![GoDoc](https://godoc.org/github.com/Azure/go-ntlmssp?status.svg)](https://godoc.org/github.com/Azure/go-ntlmssp) [![Build Status](https://travis-ci.org/Azure/go-ntlmssp.svg?branch=dev)](https://travis-ci.org/Azure/go-ntlmssp)

Protocol details from https://msdn.microsoft.com/en-us/library/cc236621.aspx
Implementation hints from http://davenport.sourceforge.net/ntlm.html

This package only implements authentication, no key exchange or encryption. It
only supports Unicode (UTF16LE) encoding of protocol strings, no OEM encoding.
This package implements NTLMv2.

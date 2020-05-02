[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://en.wikipedia.org/wiki/MIT_License)
[![Build Status](https://travis-ci.org/gford1000-go/chanmgr.svg?branch=master)](https://travis-ci.org/gford1000-go/chanmgr)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/gford1000-go/chanmgr)


ChanMgr | Geoff Ford
====================

The chanmgr package provides a simple way to introduce channel based processing within a go programme.  

An example of use is available in GoDocs.

To use, create a slice of `InOut{}` that specify the `Processor` to be called, and whether its results are wanted,
assign to a chanmgr via a call to `New()`, and then start sending data to be processed by the `Processor` using
`InOut.Send()`.  

`Send` returns a `Response` that can be checked for results availability, and then to retrieve
them when required using `Response.Get()`.  

`InOut.SendRecv()` combines the `Send` with the `Get` to block until a response is available from the `Processor`.

The chanmgr's request buffer size is set using `Config.RequestBuffer` and defaults to 1, which will therefore 
cause blocking on `Send` request.  

When chanmgr buffer size is larger than one, chanmgr ensures that responses from the `Processor` are associated
with the particular request.  `Send` allows a context value to be sent which is returned with the response, to allow
a different goroutine to handle the response from the submitter, and continue with the same context.


Installing and building the library
===================================

This project requires Go 1.14.2

To use this package in your own code, install it using `go get`:

    go get github.com/gford1000-go/chanmgr

Then, you can include it in your project:

	import "github.com/gford1000-go/chanmgr"

Alternatively, you can clone it yourself:

    git clone https://github.com/gford1000-go/chanmgr.git

Testing and benchmarking
========================

To run all tests, `cd` into the directory and use:

	go test -v


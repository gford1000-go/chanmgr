[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://en.wikipedia.org/wiki/MIT_License)
[![Build Status](https://travis-ci.org/gford1000-go/chanmgr.svg?branch=master)](https://travis-ci.org/gford1000-go/chanmgr)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/gford1000-go/chanmgr)


ChanMgr | Geoff Ford
==================================

The chanmgr package provides a simple way to introduce channel based processing within a go programme.  

An example of use is available in GoDocs.

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


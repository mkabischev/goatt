# goatt [![Build Status](https://travis-ci.org/mguzelevich/goatt.svg?branch=master)](https://travis-ci.org/mguzelevich/goatt)

API testing tools. Some tools for automation testing protocols (mq, http, etc.)

actual version installation:

```
$ go get github.com/mguzelevich/goatt/...
```

# NATS.io client

scanario example: `https://github.com/mguzelevich/goatt/blob/master/yaml/nats.example.yaml`

play scenario

```
$GOPATH/bin/goatt -yaml yaml/nats.example.yaml
```

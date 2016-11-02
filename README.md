gorunner
========

[![Build Status](https://secure.travis-ci.org/lirios/ci.png?branch=develop)](http://travis-ci.org/lirios/ci)

gorunner is an attempt to create a continuous integration web server written in Golang.

This project is a work-in-progress but development is not very active. I accept pull requests but also if you want to take it in a different direction let me know and we can collaborate.

Installation instructions
----

Assuming $GOPATH/bin is on your path:

	go get github.com/lirios/ci
	cd $GOPATH/src/github.com/lirios/ci
	gorunner

## Configuration

Before running you need to create a `config.ini` file in the same
directory where the executable will be.

Here's an example:

```
[Server]
URL=https://www.somewebsite.com
Port=localhost:8090
DbRootPath=data/
OutputPath=output/

[Slack]
WebHookURL=XXX
Channel=#events
```

With this configuration the CI server will bind to the `8090` port
on `localhost` and will save data files into the `data/` directory.

Artifacts and logs will be saved under the `output/` tree.

Slack notifications will go into the `#events` channel.

Technologies
----

* Go (golang)
* Javascript
  * Angularjs
  * Websockets

Why Go?
----

Go's ability to handle many connections would be beneficial for:

* running multiple build scripts and monitoring progress
* connecting to a cluster of gorunner servers
* live updates to builds in the UI via websockets, etc

![gorunner](https://raw.githubusercontent.com/lirios/ci/develop/promo.png "gorunner")

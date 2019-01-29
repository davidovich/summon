Summon
======

Summon is used to create a central location to gather scripts, or
go executable references.

It solves the maintenance problem of mutliple copies of same
code snippets distributed in many repos, leveraging go modules and version
management.

> NOTE: This is still a WIP and experimental. This readme is a design document and
not every feature is implemented yet.

It builds upon packr to pack arbitrary files in an executable
which you then bootstrap at destination using standard go get or [gobin](https://github.com/myitcv/gobin):

Configuration
-------------

Create a data repository with this structure (summon-cli will allow bootstrapping this structure in the future):

```
.
├── Makefile
├── assets
│   ├── bin
│   │   └── packr2.summon
│   ├── text.txt
│   └── config.yaml
├── boxer
│   └── box.go
├── go.mod
├── go.sum
└── summon
    └── summon.go
```

You just need to populate the `assets` directory with your own data.

The `boxer/` directory is used to provide Boxes to be found by [packr2](https://github.com/gobuffalo/packr/tree/master/v2).
The `summon/` direcotry is the entry point to the summon library, and creates the main executable.

There is an example setup at https://github.com/davidovich/summon-example-assets.

The `assets/config.yaml` contains a configuration file to customize summon. You can define:

    * aliases
    * default output-dir
    * gobin flags
    * go gettable executables

Build
-----

In an asset data repository:

0) (Comming soon) invoke summon-cli create
    This will create code template similar as above
1) Use the provided Makefile to invoke the packr2 process: make
2) Commit the resulting -packr files so clients can go get the data repo
3) Tag the repo with semantic version (with the `v`) prefix

Install
-------

Install using gobin the asset repo which will become the summon executable.
If the consuming repo needs to version the data alonside the consumer (each consumer repo could have a specific version of data),
you have two alternatives:

* use gobin to install summon in the consuming repo:

```
GO111MODULE=off go get -u github.com/myitcv/gobin
# install the data repository as summon executable
GOBIN=./ gobin [your-go-repo-import]/summon[@wanted-version-or-branch]
```

* declare a [tools.go](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module) file (will update when I get to testing this).

Use-cases
---------

### Makefile library

In makefiles it can be useful to centralize certain libraries, notice how
summon returns the path ot where the resource was instantiated:

```
include $(shell summon version.mk)
```

By default, summon will put summoned scripts at the `.summoned/` directory at root of the current directory.

### Running a go binary (soon)

`summon run [executable]` allows to run go gettable executables referenced in the `/bin` directory of the assets folder

### dumping the data at a location (soon)

```
summon --all --to .dir
```
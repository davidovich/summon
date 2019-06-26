[![CircleCI](https://circleci.com/gh/davidovich/summon.svg?style=svg)](https://circleci.com/gh/davidovich/summon)
![Custom badge](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Fdavidovich.github.io%2Fshields%2Fsummon%2Fsummon.json)
[![GoDoc](https://godoc.org/github.com/davidovich/summon?status.svg)](https://godoc.org/github.com/davidovich/summon)

Summon
======

Summon is used to manage a central location of data or
executable references, allowing distribution to any go-enabled environment. You can use it in your team to share common snippets of code or domain knowlege.

It solves the maintenance problem of multiple copies of same
code snippets distributed in many repos (like general makefile recipies), leveraging go modules and version
management. It also allows configuring a standard set of tools that a dev team can readily
invoke by name.

You can make an analogy with a data singleton which always has the desired
state (packed scripts or pinned versions of binaries).

> NOTE: This project is still a WIP and experimental.

To install, you first need to create something to install by populating a [data repository](#Configuration). Then, this data repo is installed by using [gobin](https://github.com/myitcv/gobin) (or go install):

```
gobin [your-summon-data-repo]/summon
```

Assuming there is a my-team-utility.sh script hosted in the data repo, (see how to [configure](#Configuration) below) you can do things like:

```
bash $(summon my-team-utility.sh)
```

How it Works
------------

Summon is a library which is consumed by an asset repository (which, by default, has also the `summon` name). This asset repository, managed by your team, provides the summon executable command (it's main() function is in summon/summon.go).
The library provides the command-line interface, so no coding is necessary in the assert repo.

Summon also provides a boostrapping feature in the scaffold command.

Summon builds upon [packr2](https://github.com/gobuffalo/packr/tree/master/v2) to convert an arbitrary tree of files in go compilable source
which you then bootstrap at destination using standard go get or [gobin](https://github.com/myitcv/gobin). The invoked files are then placed locally and the summoned file path is returned so it can be consumed by other shell operations.

Configuration
-------------

### Data repository

Use summon's scaffold feature to create a data repository, which will become your singleton data executable.

> Scaffolding is new in v0.1.0

```
gobin -run github.com/davidovich/summon/scaffold init [module name]
```

> Be sure to change the [module name] part (usually you will use a module path compatible with where you will host the data repo on a code hosting site).

You will then have something resembling this structure:

```
.
├── Makefile
├── README.md
├── assets
│   └── summon.config.yaml
├── go.mod
└── summon
    └── summon.go
```

There is an example setup at https://github.com/davidovich/summon-example-assets.

You just need to populate the `assets` directory with your own data.

The `summon/` directory is the entry point to the summon library, and creates the main command executable. This directory will also host
[packr2](https://github.com/gobuffalo/packr/tree/master/v2) generated data files which encode asset data into go files.

The `assets/summon.config.yaml` is an (optional) configuration file to customize summon. You can define:

    * aliases
    * default output-dir
    * executables


```yaml
version: 1
outputdir: .summoned
aliases:
    simple-handle: a/file/in/asset-dir
# exec section declares invokables with their handle
# a same handle name cannot be in two invokers at the same time
exec:
    bash -c:
    # ^ invoker
        # (notice script can be inlined because of invoker type)
        hello: echo hello
        # ^ handle to script (must be unique)

    gobin -run: # go gettable executables
        gobin: github.com/myitcv/gobin@v0.0.8
        gohack: github.com/rogppepe/gohack

    python -c:
        hello-python: print("hello from python!")
```

You can invoke executables like so:

```
summon run gohack ...
```

This will install gohack using `gobin -run` and forward the arguments that you provide.

Build
-----

In an empty asset data repository directory:

0) Invoke `go run github.com/davidovich/summon/scaffold init [repo host (module name)]`
    This will create code template similar as above
1) Add assets that need to be shared amongst consumers
2) Use the provided Makefile to invoke the packr2 process: `make`
3) Commit the resulting -packr files so clients can go get the data repo
4) Tag the repo with semantic version (with the `v`) prefix.
5) Push to remote.
6) Install.


Install
-------

Install (using gobin) the asset repo which will become the summon executable.
If the consumer site needs to version the data alonside the consumer (each site could have a specific version of data),
you have two alternatives:

* use gobin to install summon in the consuming site:

```
GO111MODULE=off go get -u github.com/myitcv/gobin
# install the data repository as summon executable at the site
GOBIN=./ gobin [your-go-repo-import]/summon[@wanted-version-or-branch]
```

* declare a [tools.go](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module) file (will update when I get to testing this).

Use-cases
---------

### Makefile Library

In makefiles it can be useful to centralize certain libraries, notice how
summon returns the path ot where the resource was instantiated:

```
include $(shell summon version.mk)
```

By default, summon will put summoned scripts at the `.summoned/` directory at root of the current directory.

### Templating

Files in the asset directory can contain go templates. This allows applying customization using json data.

> New in v0.3.0, summon now uses the [Sprig templating library](https://github.com/Masterminds/sprig), which provides many useful templating functions.

For example, consider this file in a summon asset provider:

```
/assets
   template.file
```
With this content:

```
Hello {{ .Name }}!
```

`summon template.file --json '{ "Name": "David" }'`

You will get a summoned `template.file` file with this result:

```
Hello David!
```

> New in v0.2.0, you can summon a whole asset hierarchy by using a directory reference when summoning.


Templates can also be used in the filenames given in the data hierarchy. This can be useful to scaffold simple projects.

```
/assets
   /template
      {{.FileName}}
```

Then you can summon this hierarchy by introducing a `FileName` in the json parameter.

`summon template/ --json '{ "FileName": "myRenderedFileName" }' -o dest-dir`

will yield:

```
./dest-dir
   myRenderedFileName
```

### Running A Binary

`summon run [executable]` allows to run executables declared in the config file

### Dumping the Data at a Location

```
summon --all --out .dir
```

### Output a File to stdout

```
summon my-file -o-
```

### Output a Template File Without Rendering

```
summon my-template -o- --raw
```

### List Summon Contents

```
summon ls

summon ls --tree # pretty print hierarchy
```

### View Data Version Information

```
summon -v
```

### Configure Bash Completion

> New in v0.8.0

```
source <(summon completion)
```

Alternatives
------------

Why not use git directly?

While you could use git directly to bring an asset directory with a simple git clone, the result does not have executable properties.
In summon you leverage go execution to bootstrap in one phase. So your data can do:

```
go run github.com/davidovich/summon-example-assets/summon --help
# or list the data deliverables
go run github.com/davidovich/summon-example-assets/summon ls
# or
# let summon configure the path so it can invoke a go executable
# (here go-gettable-executable is a reference to a go gettable repo), and will
# result in an executable tailored for your destination os and architecture (because built on the fly).
go run github.com/davidovich/summon-example-assets/summon run go-gettable-executable
```

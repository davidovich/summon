Summon
======

Summon is used to create a central location to gather scripts, or
go executable references.

It solves the maintenance problem of mutliple copies of same
code snippets distributed in many repos.

It builds upon gobin and packr to pack arbitrary files in an executable
which you then bootstrap at destination using [gobin](https://github.com/myitcv/gobin):

```
GO111MODULE=off go get -u github.com/myitcv/gobin
gobin github.com/davidovich/summon
```

Usages
------

In makefiles it can be useful to centralize certain portions that never change:

```
include $(shell summon version.mk)
```

By default, summon will put scripts at the `.summon/` directory at root of the current directory.
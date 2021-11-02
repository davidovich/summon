[![CircleCI](https://circleci.com/gh/davidovich/summon.svg?style=svg)](https://circleci.com/gh/davidovich/summon)
![Custom badge](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Fdavidovich.github.io%2Fshields%2Fsummon%2Fsummon.json)
[![GoDoc](https://godoc.org/github.com/davidovich/summon?status.svg)](https://godoc.org/github.com/davidovich/summon)

- [Summon](#summon)
  - [How it Works](#how-it-works)
  - [Configuration](#configuration)
    - [Data repository](#data-repository)
    - [Summon config File](#summon-config-file)
  - [Build](#build)
  - [Install](#install)
  - [Use-cases](#use-cases)
    - [Makefile Library](#makefile-library)
    - [Templating](#templating)
    - [Running A Binary](#running-a-binary)
      - [Templated Invokables](#templated-invokables)
      - [Keeping DRY](#keeping-dry)
      - [Template Functions Available in Summon](#template-functions-available-in-summon)
        - [Summon Function](#summon-function)
        - [Arg and Args Function](#arg-and-args-function)
        - [Run Function](#run-function)
      - [Removing the run subcommand](#removing-the-run-subcommand)
    - [Dump the Data at a Location](#dump-the-data-at-a-location)
    - [Output a File to stdout](#output-a-file-to-stdout)
    - [Output a Template File Without Rendering](#output-a-template-file-without-rendering)
    - [List Summon Contents](#list-summon-contents)
    - [Evaluate what will be run (--dry-run)](#evaluate-what-will-be-run---dry-run)
    - [View Data Version Information](#view-data-version-information)
    - [Configure Bash Completion](#configure-bash-completion)
  - [TODO](#todo)
  - [FAQ](#faq)

# Summon

Summon is used to manage a central location of data or
executable references, allowing distribution to any go-enabled environment. You can use it in your team to share common snippets of code or domain knowlege.

It solves the maintenance problem of multiple copies of same
code snippets distributed in many repos (like general makefile recipies), leveraging go modules and version
management. It also allows configuring a standard set of tools that a dev team can readily
invoke by name.

You can make an analogy with a data singleton which always has the desired
state (packed scripts or pinned versions of binaries).

> NOTE: This project is still a WIP and experimental.

To install, you first need to create something to install by populating a
[data repository](#Configuration). Then, this data repo is installed by using
the `go install` command:

```bash
go install [your-summon-data-repo]/summon@latest
```

Assuming there is a my-team-utility.sh script hosted in the data repo, (see how
to [configure](#Configuration) below) you can do things like:

```bash
bash $(summon my-team-utility.sh)
```

## How it Works

Summon is a library which is consumed by an asset repository (which, by default,
has also the `summon` name). This asset repository, managed by your team,
provides the summon executable command (it's main() function is in
summon/summon.go). The library provides the command-line interface, so no
coding is necessary in the assert repo.

Summon also provides a boostrapping feature in the scaffold command.

Summon builds upon the new go 1.16 [embed.FS](https://pkg.go.dev/embed) feature used to pack assets in a
go binary. You then install this at destination using standard `go install`.

When you invoke this binary with a contained asset path, the invoked files are
placed locally and the summoned file path is returned so it can be consumed by
other shell operations.

## Configuration

### Data repository

Use summon's scaffold feature to create a data repository, which will become your singleton data executable.

> Scaffolding is new in v0.1.0

```bash
# go run package at a version requires go 1.17 and up
go run github.com/davidovich/summon/scaffold@latest init [module name]
```

> Be sure to change the [module name] part (usually you will use a module path
> compatible with where you will host the data repo on a code hosting site).

You will then have something resembling this structure:

```bash
.
├── Makefile
├── README.md
├── assets
│   └── summon.config.yaml
├── go.mod
└── summon.go
```

There is an example setup at https://github.com/davidovich/summon-example-assets.

You just need to populate the `assets` directory with your own data.

The `summon.go` file of the main module is the entry point to the summon
library, and creates the main command executable.

### Summon config File

The `assets/summon.config.yaml` is an (optional) configuration file to
customize summon. You can define:

- aliases
- default output-dir
- executables

> Breaking in v0.11.0: Handles now take an array of params

```yaml
version: 1
outputdir: .summoned
aliases:
    simple-handle: a/file/in/asset-dir

templates: |
  {{/* new starting at v0.12.0, global templates available to command params */}}
  {{- define "maybeChangeUser" -}}
    {{- if (env "SUMMON_CHANGE_USER") -}}
      -u {{ env "SUMMON_CHANGE_USER" }}
    {{- end -}}
  {{- end -}}

# exec section declares invokables with their handle
# a same handle name cannot be in two invokers at the same time
# they are grouped by invoker.
exec:
    bash -c:
    # ^ invoker
        # (script can be inlined with | yaml operator)
        hello: [echo, hello]
                 # ^ optional params that will be passed to invoker
                 # these can contain templates (starting at v0.10.0)
        # ^ handle to script (must be unique). This is what you use
        # to invoke the script: `summon run hello`.

    go run: # go gettable executables
        gohack: [github.com/rogppepe/gohack@latest]

    python -c:
        hello-python: [print("hello from python!")]

    # Expose docker containers commands without having
    # to remember mounting volumes, etc.
    # `?` YAML construct allows multi-line keys for easier reading
    ? docker run -ti --rm -w /{{ env "PWD" | base }}
      -v {{ env "PWD" }}:/{{ env "PWD" | base }}
      {{ template "maybeChangeUser" }}
      alpine
    : ls: [ls]
```

You can invoke executables like so:

```bash
summon run gohack ...
```

This will call and run gohack using `go run` and forward the arguments that you
provide.

> New in v0.10.0, summon now allows templating in the invocable section. See
> [Templating](#/templating).

## Build

In an empty asset data repository directory:

- First (and once) invoke `go run github.com/davidovich/summon/scaffold@latest init [repo host (module name)]`
    This will create code template similar as above

1. Add assets that need to be shared amongst consumers
2. Use the provided Makefile to create the asset executable: `make`
3. Commit the all the files so clients can go get the data repo
4. Tag the repo with semantic version (with the `v`) prefix.
5. Push to remote.
6. On a consumer machine, install.

## Install

Install (using `go install`) the asset repo which will become the summon executable.

```bash
go install [your-go-repo-module]/summon@latest
```

## Use-cases

### Makefile Library

In makefiles it can be useful to centralize certain libraries, notice how
summon returns the path ot where the resource was instantiated:

```bash
include $(shell summon version.mk)
```

By default, summon will put summoned scripts at the `.summoned/` directory at
the root of the current directory. This can be changed with the `-o` flag.

### Templating

Files in the asset directory can contain go templates. This allows applying
customization using json data, just before rendering the file (and its contents).

> New in v0.3.0, summon now uses the [Sprig templating library](https://github.com/Masterminds/sprig), which provides many useful templating functions.

For example, consider this file in a summon asset provider:

```bash
/assets
   template.file
```

With this content:

```txt
Hello {{ .Name }}!
```

`summon template.file --json '{ "Name": "David" }'`

You will get a summoned `template.file` file with this result:

```shell
Hello David!
```

> New in v0.2.0, you can summon a whole asset hierarchy by using a directory
> reference when summoning.

Templates can also be used in the filenames given in the data hierarchy. This
can be useful to scaffold simple projects.

```bash
/assets
   /template
      {{.FileName}}
```

Then you can summon this hierarchy by introducing a `FileName` in the json parameter.

`summon template/ --json '{ "FileName": "myRenderedFileName" }' -o dest-dir`

will yield:

```bash
./dest-dir
   myRenderedFileName
```

### Running A Binary

`summon run [executable]` allows to run executables declared in the
[config file](#/summon-config-file).

> New in v0.10.0:
> * you can use go templates in the `exec:` section.
> * you can summon embedded data in the `exec:` section.

#### Templated Invokables

Suppose you want to make a wrapper around a docker utility. The specific
docker invocation can be quite cryptic. Help your team by adding an invocable
in the config file:

```yaml
...
exec:
  docker run -v {{ env "PWD" }}:/mounted-app alpine:
      ls: [ls, /mounted-app]
```

Calling `summon run ls` would render the
[{{ env "PWD" }}](https://masterminds.github.io/sprig/os.html) part to the
current directory, resulting in this call:

`docker run -v [working-dir]:/mounted-app alpine ls /mounted-app`

#### Keeping DRY

> New in v0.12.0

Sometimes you will use Summon as a proxy on a docker container. Some
parameters will always need to be passed (volume mounts for example). You
can use YAML anchors to define the static (but required) params in
`summon.config.yaml`:

```yaml
.base: &baseargs
    - echo
    - b
    - c

exec:
    bash -c:
      echo: [*baseargs, d]
```

Here, when you run with the `echo` handle, the arrays will be flattened to produce
`[echo, b, c, d]` for the construction of the command.

#### Template Functions Available in Summon

Summon comes with template functions that can be used in the config file or
contained assets.

##### Summon Function

Say you would like to bundle a script in the data repo and also use it as an
invocable (new in v.0.10.0). You would use the `summon` template function bundled in summon:

```yaml
exec:
  bash -c:
    hello: ['{{ summon "hello.sh" }}']
```

Assuming you have a `hello.sh` file in the assets repo, this would result in
sommoning the file in a temp dir and calling the invoker:

```shell
bash -c /tmp/hello.sh
```

> Note that `hello.sh` could also contain templates that will be
rendered at instanciation time.

##### Arg and Args Function

> New in v0.12.0

Summon provisions the `args`, `arg` functions and `.osArgs` slice of arguments. You can use
these in a template of the params array.

* `args` will contain unknown args passed from the command-line (see `ls` handle
  defined in the [config section](#/summon-config-file))

    ```bash
    summon run ls -al
                 [ ^ args array starts here ]
    ```

    Here, `{{ args }}` would return `[-al]`.

* `arg` allows accessing one arg, with an error message if arg is not found

    ```yaml
    ...
    exec:
       bash -c:
          ls: [ls, '{{ arg 0 "error msg" }}']
    ```

When used, summon will remove the consumed args, as this would
surprisingly double the args. In other words, when accessing `{{ args }}`,
summon will not append the resulting args, and using `{{ arg 0 "error" }}`,
summon would only append the unconsumed args (after index 0).

* `.osArgs` contains the whole command-line slice

If the result of using args is a string representation of an array, like
`[a b c d]` this array will be flattened to the final args array.

##### Run Function

> New in v0.12.0

The `run` function allows running a configured handle of the config file, right
inside the config file. This effectively opens many use cases of executing
code to control arguments. Called subprocesses can have side effects and can
be used to execute pre conditions.

Consider:

We want to mount volumes of a docker container, conditionally.

```yaml
exec:
  docker run -it --rm {{ run "eval-mounts" }} alpine:
    ls: ['ls']
  bash -c:
    eval-mounts: ["echo -v {{ env PWD }}:/app"]
```

When inovking `summon run ls`, summon will first invoke:

`bash -c 'echo -v current_dir:/app'` which yields `-v current_dir:/app` and
then call the `ls` handle to produce:

`docker run -it --rm -v current_dir:/app alpine`

> Warning: `run` must not start a recursive process as `summon` does not currently
> protect from this type of call. The consequence of doing this will probably
> result in a fork bomb.

#### Removing the run subcommand

> New in v0.12.0

In some situations, the data provider is more a proxy to other commands, so
it can make sense to optimize this use-case and remove the `run` subcommand.

This is done by passing `summon.WithoutRunCmd()` option to the `summon.Main()`
entry-point function. In that mode, all invocable handles become part of the
main command.

In this mode, the `ls` subcommand to list embedded assets becomes a `--ls` flag.

### Dump the Data at a Location

```bash
summon --all --out .dir
```

### Output a File to stdout

```bash
summon my-file -o-
```

### Output a Template File Without Rendering

```bash
summon my-template -o- --raw
```

### List Summon Contents

```bash
summon ls

summon ls --tree # pretty print hierarchy
```

### Evaluate what will be run (--dry-run)

> New in v0.10.0

```bash
summon run -n ls -al
Would execute `/usr/local/bin/docker run -ti --rm -w /application -v [current-dir]:/application alpine ls -al`...
```

### View Data Version Information

```bash
summon -v
```

### Configure Bash Completion

> New in v0.8.0

```bash
source <(summon completion)
```

## TODO

- [ ] Add a `required` template function to enforce `.args` presence, and error
  with a message.
- [ ] Add help documentation for proxied commands
- [ ] Explore ways to hook completions from proxied commands.

## FAQ

- Why is the `exec:` config file ordered by "invoker" ?

  Summon is oriented at providing an easy CLI interface to complex sub programs.
  In this regard, it tends to group invocations in the same execution "environment".

  This helps in scenarios of supplying a dev container from which are surfaced
  tools for your team.

- Why not use git directly ?

  While you could use git directly to bring an asset directory with a simple git clone, the result does not have executable properties.

  In summon you leverage go execution to bootstrap in one phase. So your data can do:

  ```bash
  go run github.com/davidovich/summon-example-assets/summon --help
  # or list the data deliverables
  go run github.com/davidovich/summon-example-assets/summon ls
  ```

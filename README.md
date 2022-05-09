[![CircleCI](https://circleci.com/gh/davidovich/summon.svg?style=svg)](https://circleci.com/gh/davidovich/summon)
![Custom badge](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Fdavidovich.github.io%2Fshields%2Fsummon%2Fsummon.json)
[![GoDoc](https://godoc.org/github.com/davidovich/summon?status.svg)](https://godoc.org/github.com/davidovich/summon)

- [Summon](#summon)
  - [How it Works](#how-it-works)
  - [Configuration](#configuration)
    - [Data repository](#data-repository)
      - [Migration from versions prior to v0.13.0](#migration-from-versions-prior-to-v0130)
    - [Summon config File](#summon-config-file)
  - [Build](#build)
  - [Install](#install)
  - [Use-cases](#use-cases)
    - [Makefile Library](#makefile-library)
    - [Templating](#templating)
    - [Running a Binary](#running-a-binary)
      - [Templated Execution handles](#templated-execution-handles)
      - [Enhanced command description](#enhanced-command-description)
      - [Keeping DRY](#keeping-dry)
      - [Template Functions Available in Summon](#template-functions-available-in-summon)
        - [Summon Function](#summon-function)
        - [Arg and Args Function](#arg-and-args-function)
        - [Run Function](#run-function)
        - [flagValue Function](#flagvalue-function)
        - [.flag field](#flag-field)
      - [A Note on Completions](#a-note-on-completions)
      - [Removing the run subcommand](#removing-the-run-subcommand)
    - [Dump the Data at a Location](#dump-the-data-at-a-location)
    - [Output a File to stdout](#output-a-file-to-stdout)
    - [Output a Template File Without Rendering](#output-a-template-file-without-rendering)
    - [List Summon Contents](#list-summon-contents)
    - [Evaluate what will be run (--dry-run)](#evaluate-what-will-be-run---dry-run)
    - [View Data Version Information](#view-data-version-information)
    - [Configure Bash Completion](#configure-bash-completion)
  - [TODO](#todo)
  - [Acknowledgments](#acknowledgments)
  - [FAQ](#faq)

# Summon

Summon is used to manage a central location of data or
executable references, allowing distribution to any go-enabled environment.
You can use it in your team to share common snippets of code or domain knowledge.

It solves the maintenance problem of multiple copies of same
code snippets distributed in many repos (like general makefile recipes),
leveraging go modules and version management. It also allows configuring a
standard set of tools that a dev team can readily invoke by name.

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
coding is necessary in the asset repo.

Summon also provides a boostrapping feature in the scaffold command.

> New in v0.13.0

Summon builds upon the new go 1.16 [embed.FS](https://pkg.go.dev/embed) feature
used to pack assets in a go binary. You then install this at destination using
standard `go install`.

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
├── go.mod
├── go.sum
└── summon
    ├── assets
    │   └── summon.config.yaml
    └── summon.go
```

There is an example setup at https://github.com/davidovich/summon-example-assets.
Also, a simple fake utility is also hosted in the `examples/` directory.

You just need to populate the `summon/assets` directory with your own data.

The `summon/summon.go` file of the main module is the entry point to the summon
library, and creates the main command executable.

#### Migration from versions prior to v0.13.0

The v0.13.0 version uses the embed.FS and the `//go:embed assets/*` directive.
Prior versions used to reference the `assets/` dir at the root of the repo.
This means that the data being embedded must now be a sibling of the source
file containing `package main`.

### Summon config File

The `summon/assets/summon.config.yaml` is an (optional) configuration file to
customize summon. You can define:

- aliases
- default output-dir
- handles to configured executables

> Breaking in v0.11.0: Handles now take an array of params

```yaml
version: 1 # although at version 1, this config is not quite stable yet, but it
           # is getting closer.

outputdir: ".summoned" # where summoned files are placed
hideAssetsInHelp: true # should the assets be shown in the help ?

aliases:
  simple-handle: a/file/in/asset-dir

templates: |
  {{/* new starting at v0.12.0, global templates available to command params */}}
  {{- define "maybeChangeUser" -}}
    {{- if (env "SUMMON_CHANGE_USER") -}}
      -u {{ env "SUMMON_CHANGE_USER" }}
    {{- end -}}
  {{- end -}}

# exec section declares flags and execution handles and their handle.
# A same handle name cannot be in two exec:handles at the same time.
exec:
  flags: # global flags that can be used in any `args:` section
    hello:
      effect: '{{.flag}}' # when the user uses the flag, it's value will be in
                          # the .flag variable
      default: 'world' # value to use if the flag is used alone (without value, `--hello`).
      shorthand: 'o' # -o can be used instead

  handles: # new and breaking in v0.15.0, this corresponds to the original
           # exec: section.
    hello: [bash, -c, 'echo hello'] # simple form of command configuration
  #     ^          ^ args
  #     |            these can contain templates (starting at v0.10.0)
  #     |- handle to script (must be unique). This is what you use
  #        to invoke the script: `summon run hello`.

  # New in v0.14.0: complex command and sub-command specification
  # Here, we define a proxy to gohack, with completion
  # See a complete definition in the examples/cmd-proxy/assets dir

    gohack [command]: #
      cmd: [go, run]
      args: &rog [github.com/rogpeppe/gohack@latest]
      # completion must produce a `\n` separated string that is used as candidates
      completion: '{{ printf "get\nundo\nstatus\nhelp" }}'
      subCmd: # subCmd is new in v0.14.0
        get:
          args: [*rog, get,'{{ flagValue "vcs" }}'] # note that this command
                                                    # definition is separate
                                                    # from the top level gohack
                                                    # so we use an anchor to keep
                                                    # DRY.
          flags:
            vcs:
              effect: '{{.flag}}'
              default: '-vcs'
        undo: [*rog, undo]

    hello-python: [python, -c, print("hello from python!")]

    # Expose docker containers commands without having
    # to remember mounting volumes, etc.
    ls: [docker, run, -ti, --rm, -w, '{{ env "PWD" | base }}',
      -v, '{{ env "PWD" }}:/{{ env "PWD" | base }}',
      '{{ template "maybeChangeUser" }}',
      alpine, ls]
```

You can invoke executables like so:

```bash
summon run gohack ...
```

This will call and run gohack using `go run` and forward the arguments that you
provide.

> New in v0.10.0, summon now allows templating in the invocable section. See
> [Templating](#templating).

## Build

In an empty asset data repository directory:

- First (and once) invoke `go run github.com/davidovich/summon/scaffold@latest init [repo host (module name)]`
    This will create code template similar as above in the current directory
    (this can be modified by using the `-o` [output dir] flag).

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

### Running a Binary

`summon run [handle]` allows to run executables declared in the
[config file](#summon-config-file).

> New in v0.10.0:
> - you can use go templates in the `exec:` section.
> - you can summon embedded data in the `exec:` section.

#### Templated Execution handles

Suppose you want to make a wrapper around a docker utility. The specific
docker invocation can be quite cryptic and long. Help your team by adding an
exec environment in the config file:

```yaml
...
exec:
  handles:
    ls: [docker, run, -v, '{{ env "PWD" }}:/mounted-app', alpine, ls, /mounted-app]
```

Calling `summon run ls` would render the
[{{ env "PWD" }}](https://masterminds.github.io/sprig/os.html) part to the
current directory, resulting in this call:

`docker run -v [working-dir]:/mounted-app alpine ls /mounted-app`

In effect, this feature allows creating your own cli that can wrap complex
containers. The cli is a kind of trampoline to the container.

Note that the whole environment line can be templated.

#### Enhanced command description

> New in v0.14.0

You can now build complex command line interfaces in a declarative way using a
command spec (see `pkg/config/config.go` for the struct definitions).

Typically, this will be used to simplify complex tools, or give a simple
interface to a complex docker invocation.

Below is a synthetic example that uses every available command and flag config.

```yaml
exec:
  handles:
    handle [possible param hint to user]: # 'handle' is used to invoke this docker container
      cmd: [docker] # this can be a complex docker invocation like mounting volumes (-v),
                    # container removal arg (--rm), passed environment (-e), interactive
                    # terminal (-ti), etc.]
      args: ['hardcoded-arg-1', '{{ arg 0 }}', '{{ flagValue "my-flag" }}']
      join: false # should the args array be joined by a space? Useful for
                  # `bash -c` type commands
      help: help that will be printed when user invokes `--help`
      hidden: false # should this command appear in the help or completion ?
      completion: '{{ }}' # dynamic completion candidates separated by `\n`.
                          # declared commands handles need not be listed here. But
                          # calling a surrogate process might be handy to
                          # complete a proxied command (especially if the
                          # command lives in a container!).
      subCmd:
        first: # sub-command name, as invoked on the command line.
          args: ['this is a complete new command description']
          # ...
          subCmd:
            second-sub-cmd: [] # same recursive structure
      flags:
        my-flag:
          effect: '{{.flag}}' # template to construct the value. The user
                              # provided value is put in the .flag variable.
                              # You can place the flag explicitly with the
                              # flagValue template function.
          shorthand: 'one letter shorthand (invoked with a single dash: i.e -i)'
          default: if the user provides no value, use this
          help: Help for the flag
          explicit: true # Use this flag to disable automatic appending of
                          # the flag effect to the args. If used in an argument
                          # args array, explicit is true.
```

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
  handles:
    echo: [docker, run, -ti, -v, '{{ env "PWD"}}:/workdir', -w, /workdir, alpine, *baseargs, d]
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
  handles:
    hello: [bash, '{{ summon "hello.sh" }}']
```

Assuming you have a `hello.sh` file in the assets repo, this would result in
summoning the file in a temp dir and calling the invoker:

```shell
bash /tmp/hello.sh
```

> Note that `hello.sh` could also contain templates that will be
rendered at instantiation time.

##### Arg and Args Function

> New in v0.12.0

Summon provisions the `args`, `arg` functions and `.osArgs` slice of arguments. You can use
these in a template of the params array.

- `args` will contain unknown args passed from the command-line (see `ls` handle
  defined in the [config section](#summon-config-file))

    ```bash
    summon run ls -al
                 [ ^ args array starts here ]
    ```

    Here, `{{ args }}` would return `[-al]`.

- `arg` allows accessing one arg, with an error message if arg is not found

    ```yaml
    ...
    exec:
      handles:
        ls: [bash, ls, '{{ arg 0 "error msg" }}']
    ```

When used, summon will remove the consumed args, as this would
surprisingly double the args in the resulting invocation. In other words, when
accessing `{{ args }}`, summon will not append the resulting args, and using
`{{ arg 0 "error" }}`, summon would only append the unconsumed args (after index 0).

- `.osArgs` contains the whole command-line slice

If the result of using args is a string representation of an array, like
`[a b c d]` this array will be flattened to the final args array.

##### Run Function

> New in v0.12.0

The `run` function allows running a configured handle of the config file, right
inside the config file. This effectively opens many use cases of executing
code to control arguments. Called sub-processes can have side effects and can
be used to execute pre conditions.

Consider:

We want to mount volumes of a docker container, conditionally.

```yaml
exec:
  handles:
    # here, "eval-mounts" is a reference to the corresponding handle
    ls:
      cmd: [docker, run, -it, --rm, '{{ run "eval-mounts" }}', alpine]
      args: ['ls']
    eval-mounts: [bash, -c, "echo -v {{ env PWD }}:/app"]
    #    ^ used to compute the volumes.
```

When inovking `summon run ls`, summon will first invoke:

`bash -c 'echo -v current_dir:/app'` which yields `-v current_dir:/app` and
then call the `ls` handle to produce:

`docker run -it --rm -v current_dir:/app alpine`

> WARNING: `run` must not start a recursive process as `summon` does not currently
> protect from this type of call. The consequence of doing this will probably
> result in a fork bomb.

##### flagValue Function

> New in v0.14.0

The `{{ flagValue "my-flag" }}` function is used in the `args:` section. To
render this function, summon looks at the `flags:` section to find the
corresponding flag and inserts it's rendered `effect:`. If the render produces
no value, the block is a no-op and has no effect in the passed arguments.

If the flag is not found in the top level handle command, or subCmd, summon
will render an empty value.

Use the `--dry-run` or `-n` to debug what the invocation would look like.

##### .flag field

> New in v0.14.0

The `{{ .flag }}` template field is only used and valid in the `effect:` flag
configuration field. It takes the value provided by the user. For example,
if the user provides `--my-flag=my-value` flag, the `.flag` template field
will hold the `my-value` value.

#### A Note on Completions

Surfacing a completion from a docker container hosted command can be a challenge.
While developing this feature, experimentation was done to trigger the completion
mechanisms of the target program. This is used to populate the completion from
the host machine by using the completion result in the container.

For example, we present below the completion command for a [`posener/complete`](https://github.com/posener/complete) based implementation.

Also, a [cobra](https://github.com/spf13/cobra) based program (kubectl).

```yaml
exec:
  handles:
    # tanka delegates it's completion to posener/complete. Fake a completion
    # call by setting the COMP_LINE enviroment var.
    # delegate this to a simple bash-c handle
    tk:
      cmd: &bash-c [bash, -c] # here bash -c is used to test, but normally this is a complex
                              # docker container invocation.
      completion: '{{ run "bash-c" (printf "COMP_LINE=''%s'' tk" (join " " args)) }}'

    kubectl: # cobra based commands are a bit more involved as we need to
             # filter the control characters (:0) that it outputs.
      cmd: *bash-c
      completion: '{{ (split ":" (run "bash-c" (printf "kubectl __complete %s ''%s''" (join " " (rest (initial args))) (last args))))._0 }}'

    bash --norc --noprofile -c: # this is a simple bash environment to delegate calls
      bash-c:
        hidden: true
```

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

This is a non exhaustive list of things to think about.

- [ ] Give precise config line numbers when a template rendering fails (#78)
- [ ] Add debugging messages for introspection.
- [X] Add help documentation for proxied commands (#77)
- [X] Explore ways to hook completions from proxied commands (#77)

## Acknowledgments

Built on the shoulders of giants.

- The summon library would not be possible without the excellent [Cobra](https://github.com/spf13/cobra)
library. Summon uses the dynamic command structure and completion offered
by Cobra.

- [go-yaml v3](https://github.com/go-yaml/yaml/tree/v3) Powers the polymorphic
nature of the yaml config file with its Node parsing API.
(And soon the exact template line numbers: #78).

- The [Masterminds Sprig Library](github.com/Masterminds/sprig/v3)
  allows doing amazing stuff in templates.

- The [Go Tree](github.com/DiSiqueira/GoTree) to present the asset tree like the
  shell `tree` command.

## FAQ

- Why not use git directly ?

  While you could use git directly to bring an asset directory with a simple git clone, the result does not have executable properties.

  In summon you leverage go execution to bootstrap in one phase. So your data can do:

  ```bash
  go run github.com/davidovich/summon-example-assets/summon@latest --help
  # or list the data deliverables
  go run github.com/davidovich/summon-example-assets/summon@latest ls
  ```

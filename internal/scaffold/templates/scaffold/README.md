Summon Data Provider Repo
=========================

This repository hosts a [{{ .SummonerName }}](https://github.com/davidovich/summon) data provider.

You install the latest version of this provider as `{{ .SummonerName }}`:

```shell
go install {{ .ModName }}/{{.SummonerName}}@latest
```

And then use the `{{ .SummonerName }}` executable to summon assets.

Summon `some-asset` like so:

```shell
{{ .SummonerName }} some-asset
```

By default, summon will instantiate the asset in the `.summoned/` directory and return its path. This can be overriden in the `asssets/summon.config.yaml` file or by using the `-o` flag.

Get more help with `{{ .SummonerName }} -h`.

Updating assets
---------------

1) Make modifications (additions, removals) in the `assets/` dir
2) Invoke `make`
3) Commit changes
4) Tag with a semantic version (prefix with `v`)
5) git push --tags origin HEAD

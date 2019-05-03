Summon Data provider repo
=========================

This repository hosts a [{{ .SummonerName }}](https://github.com/davidovich/summon) data
provider.

You install this provider as `{{ .SummonerName }}`:

```
go install {{ .ModName }}/{{.SummonerName}}
```

And then use the `{{ .SummonerName }}` executable to summon assets.

Summon `some-asset` like so:

```
{{ .SummonerName }} some-asset
```

By default, summon will instantiate the asset in the .summoned/ directory and return its path. This can be overriden in the `asssets/summon.config.yaml` file or the `-o` flag.

Get more help with `{{ .SummonerName }} -h`.
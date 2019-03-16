module github.com/davidovich/summon

require (
	github.com/davidovich/summon/pkg/scaffold v0.0.0-00010101000000-000000000000
	github.com/gobuffalo/packr/v2 v2.0.6
	github.com/lithammer/dedent v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/afero v1.2.1
	github.com/spf13/cobra v0.0.3
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.2.2
)

// local module is needed for packr which builds our embedded assets
// see https://stackoverflow.com/a/52302389/28275
replace github.com/davidovich/summon/pkg/scaffold => ./pkg/scaffold

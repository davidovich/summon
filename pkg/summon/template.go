package summon

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/cqroot/prompt"
)

func (d *Driver) prepareTemplate() (*template.Template, error) {
	t := d.templateCtx

	if t == nil {
		t = template.New(Name)
	} else {
		var err error
		t, err = t.Clone()
		if err != nil {
			return nil, err
		}
	}

	t.Option("missingkey=zero").
		Funcs(sprig.TxtFuncMap()).
		Funcs(summonFuncMap(d))

	return t, nil
}

func executeTemplate(t *template.Template, data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := t.Execute(buf, data)

	// The zero value for an interface is a nil interface{} which
	// has a string representation of <no value>. Strip this out.
	// https://github.com/golang/go/issues/24963
	return strings.ReplaceAll(buf.String(), "<no value>", ""), err
}

func (d *Driver) renderTemplate(tmpl string) (string, error) {
	t, err := d.prepareTemplate()
	if err != nil {
		return tmpl, err
	}

	t, err = t.Parse(tmpl)
	if err != nil {
		return tmpl, err
	}

	data := d.opts.data
	return executeTemplate(t, data)
}

func summonFuncMap(d *Driver) template.FuncMap {
	initConsumed := func() {
		if d.opts.argsConsumed == nil {
			d.opts.argsConsumed = make(map[int]struct{}, len(d.opts.args))
		}
	}
	consumeAllArgs := func() {
		initConsumed()
		for i := range d.opts.args {
			d.opts.argsConsumed[i] = struct{}{}
		}
	}
	return template.FuncMap{
		"run": func(args ...string) (string, error) {
			driverCopy := Driver{
				opts:        d.opts,
				config:      d.config,
				fs:          d.fs,
				baseDataDir: d.baseDataDir,
				templateCtx: d.templateCtx,
				execCommand: d.execCommand,
				configRead:  d.configRead,
				cmdToSpec:   d.cmdToSpec,
				prompts:     d.prompts,
				prompter:    d.prompter,
			}
			driverCopy.opts.argsConsumed = map[int]struct{}{}
			driverCopy.opts.cobraCmd = nil
			driverCopy.opts.helpWanted.helpFlag = ""

			b := &strings.Builder{}
			err := driverCopy.Run(Ref(args[0]), Args(args[1:]...), Out(b))

			if d.opts.dryrun {
				b.WriteString("[")
				b.WriteString(args[0])
				b.WriteString(" (dry-run)]")
			}

			if d.opts.debug {
				fmt.Fprintf(os.Stderr, "Output [%s] -> `%s`...\n", args[0], b)
			}
			return strings.TrimSpace(b.String()), err
		},
		"summon": func(path string, arg ...any) (string, error) {
			dest := os.TempDir()
			if len(arg) > 0 {
				if reflect.TypeOf(arg[0]).Kind() == reflect.String {
					dest = arg[0].(string)
				}
			}
			return d.Summon(Filename(path), Dest(dest))
		},
		"flagValue": func(flag string) (string, error) {
			for _, toRender := range d.flagsToRender {
				if toRender.name == flag {
					toRender.explicit = true
					return toRender.renderTemplate()
				}
			}
			return "", nil
		},
		"arg": func(index int, missingErrors ...string) (string, error) {
			missingError := strings.Join(missingErrors, " ")
			if d.opts.args == nil {
				return "", fmt.Errorf(missingError)
			}
			if index >= len(d.opts.args) {
				return "", fmt.Errorf("%s: index %v out of range, args: %s", missingError, index, d.opts.args)
			}

			retrieved := d.opts.args[index]
			initConsumed()
			d.opts.argsConsumed[index] = struct{}{}
			return retrieved, nil
		},
		"args": func() []string {
			consumeAllArgs()

			return d.opts.args
		},
		"swallowargs": func() string {
			consumeAllArgs()

			return ""
		},
		"prompt": func(slot, ask string, params any) (result string, err error) {
			defaultValue := ""
			selectors := []string{}

			switch t := params.(type) {
			case string:
				defaultValue = t
			case []any:
				for _, e := range t {
					selectors = append(selectors, e.(string))
				}
			default:
				return "", fmt.Errorf("last parameter should be a default value or a list of choices")
			}

			d.prompter.NewPrompt(ask)

			if len(selectors) != 0 {
				result, err = d.prompter.Choose(selectors)
			} else {
				result, err = d.prompter.Input(defaultValue)
			}

			if err != nil {
				return "", err
			}

			// record result for future use
			d.prompts[slot] = result
			return result, nil
		},
		"promptValue": func(slot string) (string, error) {
			p, ok := d.prompts[slot]
			if !ok {
				return "", fmt.Errorf("no previous prompts were filled for slot '%s'", slot)
			}
			return p, nil
		},
	}
}

type Prompt struct {
	pr         *prompt.Prompt
	promptStr  string
	selectors  []string
	defaultVal string
}

func (p *Prompt) NewPrompt(userPrompt string) {
	p.pr = prompt.New().Ask(userPrompt)
}

func (p *Prompt) Choose(choices []string) (string, error) {
	if p.pr == nil {
		return "", fmt.Errorf("prompter is not initialized")
	}

	return p.pr.Choose(choices)
}

func (p *Prompt) Input(defaultVal string) (string, error) {
	if p.pr == nil {
		return "", fmt.Errorf("prompter is not initialized")
	}
	return p.pr.Input(defaultVal)
}

package testutil

import (
	"github.com/spf13/afero"
)

// GetFs is a function that returns an Fs
var GetFs func() afero.Fs

// SetFs sets an app Fs
var SetFs func(fs afero.Fs)

// ReplaceFs replaces the real filesystem by a memory implementation
func ReplaceFs() func() {
	oldFs := GetFs()
	SetFs(afero.NewMemMapFs())
	return func() {
		SetFs(oldFs)
	}
}

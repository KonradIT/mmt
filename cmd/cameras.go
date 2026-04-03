package cmd

// Import camera backends for self-registration.
import (
	_ "github.com/konradit/mmt/pkg/android"
	_ "github.com/konradit/mmt/pkg/dji"
	_ "github.com/konradit/mmt/pkg/gopro"
	_ "github.com/konradit/mmt/pkg/insta360"
)

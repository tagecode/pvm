package cmd

import (
	"tagecode/pvm/internal/app"
)

// App groups PVM application services for a single PVM_HOME.
type App struct {
	Home     string
	Registry *app.Registry
	Aliases  *app.AliasStore
	Install  *app.Installer
	Switch   *app.Switcher
	Setup    *app.SetupService
}

func NewApp(home string) *App {
	return &App{
		Home:     home,
		Registry: app.NewRegistry(home),
		Aliases:  app.NewAliasStore(home),
		Install:  app.NewInstaller(home),
		Switch:   app.NewSwitcher(home),
		Setup:    app.NewSetupService(home),
	}
}

func (a *App) ResolveID(target string) (string, error) {
	if target == "default" {
		return a.Registry.ResolveID("default")
	}
	return a.Registry.ResolveID(target)
}

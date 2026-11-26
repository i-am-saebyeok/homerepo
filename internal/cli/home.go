package cli

import "github.com/i-am-saebyeok/homerepo/internal/home"

func (app App) loadHome() (home.Home, error) {
	wd, err := app.Getwd()
	if err != nil {
		return home.Home{}, err
	}
	root, err := home.Find(wd)
	if err != nil {
		return home.Home{}, err
	}
	return home.Load(root)
}

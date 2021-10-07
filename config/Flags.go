package config

import "flag"

// Flags provides a structure for storing flag-based options.
type Flags struct {
	Username, Password string
	Character          string
	Connect            string
	Fullscreen         bool
	GraphicsScale      float64
}

// Parse calls flag.Parse() on its fields.
func (f *Flags) Parse() {
	flag.StringVar(&f.Username, "username", "", "username")
	flag.StringVar(&f.Password, "password", "", "password")
	flag.StringVar(&f.Character, "character", "", "name of character")
	flag.StringVar(&f.Connect, "connect", "", "SERVER:PORT")
	flag.Float64Var(&f.GraphicsScale, "scale", 4, "graphics scaling")
	flag.BoolVar(&f.Fullscreen, "fullscreen", false, "fullscreen")
	flag.Parse()
}

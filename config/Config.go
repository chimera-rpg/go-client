package config

import (
	"fmt"
	"io/ioutil"

	"github.com/chimera-rpg/go-client/binds"
	"gopkg.in/yaml.v2"
)

// Config represents our global configuration.
type Config struct {
	Window     WindowConfig
	Game       GameConfig
	Servers    map[string]*ServerConfig
	LastServer string // Last server accessed by the client.
	path       string `yaml:"-"`
}

// Read attempts to parse the given YAML file and set it as the target path for saving.
func (c *Config) Read(p string) (err error) {
	c.path = p
	r, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(r, c); err != nil {
		return err
	}
	return nil
}

// AsYAMLString dumps the config as a yaml string.
func (c *Config) AsYAMLString() string {
	bytes, _ := yaml.Marshal(c)
	return string(bytes)
}

// Save writes the config to disk.
func (c *Config) Write() error {
	if c.path == "" {
		return fmt.Errorf("no config path defined")
	}
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(c.path, bytes, 0644)
	return err
}

// GameConfig is the configuration for the game state.
type GameConfig struct {
	Graphics      GameGraphicsConfig
	CommandPrefix string
	Bindings      binds.Bindings
	Containers    map[string]*ContainerConfig
}

// GameGraphicsConfig is the configuration for the game's graphics.
type GameGraphicsConfig struct {
	ObjectScale float64
}

// ServerConfig is the configuration for per-server settings.
type ServerConfig struct {
	Username         string
	Password         string
	Character        string
	RememberPassword bool
}

// WindowConfig is the configuration of the window's sizes.
type WindowConfig struct {
	Width, Height int
	X, Y          int
	Fullscreen    bool
}

// ContainerConfig is the config for a container, wow.
type ContainerConfig struct {
	Mode      int
	Aggregate bool
}

func (g *GameConfig) GetContainerConfig(which string) *ContainerConfig {
	if c, ok := g.Containers[which]; ok {
		return c
	}
	c := &ContainerConfig{}
	g.Containers[which] = c
	return c
}

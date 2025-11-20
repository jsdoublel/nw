package app

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

type config struct {
	Username    string           `toml:"username"`
	ApiKey      string           `toml:"api_key"`
	Features    featuresConfig   `toml:"features"`
	Appearance  appearanceConfig `toml:"appearance"`
	Keybinds    keybindConfig    `toml:"keybinds"`
	Directories directoryConfig  `toml:"directories"`
}

type featuresConfig struct {
	DisableDiscordRPC bool `toml:"discord_rpc"`
	AlwaysIncludeTMDB bool `toml:"always_include_tmdb"`
}

type appearanceConfig struct {
	Border        string       `toml:"border"` // rounded, normal, double
	ApplyBackdrop bool         `toml:"backdrop"`
	LightMode     bool         `toml:"light_mode"`
	Colors        colorPalette `toml:"colors"`
}

type colorPalette struct {
	Primary   string `toml:"primary"`
	Secondary string `toml:"secondary"`
	Success   string `toml:"success"`
	Error     string `toml:"error"`
}

type keybindConfig struct {
	Quit        []string `toml:"quit"`
	Delete      []string `toml:"delete"`
	Yes         []string `toml:"yes"`
	No          []string `toml:"no"`
	Left        []string `toml:"left"`
	Right       []string `toml:"right"`
	Up          []string `toml:"up"`
	Down        []string `toml:"down"`
	MoveLeft    []string `toml:"move_left"`
	MoveRight   []string `toml:"move_right"`
	MoveUp      []string `toml:"move_up"`
	MoveDown    []string `toml:"move_down"`
	AddList     []string `toml:"add_list"`
	SearchFilms []string `toml:"search_films"`
	Update      []string `toml:"update"`
	StopWatch   []string `toml:"stop_watch"`
	About       []string `toml:"about"`
}

type directoryConfig struct {
	Data    string `toml:"data"`
	Posters string `toml:"posters"`
}

var (
	Config    config
	ConfigErr error
)

func configInit() {
	var data []byte
	data, ConfigErr = os.ReadFile(ConfigPath())
	if ConfigErr != nil {
		return
	}
	_, ConfigErr = toml.Decode(string(data), &Config)
}

// Returns the expected config file path
func ConfigPath() string {
	return filepath.Join(xdg.ConfigHome, "nw", "nw.toml")
}

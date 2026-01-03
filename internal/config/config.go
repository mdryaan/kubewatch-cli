package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Namespace      string
	KubeConfig     string
	OutputFormat   string
	LabelSelector  string
	RefreshInterval int
	AllNamespaces  bool
	NoColor        bool
}

func Load() *Config {
	return &Config{
		Namespace:       viper.GetString("namespace"),
		KubeConfig:      viper.GetString("kubeconfig"),
		OutputFormat:    viper.GetString("output"),
		LabelSelector:   viper.GetString("selector"),
		RefreshInterval: viper.GetInt("interval"),
		AllNamespaces:   viper.GetBool("all-namespaces"),
		NoColor:         viper.GetBool("no-color"),
	}
}

func SetDefaults() {
	viper.SetDefault("namespace", DefaultNamespace)
	viper.SetDefault("output", DefaultOutputFormat)
	viper.SetDefault("interval", DefaultRefreshInterval)
	viper.SetDefault("all-namespaces", false)
	viper.SetDefault("no-color", false)
}

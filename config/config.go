package config

import (
	"log"
	"time"

	"github.com/11me/pulsy/monitor"
	"github.com/11me/pulsy/notifier"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type monitorConfig struct {
	URL      string        `mapstructure:"url"`
	Retry    int           `mapstructure:"retry"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type notifierConfig struct {
	Name    string                 `mapstructure:"name"`
	Options map[string]interface{} `mapstructure:"options"`
}

var monitorsConfig []*monitorConfig
var notifiersConfig []*notifierConfig

func ReadConfig() error {
	defaultCfgFile := "config.yaml"

	cmd := &cobra.Command{
		Use:   "pulsy",
		Short: "Pulsy is an open-source monitoring tool that keeps an eye on your critical systems",
		RunE: func(cmd *cobra.Command, args []string) error {
            return nil
		},
	}

	var cmdCfgFile string
	cmd.PersistentFlags().StringVarP(&cmdCfgFile, "config", "c", "", "config file (default is config.yaml)")

	if err := cmd.Execute(); err != nil {
		return err
	}

	viper.SetConfigType("yaml")
	if cmdCfgFile != "" {
		viper.SetConfigFile(cmdCfgFile)
	} else {
        log.Println("Using default", defaultCfgFile)
		viper.SetConfigFile(defaultCfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return err
		}
	}

	if err := viper.UnmarshalKey("monitors", &monitorsConfig); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("notifiers", &notifiersConfig); err != nil {
		return err
	}
	return nil
}

func LoadMonitors() []*monitor.Monitor {
	monitors := make([]*monitor.Monitor, 0, len(monitorsConfig))

	for _, m := range monitorsConfig {
		monitors = append(monitors, &monitor.Monitor{
			URL:      m.URL,
			Timeout:  m.Timeout,
			Retry:    m.Retry,
			Interval: m.Interval,
		})
	}
	return monitors
}

func LoadNotifiers() []notifier.Notifier {
	notifiers := make([]notifier.Notifier, 0, len(notifiersConfig))
	for _, n := range notifiersConfig {
		factory := notifier.MakeNotifierFactory(n.Name)
		notifiers = append(notifiers, factory(n.Options))
	}
	return notifiers
}


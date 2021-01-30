package main

import (
	"os"

	"github.com/regclient/regclient/regclient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const usageDesc = `Utility for accessing docker registries
More details at https://github.com/regclient/regclient`

var log *logrus.Logger

var rootCmd = &cobra.Command{
	Use:           "regctl <cmd>",
	Short:         "Utility for accessing docker registries",
	Long:          usageDesc,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var rootOpts struct {
	verbosity string
	logopts   []string
	format    string // for Go template formatting of various commands
}

func init() {
	log = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.WarnLevel,
	}
	rootCmd.PersistentFlags().StringVarP(&rootOpts.verbosity, "verbosity", "v", logrus.WarnLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().StringArrayVar(&rootOpts.logopts, "logopt", []string{}, "Log options")
	rootCmd.PersistentPreRunE = rootPreRun
}

func rootPreRun(cmd *cobra.Command, args []string) error {
	lvl, err := logrus.ParseLevel(rootOpts.verbosity)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	for _, opt := range rootOpts.logopts {
		if opt == "json" {
			log.Formatter = new(logrus.JSONFormatter)
		}
	}
	return nil
}

func newRegClient() regclient.RegClient {
	config, err := ConfigLoadDefault()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to load default config")
	} else {
		log.WithFields(logrus.Fields{
			"config": config,
		}).Debug("Loaded default config")
	}

	rcOpts := []regclient.Opt{regclient.WithLog(log)}
	if config.IncDockerCred == nil || *config.IncDockerCred {
		rcOpts = append(rcOpts, regclient.WithDockerCreds())
	}
	if config.IncDockerCert == nil || *config.IncDockerCert {
		rcOpts = append(rcOpts, regclient.WithDockerCerts())
	}

	rcHosts := []regclient.ConfigHost{}
	for name, host := range config.Hosts {
		rcHosts = append(rcHosts, regclient.ConfigHost{
			Name:    name,
			User:    host.User,
			Pass:    host.Pass,
			TLS:     host.TLS,
			Scheme:  host.Scheme,
			RegCert: host.RegCert,
		})
	}
	if len(rcHosts) > 0 {
		rcOpts = append(rcOpts, regclient.WithConfigHosts(rcHosts))
	}

	return regclient.NewRegClient(rcOpts...)
}

package main

import (
	"os"

	"github.com/barasher/http-trigger/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	retOk int = 0
	retKo int = 1
)

func main() {
	var confFile string
	cmd := &cobra.Command{
		Use: "http-trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServer(confFile)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().StringVarP(&confFile, "conf", "c", "", "Configuration file")
	cmd.MarkFlagRequired("conf")

	if err := cmd.Execute(); err != nil {
		log.Error().Msgf("%v", err)
		os.Exit(retKo)
	}
	os.Exit(retOk)

}

func startServer(confFile string) error {
	conf, err := internal.LoadConfiguration(confFile)
	if err != nil {
		return err
	}

	s, err := internal.NewServer(conf)
	if err != nil {
		return err
	}

	return s.Run()
}

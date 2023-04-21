package main

import (
	"github.com/guillembonet/backup/cmd/cli/commands"
	"github.com/rs/zerolog/log"
)

func main() {
	err := commands.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}

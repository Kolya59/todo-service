package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/psu/todo-service/pkg/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var opts struct {
	ServerHost   string `long:"server_host" env:"SERVER_HOST" description:"Server host" required:"true"`
	ServerPort   string `long:"server_port" env:"SERVER_PORT" description:"Server port" required:"true"`
	ProfilerPort string `long:"prof_port" env:"PROF_PORT" description:"Profiler port" required:"false"`
	LogLevel     string `long:"log_level" env:"LOG_LEVEL" description:"Log level for zerolog" required:"false"`
}

func main() {
	// Log initialization
	zerolog.MessageFieldName = "MESSAGE"
	zerolog.LevelFieldName = "LEVEL"
	zerolog.ErrorFieldName = "ERROR"
	zerolog.TimestampFieldName = "TIME"
	zerolog.CallerFieldName = "CALLER"
	log.Logger = log.Output(os.Stderr).With().Str("PROGRAM", "todo-service").Caller().Logger()

	// Parse flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Panic().Msgf("Could not parse flags: %v", err)
	}

	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	server.StartServer(opts.ServerHost, opts.ServerPort, opts.ProfilerPort)
}

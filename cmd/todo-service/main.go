package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Kolya59/todo-service/pkg/postgres"
	"github.com/Kolya59/todo-service/pkg/server"
)

var opts struct {
	ServerHost   string `long:"server_host" env:"SERVER_HOST" description:"Server host" required:"true"`
	ServerPort   string `long:"server_port" env:"SERVER_PORT" description:"Server port" required:"true"`
	DbHost       string `long:"database_host" env:"DB_HOST" description:"Database host" required:"true"`
	DbPort       string `long:"database_port" env:"DB_PORT" description:"Database port" required:"true"`
	DbName       string `long:"database_name" env:"DB_NAME" description:"Database name" required:"true"`
	DbUser       string `long:"database_username" env:"DB_USER" description:"Database username" required:"true"`
	DbPassword   string `long:"database_password" env:"DB_PASSWORD" description:"Database password" required:"true"`
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
		log.Fatal().Msgf("Could not parse flags: %v", err)
	}

	level, err := zerolog.ParseLevel(opts.LogLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	err = postgres.InitDatabaseConnection(opts.DbHost, opts.DbPort, opts.DbUser, opts.DbPassword, opts.DbName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set db connection")
	}
	defer postgres.CloseConnection()

	server.StartServer(opts.ServerHost, opts.ServerPort, opts.ProfilerPort)
}

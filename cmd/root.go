package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"x/core/internal/config"

	"github.com/iamolegga/enviper"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	serviceName         = "core"
	shutdownGracePeriod = time.Second * 15
	writeTimeout        = time.Second * 15
	readTimeout         = time.Second * 15
	idleTimeout         = time.Second * 60
	envConfigPrefix     = "core"
	dbDriver            = "postgres"
)

var (
	cfgFile string
	conf    config.Config
	z       zerolog.Logger
)

// rootCmd represents the base comand when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   serviceName,
	Short: "all things core",
	Long:  "Core houses core api for Fragrance Exchange",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Root command executed")
		// Add any default logic here, or leave this as a placeholder
	},
}

// Execute adds all child commands to the root command and sets flags appropriately
// This is called by main.main(), it only needs to happen once to the rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	fmt.Println("running init...")
	cobra.OnInitialize(initConfig, initLogger)

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() {
	e := enviper.New(viper.New())

	if cfgFile != "" {
		e.SetConfigFile(cfgFile)
		fmt.Printf("Using config file: %s\n", cfgFile)
	} else {
		home, err := os.Getwd()
		if err != nil {
			fmt.Printf("%v", err.Error())
			os.Exit(1)
		}

		e.SetEnvPrefix("CORE")
		e.AddConfigPath(home + "/conf/local")
		e.SetConfigName("conf")
		e.SetConfigType("env")
	}

	e.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	e.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := e.ReadInConfig(); err != nil {
		fmt.Printf("\nError reading config file: %s\n", err)
		fmt.Printf("\nError reading config file: %s\n", e.ConfigFileUsed())
	} else {
		fmt.Printf("Config file loaded: %s\n", e.ConfigFileUsed())
	}

	fmt.Printf("config: %s", e.ConfigFileUsed())

	if err := e.Unmarshal(&conf); err != nil {
		fmt.Printf("\nError unmarshalling config file: %s\n", err.Error())
		fmt.Printf("\nError unmarshalling config file: %s\n", e.ConfigFileUsed())
	} else {
		fmt.Println("Config unmarshalled successfully!")
	}

	if conf.Env == "local" {
		defer fmt.Printf("parsed local config\n")
		bytes, _ := json.MarshalIndent(conf, "", " ")
		fmt.Printf("\n%s\n", string(bytes))
	}

	if conf.Env == "development" {
		fmt.Printf("parsed development config\n")
	}
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.ErrorFieldName = "error.message"

	z = zerolog.New(os.Stdout).
		With().
		Str("service", serviceName).
		Timestamp().
		Caller().
		Logger()

	log.Logger = z
}

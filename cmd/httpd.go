package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"x/core/internal/handlers"
	"x/core/internal/persist"
	"x/core/internal/service"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type AllowedMethods string

var (
	AllowedMethodsGet    AllowedMethods = "GET"
	AllowedMethodsPost   AllowedMethods = "POST"
	AllowedMethodsPatch  AllowedMethods = "PATCH"
	AllowedMethodsPut    AllowedMethods = "PUT"
	AllowedMethodsDelete AllowedMethods = "DELETE"
)

var httpCmd = &cobra.Command{
	Use:   "httpd",
	Short: "http service",
	Long:  "handles all http requests",
	Run:   runHTTPServer,
}

func init() {
	rootCmd.AddCommand(httpCmd)
}

func runHTTPServer(cmd *cobra.Command, args []string) {
	// conf validation
	if conf.HTTPServer.ServerPort == "" {
		z.Fatal().Msg("Server port is not configured")
	}
	if conf.Clerk.APIKey == "" {
		z.Fatal().Msg("Clerk API key is not configured")
	}
	if conf.Cloudinary.APIKey == "" {
		z.Fatal().Msg("Cloudinary API key is not configured")
	}

	// Parent context
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer cancel()

	// Database connection
	db, err := ConnectToDB(
		ctx,
		conf.DB,
	)
	if err != nil {
		z.Fatal().Err(err).Msgf("error connecting to database: %s", err)
	}

	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			z.Error().Err(err).Msgf("failed to parse sql db")
		}

		if err := sqlDB.Close(); err != nil {
			z.Error().Err(err).Msgf("failed to close database connection")
		} else {
			z.Info().Msg("database connection successfully closed")
		}
	}()

	// Sentry monitoring
	if err := sentry.Init(sentry.ClientOptions{
		EnableTracing:      true,
		Dsn:                conf.Sentry.DSN,
		TracesSampleRate:   conf.Sentry.SampleRate,
		ProfilesSampleRate: conf.Sentry.SampleRate,
		Environment:        conf.Env,
	}); err != nil {
		z.Error().Err(err).Msgf("error initializing sentry monitoring")
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	// Create Sentry Handler for use as http middleware
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,  // Repanic after capturing
		WaitForDelivery: false, // Don't block requests to send events
		Timeout:         2 * time.Second,
	})

	z.Info().Msg("sentry monitoring handler initialized")

	// Clerk (3rd party auth)
	clerkClient, err := clerk.NewClient(conf.Clerk.APIKey)
	if err != nil {
		z.Fatal().Err(err).Msgf("error connecting to clerk auth client: %s", err)
	}
	z.Info().Msg("clerk client initialized")

	// Cloudinary (3rd party images persistence)
	cld, err := cloudinary.NewFromURL(conf.Cloudinary.APIKey)
	if err != nil {
		z.Fatal().Err(err).Msgf("error connecting to cloudinary image store: %s", err)
	}
	cld.Config.URL.Secure = true
	z.Info().Msg("cloudinary client initialized")

	// Initialize store
	store := persist.NewPGStore(
		db,
		&log.Logger,
	)
	z.Info().Msg("core database initialized")

	// Initalize service
	service := service.NewService(
		store,
		&z,
		cld,
	)
	z.Info().Msg("core service initialized")

	// Initialize cors
	crossOrigin := cors.New(cors.Options{
		AllowedOrigins: []string{conf.HTTPServer.ServerAllowedOriginLocal, conf.HTTPServer.ServerAllowedOriginProd},
		AllowedMethods: []string{
			string(AllowedMethodsGet),
			string(AllowedMethodsPost),
			string(AllowedMethodsPatch),
			string(AllowedMethodsPut),
			string(AllowedMethodsDelete)},
		AllowCredentials: true,
	})
	z.Info().Msg("cors policy initialized")

	// Initialize handler
	h := handlers.NewHandler(
		&z,
		service,
		clerkClient,
		conf,
		sentryHandler,
	)
	router := h.RegisterRoutes()
	handler := crossOrigin.Handler(router)
	z.Info().Msg("core handler initialized")

	// Configure server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", conf.HTTPServer.ServerPort),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      handler,
	}
	z.Info().Msg("server configuration successful")

	// Initialize web server
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", shutdownGracePeriod, "duration for which the server gracefully waits for existing connecitosn to finish")
	flag.Parse()

	// run server in goroutine to prevent blocking
	go func() {
		z.Info().Msgf("core service running on port :%s", conf.HTTPServer.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			z.Error().Err(err).Msgf("unexpected server error")
		}
	}()

	// Block until we recieve our signal
	<-ctx.Done()
	z.Info().Msg("received shutdown signal, shutting down core service gracefully")

	// Create a deadline to wait for
	cx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline
	if err := server.Shutdown(cx); err != nil {
		z.Error().Err(err).Msg("error during server shutdown")
	}

	z.Info().Msg("core service successfully shutdown")
	os.Exit(0)
}

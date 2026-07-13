package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hkalexling/mango-go/internal/config"
	"github.com/hkalexling/mango-go/internal/library"
	"github.com/hkalexling/mango-go/internal/queue"
	"github.com/hkalexling/mango-go/internal/server"
	"github.com/hkalexling/mango-go/internal/storage"
	"github.com/hkalexling/mango-go/internal/tasks"
	"github.com/hkalexling/mango-go/web"
	"github.com/spf13/cobra"
)

func loadTemplates() (*server.TemplateManager, error) {
	return server.NewTemplateManager(web.Views())
}

const version = "2.0.0"

const banner = `

              _|      _|
              _|_|  _|_|    _|_|_|  _|_|_|      _|_|_|    _|_|
              _|  _|  _|  _|    _|  _|    _|  _|    _|  _|    _|
              _|      _|  _|    _|  _|    _|  _|    _|  _|    _|
              _|      _|    _|_|_|  _|    _|    _|_|_|    _|_|
                                                    _|
                                                _|_|

`

func main() {
	var configPath string

	root := &cobra.Command{
		Use:     "mango",
		Short:   fmt.Sprintf("Mango - Manga Server and Web Reader. Version %s", version),
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(banner)
			fmt.Printf("Mango - Manga Server and Web Reader. Version %s\n\n", version)

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			cfg.SetCurrent()

			st, err := storage.Open(cfg.DBPath, cfg.LibraryPath)
			if err != nil {
				return err
			}
			defer st.Close()

			ver, _ := st.Version()
			fmt.Printf("Config loaded from %s\n", cfg.Path())
			fmt.Printf("Database ready at %s (schema version %d)\n", cfg.DBPath, ver)

			lib := library.NewLibrary(cfg.LibraryPath, st, cfg.LibraryCachePath)
			if err := lib.LoadFromCache(); err != nil {
				fmt.Printf("Library cache not loaded: %v\n", err)
			}

			// Initialize download queue
			queueDB, err := queue.NewQueue(cfg.QueueDBPath)
			if err != nil {
				return fmt.Errorf("init queue: %w", err)
			}
			defer queueDB.Close()

			// Start background tasks (scan runs async; does not block HTTP)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				<-sigCh
				fmt.Println("\nShutting down...")
				cancel()
			}()

			runner := tasks.NewRunner(lib, cfg.ScanIntervalMinutes, cfg.ThumbnailGenerationIntervalHours)
			runner.SetPluginTasks(queueDB, cfg.PluginPath, cfg.LibraryPath, cfg.PluginUpdateIntervalHours)
			go runner.Start(ctx)

			tm, err := loadTemplates()
			if err != nil {
				return fmt.Errorf("load templates: %w", err)
			}

			deps := &server.Dependencies{
				Config:    cfg,
				Storage:   st,
				Library:   lib,
				Queue:     queueDB,
				Runner:    runner,
				Templates: tm,
			}

			srv := server.NewServer(deps)
			srv.RegisterRoutes()

			fmt.Println("Server starting...")
			if err := srv.Start(ctx); err != nil {
				return err
			}

			return nil
		},
	}
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to the config file")

	root.AddCommand(newAdminCmd(&configPath))

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

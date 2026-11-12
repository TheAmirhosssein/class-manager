package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/TheAmirhosssein/class-manager/config"
	accountsrouter "github.com/TheAmirhosssein/class-manager/internal/apps/accounts/delivery/http/router"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/params"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/repository"
	"github.com/TheAmirhosssein/class-manager/internal/apps/accounts/usecase"
	classmanagerrouter "github.com/TheAmirhosssein/class-manager/internal/apps/classmanager/delivery/http/router"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/database"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/hash"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/jwt"
	"github.com/TheAmirhosssein/class-manager/internal/infrastructure/server"
	"github.com/TheAmirhosssein/class-manager/internal/seed"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func main() {
	root := &cobra.Command{
		Use:   "classino",
		Short: "کلاسینو — class management service",
	}

	root.AddCommand(serveCmd(), migrateCmd(), seedCmd(), createAdminCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server (infrastructure bootstrap)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := config.GetConfig()
			db, err := database.Open(cfg.Database)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := database.MigrateUp(db); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}

			if cfg.JWT.Secret == "" {
				return fmt.Errorf("JWT_SECRET is required (set in .env or jwt.secret)")
			}
			tokens, err := jwt.NewService(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.App.Name)
			if err != nil {
				return err
			}

			engine := gin.New()
			engine.Use(gin.Recovery())
			server.RegisterHealth(engine, db)
			accountsrouter.Register(engine, cfg, db, tokens)
			classmanagerrouter.Register(engine, cfg, db, tokens)
			server.RegisterFrontend(engine, cfg.Frontend.Dir, accountsrouter.EntryPath(cfg))

			srv := &http.Server{
				Addr:              fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
				Handler:           engine,
				ReadHeaderTimeout: 10 * time.Second,
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			errCh := make(chan error, 1)
			go func() {
				log.Printf("listening on http://%s", srv.Addr)
				errCh <- srv.ListenAndServe()
			}()

			select {
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				return srv.Shutdown(shutdownCtx)
			case err := <-errCh:
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			}
		},
	}
}

func migrateCmd() *cobra.Command {
	var down bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply or roll back SQLite migrations",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.GetConfig()
			db, err := database.Open(cfg.Database)
			if err != nil {
				return err
			}
			defer db.Close()
			if down {
				return database.MigrateDown(db)
			}
			return database.MigrateUp(db)
		},
	}
	cmd.Flags().BoolVar(&down, "down", false, "roll back the last migration")
	return cmd
}

func seedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Insert development teacher, admin, and sample school",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.GetConfig()
			db, err := database.Open(cfg.Database)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := database.MigrateUp(db); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}
			if err := seed.Create(context.Background(), db); err != nil {
				return err
			}

			entry := accountsrouter.EntryPath(cfg)
			fmt.Println("Seed data ready.")
			fmt.Printf("  Teacher login: %s / %s\n", seed.TeacherMobile, seed.TeacherPassword)
			fmt.Printf("  Admin login:   %s / %s\n", seed.AdminMobile, seed.AdminPassword)
			fmt.Printf("  Admin console: /%s or #/%s\n", entry, entry)
			return nil
		},
	}
}

func createAdminCmd() *cobra.Command {
	var mobile, name, password string
	cmd := &cobra.Command{
		Use:   "create-admin",
		Short: "Create a new admin account",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.GetConfig()
			db, err := database.Open(cfg.Database)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := database.MigrateUp(db); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}

			reader := bufio.NewReader(os.Stdin)
			if strings.TrimSpace(mobile) == "" {
				fmt.Fprint(os.Stdout, "Mobile (09xxxxxxxxx): ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				mobile = strings.TrimSpace(line)
			}
			if strings.TrimSpace(name) == "" {
				fmt.Fprint(os.Stdout, "Name: ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				name = strings.TrimSpace(line)
			}
			if password == "" {
				fmt.Fprint(os.Stdout, "Password: ")
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))
				fmt.Fprintln(os.Stdout)
				if err != nil {
					return err
				}
				password = strings.TrimSpace(string(bytePassword))
			}
			if password == "" {
				return fmt.Errorf("password is required")
			}

			h := hash.New()
			adminUC := usecase.NewAdminUsecase(repository.NewAdminRepository(db), h)
			_, msg := adminUC.Create(context.Background(), params.CreateAdminParams{
				Mobile:   mobile,
				Password: password,
				Name:     name,
			})
			if msg.Error != nil {
				return fmt.Errorf("%s", msg.Description)
			}
			fmt.Println(msg.Description)
			fmt.Printf("Admin console: /%s or #/%s\n", accountsrouter.EntryPath(cfg), accountsrouter.EntryPath(cfg))
			return nil
		},
	}
	cmd.Flags().StringVar(&mobile, "mobile", "", "admin mobile (09xxxxxxxxx)")
	cmd.Flags().StringVar(&name, "name", "", "display name")
	cmd.Flags().StringVar(&password, "password", "", "password (omit to read from terminal)")
	return cmd
}

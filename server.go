package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/api"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/service"
	"github.com/nickcoast/timetravel/sqlite"
	"github.com/pelletier/go-toml"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	fmt.Println("func main start")
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	m := NewMain()
	// Execute program.
	fmt.Println("func main Run")
	if err := m.Run(ctx); err != nil {
		m.Close()
		fmt.Fprintln(os.Stderr, err)
		entity.ReportError(ctx, err)
		os.Exit(1)
	}
	// Wait for CTRL-C.
	fmt.Println("func main end")
	<-ctx.Done()

	// Clean up program.
	if err := m.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	Config     Config
	ConfigPath string
	DB         *sqlite.DB
	HTTPServer *http.Server

	InsuredService entity.InsuredService
}

//func NewServer()

// Close gracefully stops the program.
func (m *Main) Close() error {
	if m.HTTPServer != nil {
		if err := m.HTTPServer.Close(); err != nil {
			return err
		}
	}
	if m.DB != nil {
		if err := m.DB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func NewMain() *Main {
	fmt.Println("Main NewMain")
	router := mux.NewRouter()

	memoryService := service.NewInMemoryRecordService()
	sqliteService := service.NewSqliteRecordService()
	api := api.NewAPI(&memoryService, &sqliteService)

	apiRoute := router.PathPrefix("/api/v2").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	api.CreateRoutes(apiRoute)

	address := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("listening on %s", address)
	//log.Fatal(srv.ListenAndServe())
	fmt.Println("Main NewMain after ListenAndServe")

	db := sqlite.NewDB("file:main.db?cache=shared&mode=rwc&locking_mode=NORMAL&_fk=1&synchronous=2")
	// injects db into the sqlite service so it can interact with the database.
	sqliteService.SetService(db)

	return &Main{
		Config:     DefaultConfig(),
		ConfigPath: DefaultConfigPath,
		DB:         db,
		HTTPServer: srv,
	}
}

// Run executes the program. The configuration should already be set up before
// calling this function.
func (m *Main) Run(ctx context.Context) (err error) {
	fmt.Println("Main Run")
	// Expand the DSN (in case it is in the user home directory ("~")).
	// Then open the database. This will instantiate the SQLite connection
	// and execute any pending migration files.
	if m.DB.DSN, err = expandDSN(m.Config.DB.DSN); err != nil {
		fmt.Println("Main.Run ERR m.DB.DSN", m.DB.DSN)
		return fmt.Errorf("cannot expand dsn: %w", err)
	}
	if err := m.DB.Open(); err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}
	fmt.Println("Main.Run after m.DB.Open. m.DB.DSN", m.DB.DSN)

	// Instantiate SQLite-backed services.
	insuredService := sqlite.NewInsuredService(m.DB)

	// Attach insured service to Main for testing.
	m.InsuredService = insuredService

	// Copy configuration settings to the HTTP server.
	/* 	m.HTTPServer.Addr = m.Config.HTTP.Addr
	   	m.HTTPServer.Domain = m.Config.HTTP.Domain
	   	m.HTTPServer.HashKey = m.Config.HTTP.HashKey
	   	m.HTTPServer.BlockKey = m.Config.HTTP.BlockKey
	   	m.HTTPServer.GitHubClientID = m.Config.GitHub.ClientID
	   	m.HTTPServer.GitHubClientSecret = m.Config.GitHub.ClientSecret */

	// Attach underlying services to the HTTP server.
	/* 	m.HTTPServer.AuthService = authService
	   	m.HTTPServer.DialService = dialService
	   	m.HTTPServer.DialMembershipService = dialMembershipService
	   	m.HTTPServer.EventService = eventService
	   	m.HTTPServer.UserService = userService */

	// Start the HTTP server.
	/* if err := m.HTTPServer.Open(); err != nil {
		return err
	} */

	//m.HTTPServer.InsuredService = insuredService
	//log.Fatal(m.HTTPServer.ListenAndServe())
	go func() { log.Fatal(m.HTTPServer.ListenAndServe()) }()
	fmt.Println("Server started in Main.Run")
	//log.Printf("running: url=%q debug=http://localhost:6060 dsn=%q", m.HTTPServer.URL(), m.Config.DB.DSN)

	return nil
}

const (
	// DefaultConfigPath is the default path to the application configuration.
	DefaultConfigPath = "~/code/go/temelpa/wtfd.conf"

	// DefaultDSN is the default datasource name.
	DefaultDSN = "file:main.db?cache=shared&mode=rwc&locking_mode=NORMAL&_fk=1&synchronous=2"
)

// Config represents the CLI configuration file.
type Config struct {
	DB struct {
		DSN string `toml:"dsn"`
	} `toml:"db"`

	HTTP struct {
		Addr     string `toml:"addr"`
		Domain   string `toml:"domain"`
		HashKey  string `toml:"hash-key"`
		BlockKey string `toml:"block-key"`
	} `toml:"http"`
}

// DefaultConfig returns a new instance of Config with defaults set.
func DefaultConfig() Config {
	var config Config
	config.DB.DSN = DefaultDSN
	return config
}

// ReadConfigFile unmarshals config from
func ReadConfigFile(filename string) (Config, error) {
	config := DefaultConfig()
	if buf, err := ioutil.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

// expand returns path using tilde expansion. This means that a file path that
// begins with the "~" will be expanded to prefix the user's home directory.
func expand(path string) (string, error) {
	// Ignore if path has no leading tilde.
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path, nil
	}

	// Fetch the current user to determine the home path.
	u, err := user.Current()
	if err != nil {
		return path, err
	} else if u.HomeDir == "" {
		return path, fmt.Errorf("home directory unset")
	}

	if path == "~" {
		return u.HomeDir, nil
	}
	return filepath.Join(u.HomeDir, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
}

// expandDSN expands a datasource name. Ignores in-memory databases.
func expandDSN(dsn string) (string, error) {
	if dsn == ":memory:" {
		return dsn, nil
	}
	return expand(dsn)
}

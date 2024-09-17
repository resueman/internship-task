package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tender-management-api/internal/controller"
	"tender-management-api/internal/repo"
	"tender-management-api/internal/service"
	"tender-management-api/pkg/http_server"
	"tender-management-api/pkg/postgres"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/labstack/echo"
)

func usersAndOrganizationsDatabasesExist(pg *postgres.Postgres) (bool, error) {
	if err := pg.Database.Ping(); err != nil {
		return false, err
	}

	var id uuid.UUID
	err := pg.Database.QueryRow("select id from organization_responsible").Scan(&id)

	return err == nil, nil
}

func tenderAndBidDatabasesExist(pg *postgres.Postgres) (bool, error) {
	if err := pg.Database.Ping(); err != nil {
		return false, err
	}

	var id uuid.UUID
	err := pg.Database.QueryRow("select id from tender").Scan(&id)

	return err == nil, nil
}

func migrateTables(driver database.Driver, sourceUrl string, databaseName string) {
	migrations, err := migrate.NewWithDatabaseInstance(sourceUrl, databaseName, driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := migrations.Up(); err != nil {
		if err.Error() == "no change" {
			log.Println("no change made by migration scripts")
		} else {
			log.Fatal(err)
		}
	}
}

func runMigrations(postgresDB *postgres.Postgres, driver database.Driver, databaseName string) {
	userOrganizationTablesExist, err := usersAndOrganizationsDatabasesExist(postgresDB)
	if err != nil {
		log.Fatal(err)
	}

	if !userOrganizationTablesExist {
		migrateTables(driver, "file://migrations/user-organization-migrations", databaseName)

		return
	}
	tenderTablesExist, err := tenderAndBidDatabasesExist(postgresDB)
	if err != nil {
		log.Fatal(err)
	}
	if !tenderTablesExist {
		migrateTables(driver, "file://migrations/tender-bid-migrations", databaseName)
	}
}

func Run() {
	serverAddreeEnv := os.Getenv("SERVER_ADDRESS")
	dbConnEnv := os.Getenv("POSTGRES_CONN")
	_ = os.Getenv("POSTGRES_JDBC_URL")
	dbUsernameEnv := os.Getenv("POSTGRES_USERNAME")
	dbPasswordEnv := os.Getenv("POSTGRES_PASSWORD")
	dbHostEnv := os.Getenv("POSTGRES_HOST")
	dbPortEnv := os.Getenv("POSTGRES_PORT")
	databaseEnv := os.Getenv("POSTGRES_DATABASE")

	// export POSTGRESQL_URL='postgres://postgres:password@localhost:5432/example?sslmode=disable&search_path=public'
	url := dbConnEnv
	_ = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHostEnv, dbUsernameEnv, dbPasswordEnv, databaseEnv, dbPortEnv)

	log.Println("Connecting database...")
	postgresDB, err := postgres.NewDB(url)

	if err != nil {
		postgresDB, err = postgres.NewDB(dbConnEnv)
		if err != nil {
			log.Fatal("Error occurred while connecting to db: %w", err)
		}
	}
	defer postgresDB.Close()

	log.Println("Running migrations...")
	driver, err := pgmigrate.WithInstance(postgresDB.Database, &pgmigrate.Config{DatabaseName: databaseEnv})
	if err != nil {
		log.Fatal(err)
	}
	runMigrations(postgresDB, driver, databaseEnv)

	repositories := repo.NewRepositories(postgresDB)
	services := service.NewServices(repositories)
	handler := echo.New()

	log.Println("Setup routes...")
	controller.SetupRoutesHandlers(handler, services)

	log.Println("Starting server...")
	httpServer := http_server.New(handler, serverAddreeEnv)

	log.Println("Ready to process requests...")

	log.Println("Graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Println("Got signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Fatal("Notify error: %w", err)
	}

	log.Println("Shutting down...")
	err = httpServer.Shutdown()
	if err != nil {
		log.Fatal("Shutdown error: %w", err)
	} else {
		log.Println("Successful shutdown")
	}
}

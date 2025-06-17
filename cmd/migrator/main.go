package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"slices"
)

func main() {
	// Command line arguments
	var (
		dbConnection    string
		migrationsPath  string
		migrationsTable string
		verbose         bool
		direction       string
	)

	flag.StringVar(&dbConnection, "db", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.StringVar(&migrationsPath, "migrations-path", "./db/migrations", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "schema_migrations", "name of migrations table")
	flag.BoolVar(&verbose, "verbose", false, "show verbose output")
	flag.StringVar(&direction, "direction", "up", "migration direction: up or down")
	flag.Parse()

	// Validate required params
	if dbConnection == "" {
		log.Fatal("Database connection string is required. Provide it with -db flag or DATABASE_URL environment variable")
	}
	if migrationsPath == "" {
		log.Fatal("Migrations path is required")
	}

	// Validate direction
	direction = strings.ToLower(direction)
	if direction != "up" && direction != "down" {
		log.Fatal("Direction must be either 'up' or 'down'")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the database
	conn, err := pgx.Connect(ctx, dbConnection)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Create the migrations table if it doesn't exist
	if err := ensureMigrationsTable(ctx, conn, migrationsTable); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	if direction == "up" {
		runUpMigrations(ctx, conn, migrationsPath, migrationsTable, verbose)
	} else {
		runDownMigrations(ctx, conn, migrationsPath, migrationsTable, verbose)
	}
}

// runUpMigrations applies pending UP migrations
func runUpMigrations(ctx context.Context, conn *pgx.Conn, migrationsPath, migrationsTable string, verbose bool) {
	// Get all UP migration files
	migrationFiles, err := getMigrationFiles(migrationsPath, "up")
	if err != nil {
		log.Fatalf("Failed to read migration files: %v", err)
	}

	if len(migrationFiles) == 0 {
		fmt.Println("No UP migration files found")
		return
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(ctx, conn, migrationsTable)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	// Apply pending migrations
	count := 0
	for _, file := range migrationFiles {
		// Skip if already applied
		if containsMigration(appliedMigrations, file.Name) {
			if verbose {
				fmt.Printf("Skipping already applied migration: %s\n", file.Name)
			}
			continue
		}

		// Apply migration
		if verbose {
			fmt.Printf("Applying UP migration: %s\n", file.Name)
		}

		if err := applyMigration(ctx, conn, migrationsPath, migrationsTable, file.Name, true); err != nil {
			log.Fatalf("Failed to apply migration %s: %v", file.Name, err)
		}

		count++
		fmt.Printf("Applied UP migration: %s\n", file.Name)
	}

	if count == 0 {
		fmt.Println("No UP migrations to apply")
	} else {
		fmt.Printf("Successfully applied %d UP migrations\n", count)
	}
}

// runDownMigrations applies DOWN migrations (rollback)
func runDownMigrations(ctx context.Context, conn *pgx.Conn, migrationsPath, migrationsTable string, verbose bool) {
	// Get all DOWN migration files
	migrationFiles, err := getMigrationFiles(migrationsPath, "down")
	if err != nil {
		log.Fatalf("Failed to read migration files: %v", err)
	}

	if len(migrationFiles) == 0 {
		fmt.Println("No DOWN migration files found")
		return
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(ctx, conn, migrationsTable)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	// For DOWN migrations, we need to reverse the order and only apply those that have been applied
	// Sort in reverse order for rollback
	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Name > migrationFiles[j].Name
	})

	count := 0
	for _, file := range migrationFiles {
		// Convert DOWN filename to UP filename to check if it was applied
		upFileName := strings.Replace(file.Name, "_down.sql", "_up.sql", 1)
		
		// Skip if not applied (can't rollback what wasn't applied)
		if !containsMigration(appliedMigrations, upFileName) {
			if verbose {
				fmt.Printf("Skipping non-applied migration: %s (UP version %s not found)\n", file.Name, upFileName)
			}
			continue
		}

		// Apply DOWN migration
		if verbose {
			fmt.Printf("Applying DOWN migration: %s\n", file.Name)
		}

		if err := applyMigration(ctx, conn, migrationsPath, migrationsTable, file.Name, false); err != nil {
			log.Fatalf("Failed to apply DOWN migration %s: %v", file.Name, err)
		}

		count++
		fmt.Printf("Applied DOWN migration: %s\n", file.Name)
	}

	if count == 0 {
		fmt.Println("No DOWN migrations to apply")
	} else {
		fmt.Printf("Successfully applied %d DOWN migrations\n", count)
	}
}

// applyMigration applies a single migration and records it
func applyMigration(ctx context.Context, conn *pgx.Conn, migrationsPath, migrationsTable, fileName string, isUp bool) error {
	// Read migration file
	content, err := os.ReadFile(filepath.Join(migrationsPath, fileName))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Execute migration
	if _, err := tx.Exec(ctx, string(content)); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	// Record or remove migration record
	if isUp {
		// Record UP migration
		if _, err := tx.Exec(ctx,
			fmt.Sprintf("INSERT INTO %s (version, applied_at) VALUES ($1, NOW())", migrationsTable),
			fileName); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration: %v", err)
		}
	} else {
		// Remove DOWN migration record (remove the corresponding UP migration)
		upFileName := strings.Replace(fileName, "_down.sql", "_up.sql", 1)
		if _, err := tx.Exec(ctx,
			fmt.Sprintf("DELETE FROM %s WHERE version = $1", migrationsTable),
			upFileName); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to remove migration record: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Migration represents a migration file
type Migration struct {
	Name string
	Path string
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
func ensureMigrationsTable(ctx context.Context, conn *pgx.Conn, tableName string) error {
	_, err := conn.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`, tableName))
	return err
}

// getMigrationFiles returns a sorted list of migration files for the specified direction
func getMigrationFiles(path, direction string) ([]Migration, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []Migration
	suffix := fmt.Sprintf("_%s.sql", direction)
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		// Only consider files with the correct suffix (e.g., _up.sql or _down.sql)
		if !strings.HasSuffix(strings.ToLower(name), suffix) {
			continue
		}
		
		files = append(files, Migration{
			Name: name,
			Path: filepath.Join(path, name),
		})
	}

	// Sort files by name to ensure they're applied in the right order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// getAppliedMigrations returns a list of already applied migrations
func getAppliedMigrations(ctx context.Context, conn *pgx.Conn, tableName string) ([]string, error) {
	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT version FROM %s ORDER BY version", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// containsMigration checks if a migration has already been applied
func containsMigration(migrations []string, name string) bool {
	return slices.Contains(migrations, name)
}
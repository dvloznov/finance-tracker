package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

//go:embed bigquery/*.sql
var migrationFS embed.FS

// Migration represents a single migration file
type Migration struct {
	Version  int
	Name     string
	Filename string
	SQL      string
	Checksum string
}

// AppliedMigration represents a migration that has already been applied
type AppliedMigration struct {
	Version   int
	Name      string
	AppliedAt time.Time
	Checksum  string
	AppliedBy string
}

const (
	ensureSchemaMigrationsTableQuery = `
		CREATE TABLE IF NOT EXISTS finance.schema_migrations (
			version       INT64 NOT NULL,
			name          STRING NOT NULL,
			applied_at    TIMESTAMP NOT NULL,
			checksum      STRING,
			applied_by    STRING
		)`

	appliedMigrationsQuery = `
		SELECT version, name, applied_at, checksum, applied_by
		FROM finance.schema_migrations
		ORDER BY version ASC`

	recordMigrationQuery = `
		INSERT INTO finance.schema_migrations
		(version, name, applied_at, checksum, applied_by)
		VALUES (@version, @name, CURRENT_TIMESTAMP(), @checksum, @applied_by)`
)

var (
	projectID = flag.String("project", defaultProject(), "GCP project ID (or BQ_PROJECT_ID env var)")
	appliedBy = flag.String("applied-by", "migrate-cli", "Name of the tool applying migrations")
)

func defaultProject() string {
	if v := os.Getenv("BQ_PROJECT_ID"); v != "" {
		return v
	}
	if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		return v
	}
	return ""
}

func main() {
	flag.Parse()

	ctx := context.Background()

	// Validate required flags
	if *projectID == "" {
		log.Fatal("Error: -project flag is required. Please specify your GCP project ID.")
	}

	// Create BigQuery client
	client, err := bigquery.NewClient(ctx, *projectID)
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Ensure schema_migrations table exists
	if err := ensureSchemaMigrationsTable(ctx, client); err != nil {
		log.Fatalf("Failed to ensure schema_migrations table: %v", err)
	}

	// Read migration files
	migrations, err := readMigrations()
	if err != nil {
		log.Fatalf("Failed to read migrations: %v", err)
	}

	log.Printf("Found %d migration files", len(migrations))

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(ctx, client)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	log.Printf("Found %d already applied migrations", len(appliedMigrations))

	// Build map of applied versions
	appliedVersions := make(map[int]string)
	for _, am := range appliedMigrations {
		appliedVersions[am.Version] = am.Checksum
	}

	// Apply pending migrations
	appliedCount := 0
	for _, migration := range migrations {
		if v, ok := appliedVersions[migration.Version]; ok {
			if v == migration.Checksum {
				log.Printf("  [SKIP] %04d_%s (already applied)", migration.Version, migration.Name)
				continue
			} else {
				log.Fatalf("  [ERROR] Migration %04d_%s was already applied with a different checksum", migration.Version, migration.Name)
			}
		}

		log.Printf("  [RUN]  %04d_%s", migration.Version, migration.Name)

		// Execute migration
		if err := executeMigration(ctx, client, migration); err != nil {
			log.Fatalf("Failed to execute migration %04d_%s: %v", migration.Version, migration.Name, err)
		}

		// Record migration in schema_migrations
		if err := recordMigration(ctx, client, migration); err != nil {
			log.Fatalf("Failed to record migration %04d_%s: %v", migration.Version, migration.Name, err)
		}

		log.Printf("  [OK]   %04d_%s", migration.Version, migration.Name)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("No new migrations to apply. Database is up to date.")
	} else {
		log.Printf("Successfully applied %d migration(s)", appliedCount)
	}
}

// ensureSchemaMigrationsTable creates the schema_migrations table if it doesn't exist
func ensureSchemaMigrationsTable(ctx context.Context, client *bigquery.Client) error {
	return executeSQL(
		ctx,
		client,
		ensureSchemaMigrationsTableQuery,
		[]bigquery.QueryParameter{},
	)
}

// readMigrations reads all migration files from the migrations directory
func readMigrations() ([]Migration, error) {
	files, err := migrationFS.ReadDir("bigquery")
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	// Pattern to match migration files: 0001_name.sql
	pattern := regexp.MustCompile(`^(\d{4})_(.+)\.sql$`)

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := pattern.FindStringSubmatch(file.Name())
		if matches == nil {
			log.Printf("Skipping file with invalid format: %s", file.Name())
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("Skipping file with invalid version: %s", file.Name())
			continue
		}

		name := matches[2]

		// Read SQL content
		content, err := migrationFS.ReadFile("bigquery/" + file.Name())
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", file.Name(), err)
		}

		sql := string(content)
		checksum := fmt.Sprintf("%x", sha256.Sum256(content))

		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			Filename: file.Name(),
			SQL:      sql,
			Checksum: checksum,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations retrieves the list of already applied migrations
func getAppliedMigrations(ctx context.Context, client *bigquery.Client) ([]AppliedMigration, error) {
	query := client.Query(appliedMigrationsQuery)
	it, err := query.Read(ctx)
	if err != nil {
		// If table doesn't exist yet, return empty list
		if strings.Contains(err.Error(), "Not found") {
			return []AppliedMigration{}, nil
		}
		return nil, fmt.Errorf("reading applied migrations: %w", err)
	}

	var applied []AppliedMigration
	for {
		var row struct {
			Version   int64
			Name      string
			AppliedAt time.Time
			Checksum  bigquery.NullString
			AppliedBy bigquery.NullString
		}

		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterating results: %w", err)
		}

		am := AppliedMigration{
			Version:   int(row.Version),
			Name:      row.Name,
			AppliedAt: row.AppliedAt,
		}

		if row.Checksum.Valid {
			am.Checksum = row.Checksum.StringVal
		}

		if row.AppliedBy.Valid {
			am.AppliedBy = row.AppliedBy.StringVal
		}

		applied = append(applied, am)
	}

	return applied, nil
}

// executeMigration executes a single migration SQL
func executeMigration(ctx context.Context, client *bigquery.Client, migration Migration) error {
	return executeSQL(
		ctx,
		client,
		migration.SQL,
		[]bigquery.QueryParameter{},
	)
}

// recordMigration records a successfully applied migration in schema_migrations
func recordMigration(ctx context.Context, client *bigquery.Client, migration Migration) error {
	return executeSQL(
		ctx,
		client,
		recordMigrationQuery,
		[]bigquery.QueryParameter{
			{Name: "version", Value: migration.Version},
			{Name: "name", Value: migration.Name},
			{Name: "checksum", Value: migration.Checksum},
			{Name: "applied_by", Value: *appliedBy},
		},
	)
}

func executeSQL(ctx context.Context, client *bigquery.Client, sql string, parameters []bigquery.QueryParameter) error {
	query := client.Query(sql)
	query.Parameters = parameters

	job, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("running query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("waiting for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job error: %w", err)
	}

	return nil
}

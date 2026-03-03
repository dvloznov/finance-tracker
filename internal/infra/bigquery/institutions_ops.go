package bigquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

const institutionsTable = "institutions"

// ListAllInstitutions retrieves all institutions from the database.
func ListAllInstitutions(ctx context.Context) ([]*bq.InstitutionRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListAllInstitutions: creating client: %w", err)
	}
	defer client.Close()

	return ListAllInstitutionsWithClient(ctx, client)
}

// ListAllInstitutionsWithClient retrieves all institutions using the provided BigQuery client.
func ListAllInstitutionsWithClient(ctx context.Context, client *bigquery.Client) ([]*bq.InstitutionRow, error) {
	query := fmt.Sprintf(`
		SELECT
			institution_id,
			name,
			created_ts,
			updated_ts
	FROM `+"`%s.%s.%s`"+`
	ORDER BY created_ts DESC
	`, projectID, datasetID, institutionsTable)

	q := client.Query(query)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListAllInstitutionsWithClient: reading query: %w", err)
	}

	var institutions []*bq.InstitutionRow
	for {
		var row bq.InstitutionRow
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListAllInstitutionsWithClient: iterating: %w", err)
		}
		institutions = append(institutions, &row)
	}

	return institutions, nil
}

// FindInstitutionByName finds an institution by normalized name.
func FindInstitutionByName(ctx context.Context, name string) (*bq.InstitutionRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("FindInstitutionByName: creating client: %w", err)
	}
	defer client.Close()

	return FindInstitutionByNameWithClient(ctx, client, name)
}

// FindInstitutionByNameWithClient finds an institution using the provided BigQuery client.
func FindInstitutionByNameWithClient(ctx context.Context, client *bigquery.Client, name string) (*bq.InstitutionRow, error) {
	normName := strings.ToUpper(strings.TrimSpace(name))
	if normName == "" {
		return nil, fmt.Errorf("FindInstitutionByNameWithClient: name cannot be empty")
	}

	query := fmt.Sprintf(`
		SELECT
			institution_id,
			name,
			created_ts,
			updated_ts
		FROM `+"`%s.%s.%s`"+`
		WHERE UPPER(TRIM(name)) = @name
		ORDER BY created_ts DESC
		LIMIT 1
	`, projectID, datasetID, institutionsTable)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "name", Value: normName},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindInstitutionByNameWithClient: reading query: %w", err)
	}

	var row bq.InstitutionRow
	err = it.Next(&row)
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindInstitutionByNameWithClient: iterating: %w", err)
	}

	return &row, nil
}

// GetInstitutionByID retrieves an institution by institution_id.
func GetInstitutionByID(ctx context.Context, institutionID string) (*bq.InstitutionRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("GetInstitutionByID: creating client: %w", err)
	}
	defer client.Close()

	return GetInstitutionByIDWithClient(ctx, client, institutionID)
}

// GetInstitutionByIDWithClient retrieves an institution using the provided BigQuery client.
func GetInstitutionByIDWithClient(ctx context.Context, client *bigquery.Client, institutionID string) (*bq.InstitutionRow, error) {
	if institutionID == "" {
		return nil, fmt.Errorf("GetInstitutionByIDWithClient: institution_id cannot be empty")
	}

	query := fmt.Sprintf(`
		SELECT
			institution_id,
			name,
			created_ts,
			updated_ts
		FROM `+"`%s.%s.%s`"+`
		WHERE institution_id = @institution_id
		LIMIT 1
	`, projectID, datasetID, institutionsTable)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: institutionID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetInstitutionByIDWithClient: reading query: %w", err)
	}

	var row bq.InstitutionRow
	err = it.Next(&row)
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetInstitutionByIDWithClient: iterating: %w", err)
	}

	return &row, nil
}

// CreateInstitution inserts a new institution and returns its ID.
func CreateInstitution(ctx context.Context, row *bq.InstitutionRow) (string, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("CreateInstitution: creating client: %w", err)
	}
	defer client.Close()

	return CreateInstitutionWithClient(ctx, client, row)
}

// CreateInstitutionWithClient inserts a new institution using the provided BigQuery client.
func CreateInstitutionWithClient(ctx context.Context, client *bigquery.Client, row *bq.InstitutionRow) (string, error) {
	if row.InstitutionID == "" {
		row.InstitutionID = uuid.NewString()
	}
	if row.CreatedTS.IsZero() {
		row.CreatedTS = time.Now()
	}
	row.UpdatedTS = bigquery.NullTimestamp{Timestamp: time.Now(), Valid: true}

	q := client.Query(`
		INSERT INTO ` + "`" + projectID + "." + datasetID + "." + institutionsTable + "`" + ` (
			institution_id, name, created_ts, updated_ts
		)
		VALUES (
			@institution_id, @name, @created_ts, @updated_ts
		)
	`)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: row.InstitutionID},
		{Name: "name", Value: row.Name},
		{Name: "created_ts", Value: row.CreatedTS},
		{Name: "updated_ts", Value: row.UpdatedTS},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("CreateInstitutionWithClient: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("CreateInstitutionWithClient: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", fmt.Errorf("CreateInstitutionWithClient: job error: %w", err)
	}

	return row.InstitutionID, nil
}

// UpdateInstitution updates an institution's name by institution_id.
func UpdateInstitution(ctx context.Context, institutionID, name string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("UpdateInstitution: creating client: %w", err)
	}
	defer client.Close()

	return UpdateInstitutionWithClient(ctx, client, institutionID, name)
}

// UpdateInstitutionWithClient updates an institution using the provided BigQuery client.
func UpdateInstitutionWithClient(ctx context.Context, client *bigquery.Client, institutionID, name string) error {
	if institutionID == "" {
		return fmt.Errorf("UpdateInstitutionWithClient: institution_id cannot be empty")
	}

	q := client.Query(`
		UPDATE ` + "`" + projectID + "." + datasetID + "." + institutionsTable + "`" + `
		SET name = @name, updated_ts = CURRENT_TIMESTAMP()
		WHERE institution_id = @institution_id
	`)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: institutionID},
		{Name: "name", Value: name},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("UpdateInstitutionWithClient: running update query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("UpdateInstitutionWithClient: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("UpdateInstitutionWithClient: job error: %w", err)
	}

	return nil
}

// UpsertInstitution finds an existing institution by name or creates a new one.
func UpsertInstitution(ctx context.Context, row *bq.InstitutionRow) (string, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("UpsertInstitution: creating client: %w", err)
	}
	defer client.Close()

	return UpsertInstitutionWithClient(ctx, client, row)
}

// UpsertInstitutionWithClient finds or creates an institution using the provided BigQuery client.
func UpsertInstitutionWithClient(ctx context.Context, client *bigquery.Client, row *bq.InstitutionRow) (string, error) {
	if row.Name != "" {
		existing, err := FindInstitutionByNameWithClient(ctx, client, row.Name)
		if err != nil {
			return "", fmt.Errorf("UpsertInstitutionWithClient: finding existing institution: %w", err)
		}
		if existing != nil {
			return existing.InstitutionID, nil
		}
	}

	if row.InstitutionID == "" {
		row.InstitutionID = uuid.NewString()
	}
	if row.CreatedTS.IsZero() {
		row.CreatedTS = time.Now()
	}
	row.UpdatedTS = bigquery.NullTimestamp{Timestamp: time.Now(), Valid: true}

	q := client.Query(`
		INSERT INTO ` + "`" + projectID + "." + datasetID + "." + institutionsTable + "`" + ` (
			institution_id, name, created_ts, updated_ts
		)
		VALUES (
			@institution_id, @name, @created_ts, @updated_ts
		)
	`)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: row.InstitutionID},
		{Name: "name", Value: row.Name},
		{Name: "created_ts", Value: row.CreatedTS},
		{Name: "updated_ts", Value: row.UpdatedTS},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("UpsertInstitutionWithClient: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("UpsertInstitutionWithClient: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", fmt.Errorf("UpsertInstitutionWithClient: job error: %w", err)
	}

	return row.InstitutionID, nil
}

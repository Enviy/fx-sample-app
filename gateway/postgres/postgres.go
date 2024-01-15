package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/config"
)

// Gateway defines methods for interacting with postgres.
type Gateway interface {
	SelectPlaceByFilter(f Place) ([]Place, error)
	InitializeDB() error
	PopulateDB() error
}

// gateway defines implementation of Gateway interface.
type gateway struct {
	db     *sqlx.DB
	txOpts *sql.TxOptions
}

// New is the Gateway interface constructor.
func New(cfg config.Provider) (Gateway, error) {
	connStr := fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=%s",
		cfg.Get("postgres.user").String(),
		cfg.Get("postgres.db_name").String(),
		cfg.Get("postgres.password").String(),
		cfg.Get("postgres.ssl_mode").String(),
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("sql Open %w", err)
	}

	return &gateway{
		db:     db,
		txOpts: &sql.TxOptions{Isolation: sql.LevelSerializable},
	}, nil
}

// selectByFilter is a sample method for variable filter queries.
func (g *gateway) SelectPlaceByFilter(f Place) ([]Place, error) {
	// Convert struct to map string interface.
	queryMap, err := toMap(f)
	if err != nil {
		return nil, fmt.Errorf("toMap %w", err)
	}

	// Build the query.
	query := "SELECT * FROM place"
	var index int
	var prefix string
	for key := range queryMap {
		if index == 0 {
			prefix = "WHERE"
		}
		query += fmt.Sprintf(" %s %s=:%s", prefix, key, key)
		index += 1
	}

	// Begin transaction.
	ctx := context.Background()
	tx, err := g.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("BeginTxx %w", err)
	}
	defer tx.Rollback()

	// Prepare query as built statement.
	var places []Place
	preparedStmnt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("PrepareNamedContext %w", err)
	}

	// Execute query with params.
	err = preparedStmnt.SelectContext(ctx, &places, queryMap)
	if err != nil {
		return nil, fmt.Errorf("SelectContext %w", err)
	}

	// End transaction.
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("Commit %w", err)
	}

	return places, nil
}

// InitializeDB sets up DB tables.
func (g *gateway) InitializeDB() error {
	// Setup the table using the defined schema.
	_, err := g.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("Exec %w", err)
	}

	return nil
}

// PopulateDB sets up test data in tables.
func (g *gateway) PopulateDB() error {
	// Open a new transaction to update table with data.
	ctx := context.Background()
	tx, err := g.db.BeginTxx(ctx, g.txOpts)
	if err != nil {
		return fmt.Errorf("schema transaction begin %w", err)
	}
	defer tx.Rollback()

	// Begin insertions to populate table with rows.
	// PrepareNamedContext(ctx context.Context, query string) (*NamedStmt, error)
	// ExecContext(ctx context.Context, arg interface{}) (sql.Result, error)
	// https://github.com/jmoiron/sqlx/blob/v1.3.5/sqlx_context.go#L315
	personQuery, err := tx.PrepareNamedContext(
		ctx,
		"INSERT INTO person (id, first_name, last_name, email, timestamp) VALUES (:id, :first_name, :last_name, :email, :timestamp)",
	)
	if err != nil {
		return fmt.Errorf("PrepareNamedContext %w", err)
	}

	_, err = personQuery.ExecContext(
		ctx,
		&Person{
			ID:        uuid.New().String(),
			FirstName: "Jason",
			LastName:  "Moiron",
			Email:     "jmoiron@jmoiron.net",
			Timestamp: time.Now().UTC().Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("first insert exec %w", err)
	}

	_, err = personQuery.ExecContext(
		ctx,
		&Person{
			ID:        uuid.New().String(),
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoeDNE@gmail.net",
			Timestamp: time.Now().UTC().Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("second insert exec %w", err)
	}

	placeQuery, err := tx.PrepareNamedContext(
		ctx,
		"INSERT INTO place (id, country, city, telcode, comments, timestamp) VALUES (:id, :country, :city, :telcode, :comments, :timestamp)",
	)
	_, err = placeQuery.ExecContext(
		ctx,
		&Place{
			ID:      uuid.New().String(),
			Country: "United States",
			City: sql.NullString{
				String: "New York",
				Valid:  true,
			},
			TelCode:   1,
			Comments:  []string{},
			Timestamp: time.Now().UTC().Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("third insert exec %w", err)
	}

	_, err = placeQuery.ExecContext(
		ctx,
		&Place{
			ID:        uuid.New().String(),
			Country:   "Hong Kong",
			TelCode:   852,
			Comments:  []string{`{"key_one": "value_one"}`},
			Timestamp: time.Now().UTC().Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("fourth insert exec %w", err)
	}

	_, err = placeQuery.ExecContext(
		ctx,
		&Place{
			ID:        uuid.New().String(),
			Country:   "Singapore",
			TelCode:   65,
			Comments:  []string{`{"key1": "value1"}`},
			Timestamp: time.Now().UTC().Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("fifth insert exec %w", err)
	}

	// Example of inserting a row using a populated struct.
	_, err = personQuery.ExecContext(
		ctx,
		&Person{
			ID:        uuid.New().String(),
			FirstName: "Jane",
			LastName:  "Citizen",
			Email:     "jane.citzen@example.com",
			Timestamp: time.Now().UTC().Unix(),
		},
	)
	if err != nil {
		return fmt.Errorf("sixth insert exec %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction %w", err)
	}

	return nil
}

// toMap parses a struct to a map accounting for sql.Nullx types.
// Supports using a single struct for reading and writing rows.
func toMap(p interface{}) (map[string]interface{}, error) {
	queryBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	var queryMap map[string]interface{}
	err = json.Unmarshal(queryBytes, &queryMap)
	if err != nil {
		return nil, err
	}

	// Resolve map values to underlying values.
	// ONLY handles sql.Nullx type values.
	nullables := []string{
		"String",
		"Bool",
		"Int64",
		"Byte",
		"Float64",
		"Int16",
		"Int32",
		"Time",
	}
	// Parse resulting maps of sql.Nullx types. Remove invalids.
	for key, value := range queryMap {
		if asMap, ok := value.(map[string]interface{}); ok {
			fmt.Printf("key: %s - as map: %v\n", key, asMap)
			if !asMap["Valid"].(bool) {
				delete(queryMap, key)
				continue
			}

			for _, nullable := range nullables {
				if subVal, ok := asMap[nullable]; ok {
					queryMap[key] = subVal
				}
			}
		}
	}

	return queryMap, nil
}

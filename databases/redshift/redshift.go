package redshift

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/yindia/xray/config"
	"github.com/yindia/xray/types"
)

// DB_PASSWORD is the environment variable that holds the database password.
var DB_PASSWORD = "DB_PASSWORD"

// Redshift_Schema_query is the SQL query used to describe a table schema in Redshift.
// Redshift_Tables_query is the SQL query used to list all tables in a schema in Redshift.
const (
	Redshift_Schema_query = `SELECT "column", type, encoding, distkey, sortkey, "notnull"  FROM pg_table_def WHERE schemaname = '%s' AND tablename = '%s';`
	Redshift_Tables_query = "SHOW TABLES FROM SCHEMA %s.public;"
)

// Redshift is a Redshift implementation of the ISQL interface.
type Redshift struct {
	Client *sql.DB
	Config config.Config
}

// NewRedshift creates a new Redshift client with the given sql.DB.
func NewRedshift(client *sql.DB) (types.ISQL, error) {
	return &Redshift{
		Client: client,
		Config: config.Config{},
	}, nil
}

// NewRedshiftWithConfig creates a new Redshift client with the given configuration.
// It returns an error if the DB_PASSWORD environment variable is not set.
// It uses the postgres driver to connect to the database.
func NewRedshiftWithConfig(cfg *config.Config) (types.ISQL, error) {
	if os.Getenv(DB_PASSWORD) == "" || len(os.Getenv(DB_PASSWORD)) == 0 {
		return nil, fmt.Errorf("please set %s env variable for the database", DB_PASSWORD)
	}
	DB_PASSWORD = os.Getenv(DB_PASSWORD)

	dsn := fmt.Sprintf("host=%s port=%v user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.Username, DB_PASSWORD, cfg.Database, cfg.SSL)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error creating a new session : %v", err)
	}

	return &Redshift{
		Client: db,
		Config: *cfg,
	}, nil
}

// Schema returns the schema of a table in Redshift.
// It takes the table name as an argument and returns a Table struct and an error.
func (r *Redshift) Schema(table string) (types.Table, error) {
	if len(r.Config.Schema) == 0 {
		r.Config.Schema = "public"
	}

	query := fmt.Sprintf(Redshift_Schema_query, r.Config.Schema, table)
	ctx := context.Background()
	rows, err := r.Client.QueryContext(ctx, query)
	if err != nil {
		return types.Table{}, fmt.Errorf("error executing query: %v", err)
	}

	var columns []types.Column
	for rows.Next() {
		var column types.Column
		var encoding string
		var distkey bool
		var sortkey int
		var notnull bool
		if err := rows.Scan(
			&column.Name,
			&column.Type,
			&encoding,
			&distkey,
			&sortkey,
			&notnull,
		); err != nil {
			return types.Table{}, fmt.Errorf("error scanning rows: %v", err)
		}
		column.Metatags = []string{encoding, fmt.Sprintf("distkey:%v", distkey), fmt.Sprintf("sortkey:%d", sortkey), fmt.Sprintf("notnull:%v", notnull)}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return types.Table{}, fmt.Errorf("error iterating over rows: %v", err)
	}

	return types.Table{
		Name:        table,
		Columns:     columns,
		ColumnCount: int64(len(columns)),
		Description: "",
		Metatags:    []string{},
	}, nil
}

func (r *Redshift) Tables(databaseName string) ([]string, error) {
	// ctx := context.Background()
	query := fmt.Sprintf(Redshift_Tables_query, databaseName)

	res, err := r.Client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}

	var tables []string

	for res.Next() {
		var table types.TableResponse
		if err := res.Scan(&table.Database, &table.SchemaName, &table.TableName, &table.TableType, &table.TableAcl, &table.Remarks); err != nil {
			return nil, fmt.Errorf("error scanning result: %v", err)
		}
		fmt.Println(table)
		tables = append(tables, table.TableName)
	}
	fmt.Println(tables)

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over result: %v", err)
	}

	return tables, nil

}

// Execute executes a query on Redshift.
// It takes a query string as input and returns the result as a byte slice and an error.
func (r *Redshift) Execute(query string) ([]byte, error) {
	ctx := context.Background()
	rows, err := r.Client.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}

	// getting the column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	// Scan the result into a slice of slices
	var results [][]interface{}
	for rows.Next() {
		// create a slice of values and pointers
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			//  create a slice of pointers to the values
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Decode base64 data
		stringRow := make([]interface{}, len(values))
		for i, val := range values {
			switch v := val.(type) {
			case []byte:
				strVal := string(v)
				if isBase64(strVal) {
					decoded, err := base64.StdEncoding.DecodeString(strVal)
					if err != nil {
						return nil, fmt.Errorf("error decoding base64 data: %v", err)
					}
					stringRow[i] = string(decoded)
				} else {
					stringRow[i] = strVal
				}
			case string:
				stringRow[i] = v
			case nil:
				stringRow[i] = nil
			default:
				stringRow[i] = fmt.Sprintf("%v", v)
			}
		}
		results = append(results, stringRow)

	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Convert the result to JSON
	queryResult := types.QueryResult{
		Columns: columns,
		Rows:    results,
	}
	jsonData, err := json.Marshal(queryResult)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}

	return jsonData, nil
}

func isBase64(s string) bool {
	if len(s)%4 != 0 {
		return false
	}
	// Try to decode the string
    _, err := base64.StdEncoding.DecodeString(s)
    // If decoding succeeds, err will be nil, and the function will return true
    // If decoding fails, err will not be nil, and the function will return false
	// Also we do not have access to decoded value, so we are not using it
	return err == nil
}

// GenerateCreateTableQuery generates a CREATE TABLE query for Redshift.
// It takes a Table struct as an argument and returns a string.
func (r *Redshift) GenerateCreateTableQuery(table types.Table) string {
	query := fmt.Sprintf("CREATE TABLE %s.%s.%s (", r.Config.Database, r.Config.Schema, table.Name)
	for i, column := range table.Columns {
		colType := strings.ToUpper(column.Type)
		query += column.Name + " " + convertTypeToRedshift(colType)

		if column.IsPrimary {
			query += " PRIMARY KEY"
			if column.AutoIncrement {
				query += fmt.Sprintf(" IDENTITY(%v, %v)", column.IdentitySeed, column.IdentityStep)
			}
		}

		if column.IsNullable == "NO" {
			query += " NOT NULL"
		}

		if i < len(table.Columns)-1 {
			query += ", "
		}
	}
	query += ");"
	return query
}

// convertTypeToRedshift converts a given column type to its equivalent in Redshift.
func convertTypeToRedshift(dataType string) string {
	// Map column types to Redshift equivalents
	switch dataType {
	case "SMALLINT", "INT2":
		return "SMALLINT"
	case "INTEGER", "INT", "INT4":
		return "INTEGER"
	case "BIGINT", "INT8":
		return "BIGINT"
	case "DECIMAL", "NUMERIC":
		return "DECIMAL"
	case "REAL", "FLOAT4":
		return "REAL"
	case "DOUBLE PRECISION", "FLOAT8", "FLOAT":
		return "DOUBLE PRECISION"
	case "CHAR", "CHARACTER", "NCHAR", "BPCHAR":
		return "CHAR"
	case "VARCHAR", "CHARACTER VARYING", "NVARCHAR", "TEXT":
		return "VARCHAR"
	case "DATE":
		return "DATE"
	case "TIME", "TIME WITHOUT TIME ZONE":
		return "TIME"
	case "TIMETZ", "TIME WITH TIME ZONE":
		return "TIMETZ"
	case "TIMESTAMP", "TIMESTAMP WITHOUT TIME ZONE":
		return "TIMESTAMP"
	case "TIMESTAMPTZ", "TIMESTAMP WITH TIME ZONE":
		return "TIMESTAMPTZ"
	case "INTERVAL YEAR TO MONTH":
		return "INTERVAL YEAR TO MONTH"
	case "INTERVAL DAY TO SECOND":
		return "INTERVAL DAY TO SECOND"
	case "BOOLEAN", "BOOL":
		return "BOOLEAN"
	case "HLLSKETCH":
		return "HLLSKETCH"
	case "SUPER":
		return "SUPER"
	case "VARBYTE", "VARBINARY", "BINARY VARYING":
		return "VARBYTE"
	case "GEOMETRY":
		return "GEOMETRY"
	case "GEOGRAPHY":
		return "GEOGRAPHY"
	// Add more type conversions as needed
	default:
		return dataType
	}
}

package bigquery

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/yindia/xray/config"
	"github.com/yindia/xray/types"
	_ "gorm.io/driver/bigquery/driver"
)

var GOOGLE_APPLICATION_CREDENTIALS = "GOOGLE_APPLICATION_CREDENTIALS"

const (
	BigQuery_SCHEMA_QUERY = "SELECT column_name, data_type FROM %s.INFORMATION_SCHEMA.COLUMNS WHERE table_name='%s'"
	BigQuery_TABLES_QUERY = "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = '%s'"
)

// The BigQuery struct is responsible for holding the BigQuery client and configuration.
type BigQuery struct {
	Client *sql.DB
	Config *config.Config
}

// NewBigQuery creates a new instance of BigQuery with the provided client.
// It returns an instance of types.ISQL and an error.
func NewBigQuery(client *sql.DB) (types.ISQL, error) {
	return &BigQuery{
		Client: client,
		Config: &config.Config{},
	}, nil
}

// NewBigQueryWithConfig creates a new instance of BigQuery with the provided configuration.
// It returns an instance of types.ISQL and an error.
func NewBigQueryWithConfig(cfg *config.Config) (types.ISQL, error) {
	if os.Getenv(GOOGLE_APPLICATION_CREDENTIALS) == "" || len(os.Getenv(GOOGLE_APPLICATION_CREDENTIALS)) == 0 {
		return nil, fmt.Errorf("please set %s env variable for the database", GOOGLE_APPLICATION_CREDENTIALS)
	}
	GOOGLE_APPLICATION_CREDENTIALS = os.Getenv(GOOGLE_APPLICATION_CREDENTIALS)

	dbType := types.BigQuery
	connectionString := fmt.Sprintf("bigquery://%s/%s", cfg.ProjectID, cfg.Database)
	db, err := sql.Open(dbType.String(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("database connecetion failed : %v", err)
	}

	return &BigQuery{
		Client: db,
		Config: cfg,
	}, nil
}

// this function extarcts the schema of a table in BigQuery.
// It takes table name as input and returns a Table struct and an error.
func (b *BigQuery) Schema(table string) (types.Table, error) {
	// execute the sql statement
	rows, err := b.Client.Query(fmt.Sprintf(BigQuery_SCHEMA_QUERY, b.Config.Database, table))
	if err != nil {
		return types.Table{}, fmt.Errorf("error executing sql statement: %v", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var columns []types.Column
	for rows.Next() {
		var column types.Column
		if err := rows.Scan(&column.Name, &column.Type); err != nil {
			return types.Table{}, fmt.Errorf("error scanning row: %v", err)
		}
		columns = append(columns, column)
	}

	return types.Table{
		Name:        table,
		Columns:     columns,
		Dataset:     b.Config.Database,
		ColumnCount: int64(len(columns)),
	}, nil
}

// Execute executes a query on BigQuery.
// It takes a query string as input and returns the result as a byte slice and an error.
func (b *BigQuery) Execute(query string) ([]byte, error) {
	rows, err := b.Client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error executing sql statement: %v", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	// Scan the result into a slice of slices
	var results []map[string]interface{}
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

		// Create a map for the current row
        rowMap := make(map[string]interface{})
        for i, colName := range columns {
            // If the value is of type []byte (which is used for RECORD data types), 
            // we attempt to unmarshal it into a map[string]interface{}
            if b, ok := values[i].([]byte); ok {
                var m map[string]interface{}
                if err := json.Unmarshal(b, &m); err == nil {
                    rowMap[colName] = m
                } else {
                    rowMap[colName] = values[i]
                }
            } else {
                rowMap[colName] = values[i]
            }
        }

		results = append(results, rowMap)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Convert the result to JSON
	queryResult := types.BigQueryResult{
		Columns: columns,
		Rows:    results,
	}

	jsonData, err := json.Marshal(queryResult)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}

	return jsonData, nil

}

// Tables returns a list of tables in a dataset.
// It takes a dataset name as input and returns a slice of strings and an error.
func (b *BigQuery) Tables(dataset string) ([]string, error) {
	// res, err := b.Client.Query("SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = '" + Dataset + "'")

	rows, err := b.Client.Query(fmt.Sprintf(BigQuery_TABLES_QUERY, dataset))
	if err != nil {
		return nil, fmt.Errorf("error executing sql statement: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	var tables []string

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("error scanning dataset")
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error interating over rows: %v", err)
	}

	return tables, nil
}

// GenerateCreateTableQuery generates a CREATE TABLE query for BigQuery.
func (b *BigQuery) GenerateCreateTableQuery(table types.Table) string {
	query := "CREATE TABLE " + table.Dataset + "." + table.Name + " ("
	for i, column := range table.Columns {
		colType := strings.ToUpper(column.Type)
		query += column.Name + " " + convertTypeToBigQuery(colType)

		if i < len(table.Columns)-1 {
			query += ", "
		}
	}
	query += ");"
	return query
}

// convertTypeToBigQuery converts a Data type to a BigQuery SQL Data type.
func convertTypeToBigQuery(dataType string) string {
	// Map column types to BigQuery equivalents
	switch dataType {
	case "ARRAY":
		return "ARRAY"
	case "BIGNUMERIC":
		return "BIGNUMERIC"
	case "BOOL":
		return "BOOL"
	case "BYTES":
		return "BYTES"
	case "DATE":
		return "DATE"
	case "DATETIME":
		return "DATETIME"
	case "FLOAT64", "FLOAT":
		return "FLOAT64"
	case "GEOGRAPHY":
		return "GEOGRAPHY"
	case "INT64", "INT", "INTEGER":
		return "INT64"
	case "INTERVAL":
		return "INTERVAL"
	case "JSON":
		return "JSON"
	case "NUMERIC":
		return "NUMERIC"
	case "RANGE":
		return "RANGE"
	case "STRING", "VARCHAR(255)", "TEXT":
		return "STRING"
	case "STRUCT":
		return "STRUCT"
	case "TIME":
		return "TIME"
	case "TIMESTAMP":
		return "TIMESTAMP"
	// Add more type conversions as needed
	default:
		return dataType
	}
}

package main

import (
	"fmt"

	"github.com/yindia/xray"
	"github.com/yindia/xray/config"
	"github.com/yindia/xray/types"
)

// export GOOGLE_APPLICATION_CREDENTIALS=path/to/secret.json
func main() {
	config := &config.Config{
		ProjectID: "textquery-379122",
		Database:  "bigquerytrends",
	}

	client, err := xray.NewClientWithConfig(config, types.BigQuery)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")

	tables, err := client.Tables(config.Database)
	if err != nil {
		panic(err)
	}

	fmt.Println("Tables :", tables)

	var response []types.Table
	for _, v := range tables {
		table, err := client.Schema(v)
		if err != nil {
			panic(err)
		}
		response = append(response, table)
	}
	fmt.Println(response)

	for _, v := range response {
		query := client.GenerateCreateTableQuery(v)
		fmt.Println(query)
	}
}

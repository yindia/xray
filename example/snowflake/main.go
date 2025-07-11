package main

import (
	"fmt"

	_ "github.com/snowflakedb/gosnowflake"
	"github.com/yindia/xray"
	"github.com/yindia/xray/config"
	"github.com/yindia/xray/types"
)

// export DB Passowrd, Export root=DB_PASSWORD
func main() {
	config := &config.Config{
		Account:   "tvhcdje-pd56667",
		Username:  "jaizadarsh",
		Database:  "SNOWFLAKE_SAMPLE_DATA",
		Port:      "443",
		Warehouse: "COMPUTE_WH",
		Schema:    "TPCH_SF10", // optional
	}

	client, err := xray.NewClientWithConfig(config, types.Snowflake)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")

	data, err := client.Tables(config.Database)
	if err != nil {
		panic(err)
	}
	fmt.Println("Tables :", data)

	var response []types.Table
	for _, v := range data {
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

package main

import (
	"fmt"

	"github.com/yindia/xray"
	"github.com/yindia/xray/config"
	"github.com/yindia/xray/types"
)

// export DB_PASSWORD=your_password
func main() {
	config := &config.Config{
		Host:     "127.0.0.1",
		Database: "employees",
		Username: "root",
		Port:     "3306",
		SSL:      "false",
	}
	client, err := xray.NewClientWithConfig(config, types.MySQL)
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

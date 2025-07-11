## Using Xray with MySQL Integration

This guide demonstrates how to use the Xray library to inspect a MySQL database in a Go application.

### Introduction to Xray

Xray is a library that provides tools for inspecting and analyzing various types of databases. In this example, we'll use Xray to connect to a MySQL database, retrieve the schema for each table, and generate SQL queries.

### Step-by-Step Guide

1. **Define Database Configuration**

  
   Add your MySQL Database configuration to your Go application:

    **Note : Set env variable DB_Password for adding password and pass your password as : `export DB_PASSWORD=your_password`**

   ```go
   import (
    "github.com/yindia/xray/config"
    "github.com/yindia/xray/types"
    "github.com/yindia/xray"
   )
   // Define your database configuration here
   dbConfig := &config.Config{
       Host:         "127.0.0.1",
       Database: "employees",
       Username:     "root",
       Port:         "3306",
       SSL:          "false",
   }

**Note** : Ensure to replace the placeholders with your actual database configuration.

1. **Connect to MySQL Database**
   
    ```go
    client, err := xray.NewClientWithConfig(dbConfig, types.MySQL)
    if err != nil {
        panic(err)
    }
2. **Retrieve Table Schema**

    ```go
    data, err := client.Tables(dbConfig.Database)
    if err != nil {
        panic(err)
    }
3. **Print Table Schema**

    ```go
    for _, tableName := range data {
        table, err := client.Schema(tableName)
        if err != nil {
            panic(err)
        }
        fmt.Println(table)
    }

4. **Generate Create Table Query**
   
   generate and print SQL CREATE TABLE queries for each table in the response slice.

   ```go
    // Iterate over each table in the response slice.
    for _, v := range response {
        // Generate a CREATE TABLE query for the current table.
        query := client.GenerateCreateTableQuery(v)
        // Print the generated query.
        fmt.Println(query)
    }
    ```
5. **Execute Queries**

    Execute queries against your  database:

    ```go
    query := "SELECT * FROM my_table" // Specify your SQL query
    result, err := client.Execute(query)
    if err != nil {
        fmt.Printf("Error executing query: %v\n", err)
        return
    }
    fmt.Printf("Query result: %s\n", string(result))


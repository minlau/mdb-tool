# MDB Tool

[![CI](https://github.com/minlau/mdb-tool/actions/workflows/ci.yml/badge.svg)](https://github.com/minlau/mdb-tool/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/minlau/mdb-tool)](https://goreportcard.com/report/github.com/minlau/mdb-tool)

**M**ultiple **D**ata**B**ase **Tool** simplifies querying data from multiple databases _at once_.

Usage case: you have multiple environments(separate for every/some clients) and separate database for each microservice,
and you want to execute same query in specific microservice database of each environment. Group Name is environment and
Group Type is microservice database.

## Preview

https://user-images.githubusercontent.com/5850570/217761596-b54b9fed-eb48-4af4-9e44-f479e65778c6.mp4

## Features

- Lightweight and fast web UI
- Query multiple databases by **group type** and view data in single table
- Query a single database by **group type** and **group name**
- Supports postgresql, mysql and firebird databases

## How to build

Docker is the easiest way to build and run. `docker-compose up` command can be used to build and run project with 
databases used for testing. To build and run locally you need(might work with lower versions):

- Go 1.19
- Node 19
- Yarn 3

#### Development

Build js app(UI) with live reload. Built UI files are saved to `static` folder

```
cd web/ui/js-app
yarn install && yarn start 
```

Build go executable without embedding UI files

```
go build ./cmd/mdb-tool -tags=dev
```

#### Production

Build js app(UI). Build UI files are saved to `static` folder

```
cd web/ui/js-app
yarn install && yarn build 
```

Build go executable with embedded UI files

```
go build ./cmd/mdb-tool
```

## How to run

Required files:

- executable
- `static` folder if it is not embedded to executable
- config file. Databases and datasources can be provided only by config file.

Program arguments:

- config - config file path. Default: config.json
- port - port of application. Default: 8080

Command example:

``
mdb-tool --config=config_file_path.json --port=8080
``

### Config

Fields definition:

- groupName - name of databases group(environment, client)
- groupType - group type of database(database name)
- type - type of database. Supported: postgresql, mysql, firebird

Not listed fields are used for connecting to database.  
Keep in mind that **groupName** and **groupType** combination must be **unique**

Example:

```
{
  "dataSources": [
    {
      "query": "select 'groupName' as \"groupName\", 'groupType' as \"groupType\", 'localhost' as hostname, 5432 as port, 'name' as name, 'username' as username, 'password' as password, 'postgresql' as type, 4 as \"maxOpenConns\", 1 as \"maxIdleConns\", 600 as \"connMaxLifetimeInSeconds\", 60 as \"connMaxIdleTimeInSeconds\"",
      
      "hostname": "localhost",
      "port": 5432,
      "name": "db",
      "username": "postgres",
      "password": "admin",
      "type": "postgresql"
    }
  ],
  "databaseConfigs": [
    {
      "groupName": "test-env-1",
      "groupType" : "messaging",
      
      "hostname": "localhost",
      "port": 5432,
      "name": "messaging",
      "username": "postgres",
      "password": "admin",
      "type": "postgresql",
      "maxOpenConns": 4,
      "maxIdleConns": 1,
      "connMaxLifetimeInSeconds": 300,
      "connMaxIdleTimeInSeconds": 60
    }
  ]
}
```

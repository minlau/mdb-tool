# MDB Tool

**M**ultiple **D**ata**B**ase **Tool** simplifies querying data of multiple databases _at once_

## Features
- Web UI
- Query multiple databases by **group type** and view data in single table
- Query a single database by **group type** and **group id**
- Supports postgresql, mysql and firebird databases
 
## How to build

#### Development

Build js app(UI) with live reload. Built UI files are saved to `static` folder
```
cd js-app
yarn install && yarn start 
```

Build go executable without embedding UI files
```
go build -tags=dev
```

#### Production

Build js app(UI). Build UI files are saved to `static` folder
```
cd js-app
yarn install && yarn build 
```

Build go executable with embedded UI files
```
go build
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
- groupId - id of databases group (unique in group)
- groupType - group type of database
- title - title displayed in UI
- type - type of database. Supported: postgresql, mysql, firebird

Not listed fields are used for connecting to database.  
Keep in mind that **groupId** and **groupType** combination must be **unique**

Example:
```
{
  "dataSources": [
    {
      "query": "select 1 as \"groupId\", 'groupType' as \"groupType\", 'title' as title, 'localhost' as hostname, 5432 as port, 'name' as name, 'username' as username, 'password' as password, 'postgresql' as type, 4 as \"maxOpenConns\", 1 as \"maxIdleConns\", 600 as \"connMaxLifetimeInSeconds\", 60 as \"connMaxIdleTimeInSeconds\"",
      
      "hostname": "localhost",
      "port": 5432,
      "name": "db1",
      "username": "postgres",
      "password": "admin",
      "type": "postgresql"
    }
  ],
  "databaseConfigs": [
    {
      "groupId": 1,
      "groupType" : "main",
      "title": "First",
      
      "hostname": "localhost",
      "port": 5432,
      "name": "db1",
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
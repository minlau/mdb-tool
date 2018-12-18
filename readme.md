# MDB Tool

**M**ultiple **D**ata**B**ase **Tool** simplifies querying data of multiple databases _at once_

## Features
- Web UI
- Query all databases by **group type** and view data in single table
- Query single database by **group type** and **group id**
- Supports postgresql, mysql and firebird databases
 
### TODO
- Better naming for group(id, type). It is confusing
- Setting in web UI to split data by group id into separate tables
- Support for non sql databases?
 
### Known issues
- Duplicate query result column names are renamed
 
## Getting Started

How to build js app:

```
cd js-app
yarn install && yarn build 
```

built files are saved to ./assets folder

### Config

Fields definition:
- groupId - id of databases group
- groupType - group type of database (unique in group)
- title - title displayed in UI
- type - type of database. Supported: postgresql, mysql, firebird

Not listed fields are used for connecting to database.  
groupId+groupType is **unique**

Example:
```
{
  "dataSources": [
    {
      "query": "select 1 as groupId, 'groupType' as groupType, 'title' as title, 'localhost' as hostname, 5432 as port, 'name' as name, 'username' as username, 'password' as password, 'postgresql' as type",
      
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
      "type": "postgresql"
    }
  ]
}
```

## Deployment

Executable requires assets folder and a config file

``
mdb --config=config.json --port=8080
``
# OceanBase Todo List

OceanBase Todo List, an extremely simple web application that shows how to use OceanBase as backend database. Initially, it was created for learning about ob-operator referenced in [a real-world example](https://oceanbase.github.io/ob-operator/docs/manual/appendix/example).

## Features

- Connect the database and execute the database migration (Create tables)
- Create initial todo list data in the database (Steps for learning about ob-operator)
- Provide RESTful API for frontend app to interact with the database
- Provide a simple frontend app to show the todo list
  - Show all todo list
  - Add a new todo item
  - Update a todo item (title and description)
  - Done/Undone a todo item
  - Delete a todo item

## How to use

```bash
# Build frontend app first
cd ui
yarn # or npm install
yarn build # or npm run build

# Run backend app
cd ..
go mod tidy
go build -o oceanbase-todo main.go
DB_HOST=xxx DB_PORT=xxx DB_USER=xxx DB_PASSWORD=xxx DB_DATABASE=xxx ./oceanbase-todo
```
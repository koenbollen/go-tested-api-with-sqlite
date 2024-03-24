# go-tested-api-with-sqlite

This is a template for a simple API written in Go, with feature tests and a SQLite database.

### TODO:
- Kubernetes manifests
- Extract cucumber steps definitions to a package

## Use this template

To use this template, click the "Use this template" button at the top of the 
page. Then:
- make sure to search and replace `go-tested-api-with-sqlite` with your 
project name.
- remove the internal/routes/redirections.go file and start your own routes.
- remove the features/redirections/* tests
- clean the initial migrations and create your own.
- update this README.md with your project's details. Still seeing these 
  instructions in a project that is a copy of this template? Embarrassing, 
  right?
- start coding!

<small>(note: the redirections routes and features tests are very minimal, not
great examples)</small>

## Development

Adding functionally to this project could go as follows:
1. Add your desired outcome to the `features/` directory. Write a feature test
   that describes the new feature or bugfix you want to implement.
1. Create a pull request with only the new feature test (still failing) and 
   discuss with the team. Here you can make API design decisions before 
   implementing any code.
1. Fix you bug, add the new route or feature until the test passes.
1. Add more tests for edge cases and non-happy paths.
1. Profit!

To run the feature tests, you can use the following command:

```bash
cd features
# run all feature files:
go test -v .
# or just one feature file:
go test -v . -- api.feature 
```
(you can also run all the tests in the root directory with `go test ./...`)

## Running a service

To run a service, you can use the following commands:

```bash
go run cmd/api/main.go
```

If you project has multiple components, you can can add them in the `cmd/` 
directory and run them the same.

## Database Migrations

Database migrations are stored in the `migrations` directory. Each migration is 
a pair of `.up.sql` and `.down.sql` files. The `.up.sql` file contains the SQL 
to apply the migration, and the `.down.sql` file contains the SQL to revert the 
migration.
Migrations are automatically applied when the service starts.

### Create new migration

To create a new migration file you can use the following script:

```bash
./migration/create-migraion add_users
$EDITOR migrations/*add_users.up.sql # Write the migration
```

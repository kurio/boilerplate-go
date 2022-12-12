# Go App

Go App is an app written in Go.

## Contributing

Checkout and prepare the dev environment:

```bash
git clone https://github.com/kurio/boilerplate-go.git
cd boilerplate-go
make prepare-dev
```

### Database Migrations

To add another migration, run:

```bash
migrate create -dir internal/mysql/migrations -ext sql MIGRATION_NAME
```

Set the `MIGRATION_NAME` to describe the migration.
Please refer to: https://github.com/golang-migrate/migrate/tree/master/cli

### Running

To start and stop the HTTP server and everything, run:

```bash
make mysql-up
make mongo-up
make redis-up
make docker
make migrate-up
make run
make stop
```

### Testing

For running all tests, use `make test`.

For running only unittests, use `make unittest`.

### Committing Changes

Commit checklist:

1. Run linter with `make lint`
2. Run all tests (see: [Testing](#Testing))
3. Commit with proper, descriptive message (see: [https://karma-runner.github.io/2.0/dev/git-commit-msg.html](https://karma-runner.github.io/2.0/dev/git-commit-msg.html))
4. Create *Pull Request*, describing the changes

# Go App

Go App is an app written in Go.

## Contributing

Checkout and prepare the dev environment:

```bash
git clone https://github.com/kurio/boilerplate-go.git
cd boilerplate-go
make prepare-dev
```

If you have a difficulty on getting modules from private repository, this command should help:

```bash
go env -w GOPRIVATE=github.com/KurioApp/*
```

### Database Migrations

To add another migration, run:

```bash
migrate create -dir internal/mysql/migrations -ext sql MIGRATION_NAME
```

Set the `MIGRATION_NAME` to describe the migration.
Please refer to: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

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

To run with opentelemetry, after `migrate-up`, use these instead:

```bash
make otel-up
make run-with-otel
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

## Profiling

Resources regarding profiling:

* https://go.dev/blog/pprof
* https://github.com/google/pprof/issues/166

In this repository, we expose the `/debug` endpoint when `debug` is set to `true`.

http://localhost:7723/debug/pprof/

To look at the flamegraph:

* Start the server
* Hit the server (using `wrk`, `hey`, `ab`, `locust`, etc.)
* Run `go tool pprof http://localhost:7723/debug/pprof/profile`
* After the profile is generated, run `go tool pprof -http=: <path-to-profile>`

Example:

```bash
# set DEBUG to true in .env
make run

# Run locust
cd ../locust
locust --config goboilerplate.conf

# Open up the Web UI after generating 30-second CPU profile
go tool pprof -http=:9999 http://localhost:7723/debug/pprof/profile
# Open up the Web UI after generating heap profile
go tool pprof -http=:9999 http://localhost:7723/debug/pprof/heap
```

p.s. On Ubuntu, make sure to have installed `graphviz` and `gv`

```bash
apt-get install graphviz gv
```

Or `graphviz` on Mac

```bash
brew install graphviz
```

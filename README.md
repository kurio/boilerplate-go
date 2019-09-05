# Go App

Go App is an app written in Go.

## Contributing

Checkout:

```bash
git clone https://github.com/kurio/boilerplate-go.git
```

Prepare the tools:

```bash
cd boilerplate-go
make lint-prepare
make mockery-prepare
```

### Running

To start and stop the HTTP server and scheduler, run:

```bash
make mongo-up
make mysql-up
make docker
make run
make stop
```

To start the HTTP server, run `make run`.

### Testing

For running all tests, use `make test`.

For running only unittests, use `make unittest`.

### Committing Changes

Commit checklist:

1. Run linter with `make lint`
2. Run all tests (see: [Testing](#Testing))
3. Commit with proper, descriptive message (see: [https://karma-runner.github.io/2.0/dev/git-commit-msg.html](https://karma-runner.github.io/2.0/dev/git-commit-msg.html))
4. Create *Pull Request*, describing the changes

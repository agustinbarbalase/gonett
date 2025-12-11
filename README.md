# gonett

Gonett is a Go version for library [Mininet](https://github.com/mininet/mininet/tree/master)

## How use it

For create a container run

```bash
sudo go run cmd/main.go create <name>
```

For list containers run

```bash
sudo go run cmd/main.go list
```

For open a terminal in container run

```bash
sudo go run cmd/main.go exec <name>
```

For delete a container run

```bash
sudo go run cmd/main.go delete <name>
```

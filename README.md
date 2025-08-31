# TUI Worker Pool

This is a TUI visualization for https://gobyexample.com/worker-pools

```console
user@~: $ go run main.go
> 4                    

Agent 001: Finished: "1"
Agent 002: Finished: "foo"
Agent 003: ⣟  Doing work... "3"
Agent 004: 
Agent 005: ⣾  Doing work... "2"
```

## Quickstart

From the project root, run:

```
go mod tidy
go run main.go
ctrl+c
```
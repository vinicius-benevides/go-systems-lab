# Go Systems Lab

Experimentos reproduzíveis sobre Go, concorrência e fundamentos de sistemas.
Cada diretório contém código executável, testes e instruções de execução.

## Experimentos

| Área | Tema | Código |
| --- | --- | --- |
| Concorrência | Goroutines não são threads | [`season-0/concurrency/01-goroutines-not-threads`](season-0/concurrency/01-goroutines-not-threads) |
| Concorrência | Context: timeout, cancelamento e propagação | [`season-0/concurrency/02-context-timeout-cancel`](season-0/concurrency/02-context-timeout-cancel) |
| Orquestração | Temporal: Workflow de saudação com Activity e retry | [`season-0/temporal/01-temporal-go-lab`](season-0/temporal/01-temporal-go-lab) |

## Requisitos

- Go 1.25.11 ou superior

## Validação

```bash
go test ./...
```

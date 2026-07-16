# Temporal: Workflow de saudação

Este laboratório implementa um Workflow durável no [Temporal](https://temporal.io/).
O processo recebe um nome, executa uma Activity que compõe uma saudação e
retorna o resultado ao processo iniciador. A Activity pode falhar na primeira
tentativa para tornar visível o retry automático do Temporal.

```text
starter
   │ inicia GreetingWorkflow
   ▼
Temporal Service ── task queue: greeting-task-queue ──► worker
                                                    │
                                                    ▼
                                            ComposeGreeting Activity
```

## O que o experimento demonstra

- o `starter` inicia e aguarda um `GreetingWorkflow`;
- o `worker` registra o Workflow e a Activity na mesma Task Queue;
- o Workflow mantém apenas orquestração determinística e delega I/O, relógio e
  espera para a Activity;
- falhas transitórias da Activity são retentadas até três vezes, com backoff
  exponencial.

## Requisitos

- Go 1.25.11 ou superior;
- um Temporal Service acessível em `localhost:7233`, no namespace `default`.

Para desenvolvimento local, inicie um servidor Temporal antes de executar os
comandos abaixo. Por exemplo, com a CLI do Temporal:

```bash
temporal server start-dev
```

## Executar

Em um terminal, na raiz do repositório, inicie o Worker:

```bash
go run ./season-0/temporal/01-temporal-go-lab/cmd/worker
```

Em outro terminal, inicie o Workflow:

```bash
go run ./season-0/temporal/01-temporal-go-lab/cmd/starter \
  -name "Vinícius" \
  -language pt-BR \
  -delay 2
```

Exemplo de saída do starter:

```text
Workflow started: WorkflowID=greeting-... RunID=...
Workflow completed
Message:      Olá, Vinícius! Seu primeiro Workflow com Temporal foi concluído com sucesso.
Generated at: 2026-07-16T12:00:00Z
Activity try: 1
```

Os identificadores e o horário variam em cada execução.

## Retry da Activity

Para simular uma falha transitória na primeira tentativa:

```bash
go run ./season-0/temporal/01-temporal-go-lab/cmd/starter \
  -name "Ada" \
  -language en \
  -simulate-failure
```

`ComposeGreeting` falha somente quando `Attempt == 1`. O Workflow configura
até três tentativas, com intervalo inicial de um segundo, coeficiente de
backoff `2` e intervalo máximo de cinco segundos. Portanto, esse cenário deve
concluir na segunda tentativa e imprimir `Activity try: 2`.

## Parâmetros do starter

| Flag | Padrão | Descrição |
| --- | --- | --- |
| `-name` | `Vinícius` | Nome obrigatório da saudação. |
| `-language` | `pt-BR` | `en` ou `english` selecionam a mensagem em inglês; os demais valores usam português. |
| `-delay` | `2` | Duração simulada da Activity em segundos. `0` usa o padrão da Activity (2 s); valores negativos são rejeitados. |
| `-simulate-failure` | `false` | Faz a primeira tentativa da Activity falhar. |
| `-workflow-id` | gerado | Identificador do Workflow; use um valor explícito para rastreá-lo no Temporal. |

## Estrutura

```text
cmd/starter       cliente que inicia e lê o resultado do Workflow
cmd/worker        processo que faz polling da Task Queue
internal/workflows  orquestração determinística e política de retry
internal/activities operação com relógio, espera e log da tentativa
internal/model      contratos de entrada e saída serializados pelo Temporal
internal/shared     constantes compartilhadas, incluindo a Task Queue
```

## Testes e validação

```bash
go test ./...
go vet ./...
```

Os testes de Workflow usam o ambiente de teste do SDK e não exigem um servidor
Temporal em execução. A execução manual do `starter` e `worker`, por outro
lado, exige o serviço acessível.

## Cuidados para evolução

Código de Workflow pode ser reexecutado pelo Temporal durante um replay. Evite
nele chamadas diretas a relógio do sistema, I/O, rede, goroutines e APIs não
determinísticas; coloque esse trabalho em Activities. Alterações incompatíveis
em Workflows já executados exigem estratégia de versionamento antes do deploy.

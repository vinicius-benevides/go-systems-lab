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
- a Activity emite heartbeats enquanto trabalha, o que permite ao Temporal
  detectar Workers que pararam no meio de uma operação longa;
- o Workflow expõe uma Query de status, sem modificar o seu estado durante a
  consulta.

## Requisitos

- Go 1.25.11 ou superior;
- um Temporal Service acessível em `localhost:7233`, no namespace `default`.

Para desenvolvimento local, inicie um servidor Temporal antes de executar os
comandos abaixo. Por exemplo, com a CLI do Temporal:

```bash
temporal server start-dev
```

Por padrão, o projeto usa `localhost:7233`, namespace `default` e a Task Queue
`greeting-task-queue`. Para conectar a outro ambiente, configure as mesmas
variáveis no `starter` e no `worker`:

```bash
export TEMPORAL_ADDRESS="temporal.example.internal:7233"
export TEMPORAL_NAMESPACE="training"
export TEMPORAL_TASK_QUEUE="greeting-training"
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
| `-language` | `pt-BR` | `pt-BR`/`pt` ou `en`/`english`. Outros valores são rejeitados antes de chamar a Activity. |
| `-delay` | `2` | Duração simulada da Activity em segundos. `0` usa 2 s; o intervalo aceito é de 0 a 60 s. |
| `-simulate-failure` | `false` | Faz a primeira tentativa da Activity falhar. |
| `-workflow-id` | gerado | Identificador do Workflow; use um valor explícito para rastreá-lo no Temporal. |
| `-wait-timeout` | `0` | Máximo para o cliente esperar o resultado. `0` espera indefinidamente; expirar não cancela o Workflow. |

## Acompanhar uma execução

Os tipos registrados usam nomes explícitos e versionados (`greeting.v1` e
`greeting.compose.v1`). Assim, um refactor do nome da função Go não muda o
contrato que o Temporal grava no histórico.

Enquanto o Workflow estiver em execução, sua Query `greeting.status` retorna
uma estrutura com a fase (`running`, `completed` ou `failed`) e, quando
disponível, o resultado. A Query pode ser feita pela API do SDK ou pela UI do
Temporal. Para experimentar o comportamento assíncrono, inicie-o com um tempo
curto de espera:

```bash
go run ./season-0/temporal/01-temporal-go-lab/cmd/starter \
  -name "Ada" -delay 30 -wait-timeout 1s
```

O starter retorna após um segundo, mas a execução durável continua no servidor
e pode ser inspecionada na Temporal UI pelo Workflow ID exibido.

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

Este laboratório já separa os nomes versionados do código Go, mas uma mudança
na lógica de decisão de um Workflow que possua históricos em aberto ainda deve
usar versionamento de Workflow (por exemplo, `workflow.GetVersion`) e um plano
de migração antes de chegar a produção.

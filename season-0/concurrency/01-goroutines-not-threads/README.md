# Goroutines não são threads

Este experimento compara a execução sequencial e concorrente de tarefas que simulam espera por I/O com `time.Sleep`.

## Hipótese

Com 20 tarefas independentes e 100 ms de latência por tarefa:

- a execução sequencial deve levar aproximadamente `20 × 100 ms = 2 s`;
- a execução concorrente deve se aproximar de `100 ms`, acrescida do custo de criar, agendar e sincronizar as goroutines.


## Executar

Na raiz do repositório:

```bash
go run ./season-0/concurrency/01-goroutines-not-threads \
  -tasks 20 \
  -latency 100ms
```

Exemplo ilustrativo de saída:

```csv
tasks,latency_ms,mode,elapsed_ms,goroutines_observed
20,100.000,sequential,2006.214,1
20,100.000,goroutines,100.473,21
```

Os valores variam conforme máquina, sistema operacional, versão do Go e carga
do ambiente. A saída CSV pode ser salva para análise:

```bash
go run ./season-0/concurrency/01-goroutines-not-threads > results.csv
```

## Resultado medido

Uma execução com Go 1.25.11 em Linux, usando os parâmetros padrão, produziu:

```csv
tasks,latency_ms,mode,elapsed_ms,goroutines_observed
20,100.000,sequential,2011.102,1
20,100.000,goroutines,102.293,21
```

Os dados são uma única amostra do ambiente de desenvolvimento, não uma garantia
de desempenho. O benchmark permite medições repetidas.

## O que é medido

`elapsed_ms` inclui todo o trabalho de cada modo. No modo concorrente, isso
inclui criação e sincronização das goroutines. `goroutines_observed` é um
snapshot obtido quando todas as goroutines do experimento já foram criadas e
aguardam o sinal de início; não é uma medição contínua nem representa threads
do sistema operacional.

Uma barreira de início garante que o snapshot não dependa de tarefas rápidas
terminarem antes da observação. Ela existe para tornar o experimento
reproduzível, não como recomendação de implementação para código de produção.

## Benchmark

```bash
go test ./season-0/concurrency/01-goroutines-not-threads \
  -bench=. \
  -benchtime=10x \
  -benchmem
```

O benchmark usa latência de 2 ms para reduzir seu tempo de execução. Como
`time.Sleep` domina o resultado, ele serve para comparar os dois modos deste
experimento, não para medir o custo isolado de uma goroutine.

## Interpretação

Uma thread é gerenciada pelo sistema operacional. Uma goroutine é gerenciada
pelo runtime do Go, que agenda muitas goroutines sobre threads do sistema.
Goroutines começam com stacks pequenas e expansíveis, mas continuam consumindo
memória e recursos do scheduler.

O ganho deste exemplo vem da sobreposição de períodos de espera independentes.
Para cargas CPU-bound, o resultado depende de paralelismo disponível,
`GOMAXPROCS`, contenção, algoritmo e perfil da carga.

## Cuidados em produção

Criar goroutines sem limite pode causar consumo excessivo de memória, pressão
no scheduler, ausência de backpressure e goroutine leaks. Sistemas reais
normalmente também precisam de limites de concorrência, cancelamento com
`context.Context`, tratamento de erros e observabilidade.

# Desafio Backend — Processamento Distribuído de Apostas em Go

Implemente um serviço em **Go**, usando **Uber Fx**, para receber e processar operações financeiras de provedores de jogos. O desafio explora situações comuns em sistemas de iGaming: disputa pelo saldo de uma carteira, mensagens repetidas, dependências fora de ordem e recuperação de processos interrompidos.

A implementação e as decisões de arquitetura ficam a cargo do candidato.

## 1. Objetivo

A aplicação deve oferecer uma API HTTP e um consumidor de mensagens que movimentem carteiras de jogadores com garantias equivalentes. Demonstre que o resultado financeiro continua correto com várias instâncias em execução e falhas entre as etapas do processamento.

A avaliação considera precisão monetária, integridade do histórico financeiro, idempotência durável, concorrência, recuperação e capacidade de explicar as escolhas feitas. Organização de código importa, mas precisa sustentar essas garantias.

## 2. Autenticação

Autenticação é opcional e não recebe pontuação própria. Priorize os cenários financeiros e de falha descritos neste documento.

Caso implemente autenticação, integre um provedor de identidade externo compatível com OIDC. Não crie um cadastro local de credenciais e senhas para este desafio.

Caso deixe essa integração para depois, registre o desenho proposto em `ARCHITECTURE.md` e mantenha uma extensão identificável, como um middleware HTTP ou uma interface `ProviderIdentityResolver`.

Os health checks são públicos. Considere a fila um canal interno confiável, preservando a validação de domínio da identidade do provedor e de sua relação com a operação recebida.

## 3. Ambiente de execução e falhas

Cada operação externa pertence a um jogador, uma carteira, um jogo, um provedor e uma rodada. Os tipos externos são `BET`, `WIN`, `LOSS`, `REFUND` e `ROLLBACK`.

Assuma entrega **at-least-once** e prepare a solução para:

- recebimento repetido de uma mesma operação, inclusive por HTTP e SQS;
- chegada de uma reversão antes da transação que ela referencia;
- processamento simultâneo de operações da mesma carteira;
- encerramento abrupto antes ou depois de um commit;
- publicação repetida de um evento de integração;
- indisponibilidade temporária do PostgreSQL ou do SQS.

Nenhuma dessas situações pode gerar movimentação duplicada, saldo negativo ou perda de um evento cujo registro foi confirmado no banco.

## 4. Stack e uso do Uber Fx

### Tecnologias obrigatórias

| Responsabilidade | Tecnologia |
| --- | --- |
| Linguagem e compilação | Go; declare a versão utilizada em `go.mod` e no Dockerfile |
| Dependências | Go Modules, com `go.mod` e `go.sum` versionados |
| Composição da aplicação | [Uber Fx](https://pkg.go.dev/go.uber.org/fx), pacote `go.uber.org/fx` |
| HTTP | `net/http` ou um roteador Go à sua escolha |
| Persistência | PostgreSQL |
| Mensageria | AWS SQS, executado localmente com LocalStack ou MiniStack |
| Ambiente local | Docker Compose |
| Evolução do banco | Migrations versionadas, com aplicação e reversão documentadas |
| Testes | `testing` e `go test`, incluindo execução com `-race` |

### Acesso ao banco

Prefira `pgx` com SQL explícito; `sqlc` pode ser usado para gerar código a partir das consultas. `database/sql` e GORM também são aceitos, desde que as transações, os locks e as constraints continuem verificáveis.

Explique em `ARCHITECTURE.md` a escolha da biblioteca, a representação persistida de `Money` e como todos os repositórios de uma operação compartilham a mesma transação SQL.

### Uso obrigatório do Fx

O Fx deve compor a aplicação de verdade: configuração, conexões, repositórios, casos de uso, handlers e workers devem receber dependências por construtores. Utilize `fx.Module` para agrupamentos coerentes, `fx.Provide` para registrar construtores e `fx.Invoke` para conectar os componentes que precisam participar da execução. Um construtor registrado só será executado se seu resultado for necessário no grafo. Consulte a [referência do Fx](https://pkg.go.dev/go.uber.org/fx).

O gerenciamento de recursos deve usar `fx.Lifecycle`:

- `OnStart`: preparar recursos e iniciar servidor e workers;
- `OnStop`: interromper novas entradas, finalizar ou liberar trabalho em andamento e fechar os recursos;
- hooks devem respeitar seus prazos; loops permanentes devem rodar em goroutines gerenciadas;
- cada worker deve ter cancelamento e término observável; não use o contexto temporário de `OnStart` como contexto de toda a execução do worker.

Os hooks de encerramento executam na ordem inversa de registro. Organize as dependências para que o banco continue disponível durante a finalização do trabalho. Veja o [ciclo de vida do Fx](https://uber-go.github.io/fx/lifecycle.html).

Mantenha o domínio independente de Fx, HTTP, SQS e bibliotecas de persistência. Receber dependências pelo construtor é suficiente para testar os casos de uso sem iniciar a aplicação inteira.

### Organização sugerida

Adapte os pacotes ao desenho escolhido. Esta árvore ilustra responsabilidades; não exige criar todas as pastas nem abstrações sem uso.

```text
cmd/
  server/main.go
internal/
  app/                 # composição com Fx
  config/
  domain/              # Money, Wallet, transações e eventos
  application/         # casos de uso e interfaces necessárias
  adapters/
    http/
    postgres/
    sqs/
  workers/             # consumo, outbox e referências pendentes
  observability/
migrations/
tests/integration/
Dockerfile
compose.yaml
go.mod
go.sum
README.md
ARCHITECTURE.md
```

## 5. Garantias obrigatórias

1. Dinheiro não pode passar por `float32` ou `float64`, nem durante parsing, cálculo, serialização ou persistência.
2. Idempotência precisa sobreviver ao reinício de todos os processos; um mapa, cache local ou `sync.Mutex` não oferece essa garantia.
3. Locks em memória e a deduplicação do SQS FIFO não substituem a coordenação entre instâncias no banco.
4. Eventos externos só podem ser publicados depois da confirmação da transação que os originou.
5. O ledger deve ser append-only: correções financeiras exigem novos lançamentos.
6. Carteiras independentes devem poder avançar em paralelo. Não serialize todas elas com um lock global.
7. Ler o saldo, calcular e gravar o resultado exige uma estratégia que impeça lost updates.
8. Unicidade, não negatividade e proteção contra alteração do ledger devem existir também no banco, por constraints, índices e mecanismos adequados de imutabilidade. Validação apenas em Go é insuficiente.

## 6. Modelagem em Go

### Encapsulamento e erros

Modele as entidades com structs, estado interno não exportado e métodos que representem transições de negócio. Use funções construtoras, como `NewMoney`, `OpenWallet` e `NewWagerTransaction`, que retornem erros quando necessário. A visibilidade em Go é por pacote; organize esses limites para proteger as invariantes. A abordagem de construtores está descrita em [Effective Go](https://go.dev/doc/effective_go#composite_literals).

Separe criação de reidratação: funções como `RehydrateWallet` devem reconstruir o estado persistido sem executar novamente movimentações, transições ou emissão de eventos.

Defina como valores zero inválidos serão detectados. Campos não exportados não impedem, por si só, a existência de `Money{}` ou de uma entidade sem inicialização válida.

Use erros explícitos para validação e regras de negócio, com classificação por tipos ou `errors.Is`/`errors.As`. Não use `panic` como fluxo esperado. Operações de I/O devem receber `context.Context` e respeitar cancelamento e timeout.

Os trechos a seguir ilustram dados e contratos, não constituem uma implementação pronta. Nomes e assinaturas podem mudar se as garantias forem mantidas.

### 6.1. Money

Uma representação possível para o escopo de duas casas decimais é um inteiro de unidades mínimas:

```go
type MoneyDTO struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type Money struct {
	minorUnits int64
	currency   string
}
```

Implemente criação a partir de string decimal, zero por moeda, soma, subtração, negação, comparação e conversão para DTO. Cada operação aritmética deve devolver um novo valor, sem modificar os operandos.

Também é aceita uma biblioteca decimal de precisão exata. Justifique a alternativa escolhida e teste seus limites.

- O contrato externo recebe e devolve valores como `{"amount":"25.00","currency":"BRL"}`.
- Use escala fixa de duas casas e código de moeda ISO 4217. Não converta por ponto flutuante para chegar ao formato final.
- Rejeite valores vazios, `NaN`, `Infinity`, notação científica, escala excedente e valores negativos nas entradas financeiras externas.
- Não arredonde silenciosamente uma entrada inválida. Caso aceite formas equivalentes, documente a normalização anterior ao hash de idempotência.
- Aritmética e comparação de valores monetários exigem moedas compatíveis.
- Se utilizar `int64`, trate overflow no parsing, na soma, na subtração e na negação.
- O domínio pode representar valores negativos para diferenças ou cálculos internos; a carteira continua proibida de terminar negativa.
- A persistência deve preservar exatamente valor e moeda, por exemplo com unidades mínimas em `BIGINT` ou decimal em `NUMERIC`.

É permitido operar apenas em BRL nos cenários principais, desde que o tipo carregue a moeda e existam testes de incompatibilidade entre moedas.

### 6.2. Wallet

A carteira é a raiz do agregado financeiro. Deve carregar identidade, jogador, moeda, saldo, versão e instantes de criação e atualização.

Exponha criação, reidratação e operações de débito/crédito, mantendo a alteração do saldo sob controle do agregado e da transação SQL.

- O par `(playerId, currency)` identifica uma única carteira.
- Débitos precisam preservar saldo maior ou igual a zero.
- A moeda de cada movimentação deve coincidir com a da carteira.
- Cada mudança financeira exige o lançamento correspondente no ledger, confirmado junto com o saldo.
- A versão inicial é `1`; depois da criação, incremente-a apenas quando houver mudança de saldo.
- Disputas entre escritores não podem descartar uma atualização confirmada.

Você pode usar a versão para controle otimista ou escolher outra estratégia, explicando a decisão.

### 6.3. WagerTransaction

Use tipos nomeados para os valores de domínio, por exemplo:

```go
type TransactionKind string

const (
	KindOpening  TransactionKind = "OPENING"
	KindBet      TransactionKind = "BET"
	KindWin      TransactionKind = "WIN"
	KindLoss     TransactionKind = "LOSS"
	KindRefund   TransactionKind = "REFUND"
	KindRollback TransactionKind = "ROLLBACK"
)

type TransactionStatus string

const (
	StatusPending          TransactionStatus = "PENDING"
	StatusPendingReference TransactionStatus = "PENDING_REFERENCE"
	StatusProcessed        TransactionStatus = "PROCESSED"
	StatusRejected         TransactionStatus = "REJECTED"
	StatusFailed           TransactionStatus = "FAILED"
)
```

Para operações externas, a transação registra os identificadores interno e externo, provedor, chave de idempotência, hash do payload, carteira, jogador, rodada, jogo, tipo, `Money`, referência externa opcional, estado e timestamps. Quando aplicável, persista também a referência interna resolvida, o código de falha e o resultado financeiro retornado ao provedor.

Ela começa em `PENDING`. Seus métodos devem expressar processamento, espera por referência, rejeição e falha permanente, validando as transições permitidas.

| Estado | Significado |
| --- | --- |
| `PENDING` | Registro aceito, com processamento ainda não concluído |
| `PENDING_REFERENCE` | A aplicação depende de uma referência ainda indisponível |
| `PROCESSED` | Operação concluída com sucesso; estado terminal |
| `REJECTED` | Operação recusada por uma regra de negócio; estado terminal |
| `FAILED` | Falha permanente de infraestrutura registrada para auditoria; estado terminal |

Uma transação terminal não deve sofrer novas transições. Replay consulta seu resultado persistido sem reaplicar a operação. Documente a máquina de estados e como distingue falhas transitórias de falhas permanentes.

Se a implementação confirmar uma operação em `PENDING` antes de executá-la, deve persistir também o necessário para sua retomada por um worker. Qualquer `PENDING` confirmado precisa ser recuperável por outra instância após uma interrupção; essa obrigação não se limita a `PENDING_REFERENCE`. É permitido concluir operações sem dependências de forma síncrona, sem persistir uma etapa intermediária de aceite.

`OPENING` é reservado à abertura interna de carteira. Rejeite esse tipo quando enviado por HTTP ou SQS.

Uma transação `OPENING` tem identidade interna estável, carteira, jogador, moeda, valor, estado e timestamps. Provedor, ID externo, chave de idempotência externa, hash do payload externo, rodada, jogo e referência não se aplicam a essa origem. Modele essa distinção no schema e nas constraints, sem exigir que `POST /wallets` forneça metadados de uma aposta. A unicidade da carteira e a atomicidade da abertura devem impedir um segundo crédito inicial.

### 6.4. WalletLedgerEntry

Cada lançamento registra `id`, `walletId`, `transactionId`, direção (`DEBIT` ou `CREDIT`), valor, saldo anterior, saldo posterior e instante de criação.

A construção deve validar a equação financeira do lançamento. O tipo não deve oferecer setters nem métodos de edição, e getters não devem permitir mutação indireta de seu estado.

Imponha no banco a unicidade de `(walletId, transactionId)` e a proteção contra edição ou exclusão. `LOSS` e operações rejeitadas não produzem lançamentos. Um ledger de partidas dobradas é opcional.

### 6.5. Inbox e outbox

| Registro | Informações e comportamento esperados |
| --- | --- |
| Inbox | Identidade da mensagem e do consumidor, hash, recebimento e conclusão; unicidade de `(consumerName, messageId)` |
| Outbox | Identidade estável do evento, agregado, tipo, payload, ocorrência, tentativas, próximo envio e publicação; suporte a retry com backoff |

Na entrada por SQS, o registro da inbox e a conclusão durável do tratamento devem compartilhar a transação SQL das alterações de domínio, do ledger e dos eventos correspondentes. Uma referência pendente pode ter sua mensagem de entrada concluída após a pendência estar persistida; o worker de referências assume a continuidade.

## 7. Operações e referências

| Tipo | Movimentação | Condição |
| --- | --- | --- |
| `BET` | Débito | Exige valor positivo e saldo suficiente |
| `WIN` | Crédito | Exige valor positivo; pode informar uma aposta da mesma rodada como referência |
| `LOSS` | Sem movimentação | Exige `money.amount` igual a `"0.00"`; não cria ledger nem altera a versão da carteira |
| `REFUND` | Crédito | Devolve integralmente o valor de uma `BET` processada |
| `ROLLBACK` | Movimento contrário ao original | Desfaz integralmente uma `BET`, `WIN` ou `REFUND` processada |

Para `REFUND` e `ROLLBACK`, `referenceExternalTransactionId` é obrigatório. Resolva-o pelo par `(providerId, referenceExternalTransactionId)`, nunca presumindo que seja um ID interno.

A operação e sua referência devem concordar em provedor, jogador, carteira, moeda e rodada. O valor da reversão precisa ser igual ao valor referenciado; reversões parciais não fazem parte do desafio.

Para este desafio, zero é aceito no saldo inicial e em `LOSS`; `BET`, `WIN`, `REFUND` e `ROLLBACK` exigem valor maior que zero. `LOSS` continua exigindo a moeda da carteira e, quando processado, produz `WagerTransactionProcessed`, sem `WalletBalanceChanged`.

Garanta que uma referência não receba duas reversões bem-sucedidas do mesmo tipo. Documente como trata combinações de `REFUND` e `ROLLBACK` sobre a mesma aposta, preservando a coerência financeira e impedindo devolução duplicada do mesmo débito.

Uma reversão que precisaria debitar mais que o saldo disponível deve ser rejeitada e auditável. Seu código de falha deve ser diferente daquele usado para uma aposta sem saldo.

### Referências ainda indisponíveis

Persista a operação como `PENDING_REFERENCE` quando a referência ainda não tiver chegado. Um worker deve tentar novamente com backoff exponencial, inclusive após reinicialização da aplicação.

Defina um número máximo de tentativas ou TTL. Quando esgotado, finalize como `REJECTED`, informando um código de referência não encontrada e produzindo o evento de rejeição. Explique também o comportamento quando a referência existe, mas ainda está pendente ou terminou sem sucesso.

Toda rejeição deve fornecer um `failureCode` estável, documentado e interpretável por máquina. O provedor precisa conseguir distinguir uma entrada corrigível de um resultado definitivo.

## 8. Concorrência

A coordenação deve ocorrer por carteira. Escolha locking pessimista, controle otimista com retry limitado, atualização atômica condicionada ou uma combinação justificável.

As garantias precisam continuar válidas com pelo menos três processos independentes, cada um com suas próprias conexões e memória. Goroutines dentro de um único processo não demonstram essa propriedade.

Teste obrigatório: uma carteira com **100.00 BRL** recebe, ao mesmo tempo, duas apostas distintas de **80.00 BRL**.

O resultado deve conter uma aposta processada, uma rejeição por saldo insuficiente, saldo final de **20.00 BRL** e um único débito no ledger. Reenvios não podem alterar esse resultado. Carteiras diferentes devem continuar sendo processadas em paralelo.

## 9. Contratos HTTP

### Abertura de carteira

```http
POST /wallets
Content-Type: application/json
```

```json
{
  "playerId": "0192f28f-5dc0-7d58-bdb2-814ad6a0f4a1",
  "initialBalance": { "amount": "1000.00", "currency": "BRL" }
}
```

Exemplo de resposta:

```json
{
  "id": "0192f291-27dd-7d3f-8071-5f8685deef37",
  "playerId": "0192f28f-5dc0-7d58-bdb2-814ad6a0f4a1",
  "balance": { "amount": "1000.00", "currency": "BRL" },
  "version": 1
}
```

Uma abertura com saldo positivo deve criar `OPENING` em `PROCESSED`, seu lançamento de crédito e os registros de outbox para `WagerTransactionProcessed` e `WalletBalanceChanged` no mesmo commit da carteira. Esses eventos de origem interna não exigem os metadados externos inaplicáveis; a versão da carteira nessa abertura é `1`. Saldo inicial zero não cria `OPENING`, ledger nem esses eventos financeiros. Tentar abrir outra carteira para o mesmo jogador e moeda deve resultar em conflito.

### Leitura

```http
GET /wallets/:walletId
GET /wallets/:walletId/ledger?cursor=...&limit=50
GET /wagering/transactions/:transactionId
GET /providers/:providerId/wagering/transactions/:externalTransactionId
```

A paginação do ledger deve usar cursor opaco e ordenação estável. As consultas de transação devem permitir acompanhar pendências e consultar códigos de rejeição ou falha.

### Envio de operação

```http
POST /wagering/transactions
Content-Type: application/json
Idempotency-Key: provider-a:transaction-123
```

```json
{
  "providerId": "provider-a",
  "externalTransactionId": "transaction-123",
  "playerId": "0192f28f-5dc0-7d58-bdb2-814ad6a0f4a1",
  "walletId": "0192f291-27dd-7d3f-8071-5f8685deef37",
  "roundId": "round-987",
  "gameId": "fortune-chimp",
  "kind": "BET",
  "money": { "amount": "25.00", "currency": "BRL" }
}
```

Exemplo após processamento:

```json
{
  "transactionId": "0192f298-345e-7e38-af88-e43f851a819d",
  "status": "PROCESSED",
  "balance": { "amount": "975.00", "currency": "BRL" },
  "idempotentReplay": false
}
```

Para reversões, acrescente `referenceExternalTransactionId` ao corpo.

O header `Idempotency-Key` é obrigatório. O cliente pode construí-lo como `{providerId}:{externalTransactionId}`, mas o servidor não deve substituir silenciosamente uma chave recebida por outra calculada.

Persista um hash determinístico dos campos de negócio, usando JSON canônico com ordenação de chaves. Exclua a chave de idempotência e os metadados de transporte desse cálculo. Documente algoritmo, campos e normalizações, garantindo equivalência entre HTTP e SQS.

- Chave e conteúdo equivalentes: retorne o resultado persistido, com `idempotentReplay: true`.
- Chave reutilizada com conteúdo diferente: devolva conflito.
- Uma operação financeira identificada por `(providerId, externalTransactionId)` não pode ser reaplicada usando outra chave.
- Para operações concluídas, o replay deve devolver o saldo observado no processamento original, mesmo que a carteira já tenha recebido outras movimentações.

Documente os códigos HTTP e os corpos de resposta para entrada inválida, conflito, rejeição de negócio, processamento pendente e indisponibilidade transitória. Essas situações precisam ser distinguíveis pelo contrato.

### Reconciliação

```http
POST /wallets/:walletId/reconciliation
```

Exemplo considerando apenas a abertura e a aposta apresentadas acima:

```json
{
  "walletId": "0192f291-27dd-7d3f-8071-5f8685deef37",
  "storedBalance": { "amount": "975.00", "currency": "BRL" },
  "calculatedBalance": { "amount": "975.00", "currency": "BRL" },
  "difference": { "amount": "0.00", "currency": "BRL" },
  "consistent": true,
  "checkedEntries": 2
}
```

Reconstrua o saldo a partir do ledger, incluindo a abertura. Faça a comparação em uma visão consistente dos dados, para evitar falsos desvios durante movimentações concorrentes. Defina `difference` como saldo armazenado menos saldo reconstruído.

Uma divergência deve aparecer na resposta, nos logs e em uma métrica. Não ajuste o saldo automaticamente para esconder a inconsistência.

### Saúde da aplicação

```http
GET /health/live
GET /health/ready
```

Liveness indica se o processo está vivo. Readiness verifica a disponibilidade das dependências PostgreSQL e SQS. Ambos dispensam autenticação.

## 10. Consumidor SQS

Provisione as filas `wager-transactions.fifo` e `wager-transactions-dlq.fifo`, incluindo a configuração de redrive.

Exemplo de corpo de mensagem:

```json
{
  "messageId": "msg-123",
  "type": "WagerTransactionRequested",
  "occurredAt": "2026-09-08T12:00:00.000Z",
  "data": {
    "providerId": "provider-a",
    "externalTransactionId": "transaction-123",
    "idempotencyKey": "provider-a:transaction-123",
    "playerId": "0192f28f-5dc0-7d58-bdb2-814ad6a0f4a1",
    "walletId": "0192f291-27dd-7d3f-8071-5f8685deef37",
    "roundId": "round-987",
    "gameId": "fortune-chimp",
    "kind": "BET",
    "money": { "amount": "25.00", "currency": "BRL" }
  }
}
```

O adaptador SQS deve chamar o mesmo caso de uso do adaptador HTTP. `data.idempotencyKey` exerce o papel do header; a inbox acrescenta deduplicação de transporte à idempotência financeira.

- Use o `messageId` do envelope como identidade durável da mensagem para o consumidor e verifique seu hash em reentregas.
- Remova a mensagem da fila somente após o commit do seu tratamento durável.
- Rejeições de negócio confirmadas são terminais e permitem a remoção da mensagem.
- Falhas transitórias exigem retry com backoff; erros permanentes ou tentativas esgotadas devem chegar à DLQ.
- Documente limites de tentativas, visibility timeout e tratamento de mensagens inválidas.
- Em `SIGTERM`, pare de buscar trabalho e conclua o processamento em andamento dentro do prazo, ou libere sua visibilidade para reentrega segura.

Explique a configuração de `MessageGroupId` e `MessageDeduplicationId`. A ordenação por carteira pode reduzir disputas, mas a correção precisa permanecer no banco mesmo quando a API recebe operações concorrentes.

## 11. Publicação com transactional outbox

O commit de uma operação deve incluir seu estado, a alteração de saldo quando houver, o ledger, a inbox quando aplicável e os registros de eventos. Nenhuma confirmação parcial pode deixar essas informações divergentes.

Um worker separado publica os registros pendentes da outbox. Ele deve suportar múltiplos publishers, disputa por registros, backoff e recuperação de trabalho abandonado.

Demonstre dois pontos de falha: interrupção depois do commit e antes da publicação; e interrupção depois da publicação e antes de marcar o registro como publicado. No primeiro caso, outra instância deve publicar o evento. No segundo, a repetição deve preservar o `eventId`, permitindo deduplicação pelo consumidor.

Documente e provisione o destino dos eventos de saída. Não os envie à fila de comandos sem um contrato explícito de roteamento e consumo.

### Eventos exigidos

| Evento | Gatilho |
| --- | --- |
| `WagerTransactionProcessed` | Conclusão bem-sucedida de uma operação, incluindo `LOSS` |
| `WagerTransactionRejected` | Rejeição definitiva por regra de negócio |
| `WalletBalanceChanged` | Alteração efetiva do saldo |
| `WagerTransactionPendingReference` | Registro de espera pela referência |

Use tipos concretos por evento e composição em Go. O envelope deve carregar `eventId`, `eventType`, `aggregateId`, `correlationId`, `causationId` opcional, `occurredAt`, `version` e `data` tipado.

Por exemplo, o payload de mudança de saldo pode ter o seguinte formato:

```go
type WalletBalanceChangedData struct {
	WalletID      string   `json:"walletId"`
	TransactionID string   `json:"transactionId"`
	Direction     string   `json:"direction"`
	Money         MoneyDTO `json:"money"`
	BalanceBefore MoneyDTO `json:"balanceBefore"`
	BalanceAfter  MoneyDTO `json:"balanceAfter"`
	WalletVersion int64    `json:"walletVersion"`
}
```

O construtor de cada evento deve determinar seu tipo e versão. Serialize o instante em UTC, no formato RFC 3339, e os valores monetários como DTOs com strings decimais. O payload persistido na outbox deve ser um snapshot estável, sem referências mutáveis para o agregado.

## 12. Observabilidade

Produza logs JSON com os identificadores disponíveis para rastrear a operação: `correlationId`, `messageId`, `transactionId`, `walletId` e `providerId`. Não registre credenciais, dados sensíveis ou payloads financeiros completos.

Exponha métricas para resultados por status, duplicatas, retries, DLQ, conflitos de concorrência, atraso da outbox, latência de processamento e divergências de reconciliação.

Inclua os health checks definidos na API. Tracing com OpenTelemetry e dashboards são diferenciais opcionais.

## 13. Verificação obrigatória

### Testes unitários

Cubra parsing e operações de `Money`, escala, limites numéricos, entradas inválidas, incompatibilidade de moedas, invariantes da carteira, transições de estado, regras dos cinco tipos externos e conflito de payload para a mesma chave. Inclua a política de valores zero de cada tipo e a abertura interna com seus metadados e eventos.

Prefira testes orientados a tabela quando ajudarem a explicitar os casos. Verifique resultados e invariantes; cobertura percentual isolada não demonstra correção.

### Testes de integração

Execute PostgreSQL e LocalStack ou MiniStack em containers reais. Verifique migrations, constraints, imutabilidade do ledger, atomicidade financeira, inbox, reentrega, outbox concorrente, retry, DLQ e recuperação após reinicialização.

Adicione uma verificação da composição Fx e de seu início e encerramento, incluindo liberação de recursos dos workers. Não substitua toda a infraestrutura por mocks.

### Testes de concorrência e recuperação

1. Envie a mesma aposta 50 vezes em paralelo e comprove um único débito.
2. Execute a disputa das duas apostas de 80.00 sobre saldo de 100.00.
3. Processe carteiras distintas simultaneamente.
4. Repita cenários relevantes com pelo menos três instâncias independentes.
5. Interrompa um consumidor depois do commit e antes da remoção da mensagem; valide a reentrega.
6. Execute dois publishers disputando a mesma outbox e valide a recuperação de publicação.
7. Entregue `REFUND` ou `ROLLBACK` antes da referência e comprove a resolução posterior ou a rejeição por expiração.
8. Reinicie a aplicação e verifique que idempotência, pendências e consistência financeira foram preservadas. Se houver aceite assíncrono, interrompa o processo após confirmar `PENDING` e antes de executar a operação; outra instância deve retomá-la.

Ao final, confira o saldo armazenado contra a soma de créditos menos débitos do ledger. Inclua cenários que cruzem HTTP e SQS para a mesma operação.

Nos testes de duplicidade, comprove que requisições ou reentregas repetidas chegaram à aplicação. A deduplicação de envio do broker não pode ser a única responsável por produzir um único efeito financeiro.

Execute `go test -race` nos testes aplicáveis. O detector de races complementa os testes, mas não prova a ausência de anomalias no banco ou entre processos.

## 14. Critérios de avaliação

| Critério | Pontos | Evidência esperada |
| --- | ---: | --- |
| Integridade financeira | 20 | Precisão, invariantes, reversões e reconciliação confiáveis |
| Concorrência | 20 | Coordenação entre processos e ausência de atualizações perdidas |
| Idempotência | 15 | Persistência, detecção de conflito e reprodução do resultado original |
| Mensageria e recuperação | 15 | Inbox, outbox, retries, DLQ e encerramento seguro |
| Modelagem e arquitetura | 10 | Encapsulamento em Go, limites de responsabilidade e composição com Fx |
| Testes | 10 | Cenários reais de infraestrutura, paralelismo e interrupção |
| Observabilidade | 5 | Diagnóstico por logs, métricas e health checks |
| Documentação | 5 | Execução reproduzível e decisões técnicas explicadas |
| **Total** | **100** | |

São eliminatórios: cálculo monetário em ponto flutuante, saldo negativo por concorrência, movimentação duplicada, idempotência restrita à memória, dependência de uma única instância para funcionar corretamente, publicação anterior ao commit, ausência de ledger auditável ou substituição integral de PostgreSQL e SQS por mocks nos testes.

Como diferenciais, considere partidas dobradas, tracing ou um experimento de carga. Se fizer teste de carga, forneça um comando reproduzível e descreva ambiente, metodologia, throughput, percentis p50/p95/p99, erros, conflitos e atraso da outbox. Não há uma meta mínima de requisições por segundo.

## 15. Entrega

Entregue o código, migrations, ambiente Docker Compose e instruções suficientes para outra pessoa reproduzir a solução a partir de um checkout limpo.

O `README.md` da solução deve explicar pré-requisitos, variáveis de ambiente, inicialização das filas, aplicação e reversão das migrations, execução da aplicação, exemplos de chamadas e comandos de teste. Inclua `.env.example` com valores locais de exemplo, sem segredos reais.

No `ARCHITECTURE.md`, registre as decisões sobre dinheiro, transações, idempotência, locks, referências pendentes, reversões, inbox/outbox, autenticação, uso do Fx e shutdown. Explicite limitações, interpretações adotadas e trabalho não concluído.

Disponibilize os comandos abaixo, ou equivalentes claramente documentados. Eles são uma expectativa para a implementação entregue, não arquivos já presentes neste repositório de enunciado.

```sh
docker compose up --build
go test ./...
go test -race ./...
go vet ./...
```

Documente separadamente como preparar as dependências dos testes e executar integração, múltiplas instâncias e simulações de falha. Se utilizar build tags, informe os comandos correspondentes.

O código deve estar formatado com `gofmt`, as dependências devem ser reproduzíveis e os testes precisam permitir verificar as garantias descritas neste desafio.

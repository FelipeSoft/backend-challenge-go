# Backend Challenge Go

Guia rápido para configurar e rodar o ambiente de desenvolvimento multi-container (APIs, consumers, publisher, PostgreSQL, LocalStack, Keycloak e migrações de banco de dados).

---

## Como Rodar o Projeto

Siga os passos abaixo para subir todo o ambiente de forma automatizada:

1. **Configurar as Variáveis de Ambiente:** Passo 1.
Clone o arquivo de exemplo e renomeie-o para `.env` na raiz do projeto:

```bash
cp .env.example .env

```


2. **Adicionar o Token do LocalStack:** Passo 2.
Abra o arquivo `.env` recém-criado e insira uma chave de acesso válida do LocalStack na variável correspondente:

```env
LOCALSTACK_AUTH_TOKEN=seu_token_valido_aqui

```

*(Você pode obter esse token acessando o seu painel em [app.localstack.cloud](https://app.localstack.cloud/?utm_source=gemini)).*


3. **Subir o Ambiente com Docker:** Passo 3.
Execute o Docker Compose forçando o build inicial do zero:

```bash
docker compose up --build -d

```

*Nota: O container de migração rodará automaticamente antes de liberar as aplicações.*


4. **Obter o Token de Acesso do Keycloak:** Passo 4.
Para autenticar suas requisições nas APIs protegidas, execute o comando cURL abaixo para gerar o seu **Bearer Token**:

```bash
curl --location 'http://localhost:8080/realms/jungle-gaming/protocol/openid-connect/token' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'client_id=backend-client' \
--data-urlencode 'client_secret=my-secret-key' \
--data-urlencode 'grant_type=client_credentials'

```


5. **Testar os Endpoints:** Passo 5.
Com o ambiente containerizado de pé e o token em mãos, basta começar a testar os endpoints da API utilizando o seu cliente HTTP preferido (Postman, Insomnia, cURL, etc.), passando o token obtido no cabeçalho de autorização (`Authorization: Bearer <seu-token>`).
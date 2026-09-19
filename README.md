# Go Backend Challenge

Quick guide to setting up and running the multi-container development environment (APIs, consumers, publisher, PostgreSQL, LocalStack, Keycloak, and database migrations).

---

## How to Run the Project

Follow the steps below to spin up the entire environment automatically:

1. **Configure the Environment Variables:** Step 1.
Clone the example file and rename it to `.env` in the project root:

```bash
cp .env.example .env

```

2. **Add the LocalStack Token:** Step 2.
Open the newly created `.env` file and insert a valid LocalStack access key into the corresponding variable:

```env
LOCALSTACK_AUTH_TOKEN=your_valid_token_here

```

*(You can get this token by accessing your dashboard at [app.localstack.cloud](https://app.localstack.cloud/?utm_source=gemini)).*

3. **Spin Up the Environment with Docker:** Step 3.
Run Docker Compose, forcing an initial build from scratch:

```bash
docker compose up --build -d

```

*Note: The migration container will run automatically before the applications are released.*

4. **Get the Keycloak Access Token:** Step 4.
To authenticate your requests to the protected APIs, run the cURL command below to generate your **Bearer Token**:

```bash
curl --location 'http://localhost:8080/realms/jungle-gaming/protocol/openid-connect/token' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'client_id=backend-client' \
--data-urlencode 'client_secret=my-secret-key' \
--data-urlencode 'grant_type=client_credentials'

```

5. **Test the Endpoints:** Step 5.
With the containerized environment up and your token in hand, start testing the API endpoints using your preferred HTTP client (Postman, Insomnia, cURL, etc.), passing the obtained token in the authorization header (`Authorization: Bearer <your-token>`).
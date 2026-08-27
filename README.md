# evap

API y páginas SSR en Go, empaquetadas como una única AWS Lambda detrás de
API Gateway HTTP API. Costo $0 en reposo: Lambda, API Gateway HTTP API y
DynamoDB On-Demand no cobran sin invocaciones/tráfico.

## Desarrollo local

```sh
cd backend
export DYNAMODB_TABLE_NAME=evap_users_local
export OAUTH_REDIRECT_BASE_URL=http://localhost:8080
export GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=...
export GITHUB_CLIENT_ID=... GITHUB_CLIENT_SECRET=...
export JWT_SIGNING_KEY=dev-secret
export LOCAL_HTTP_ADDR=:8080
go run ./cmd/api
```

Sin `OAUTH_SSM_PREFIX` definido, la configuración cae a variables de entorno
planas (útil para desarrollo). En AWS, `OAUTH_SSM_PREFIX` activa la lectura
de secretos desde SSM Parameter Store (SecureString).

## Tests

```sh
cd backend
make test              # unit tests
make test-integration  # requiere DynamoDB Local en localhost:8000
make lint               # golangci-lint
```

## Build y empaquetado para Lambda

```sh
cd backend
make package  # genera backend/dist/function.zip (runtime provided.al2023, arm64)
```

## Infraestructura

Ver [infra/](infra) para Terraform (Lambda, API Gateway HTTP API, DynamoDB,
IAM de mínimo privilegio, SSM y el rol OIDC para GitHub Actions). Los secretos
OAuth/JWT se cargan como placeholders (`CHANGEME`) y deben actualizarse
manualmente tras el primer `apply`:

```sh
aws ssm put-parameter --name /evap/prod/google_client_id --type SecureString --overwrite --value "..."
```

## CI/CD

`.github/workflows/deploy.yml`: tests + lint + `terraform plan` en PRs;
build, `terraform apply` y smoke test en push a `main`, usando OIDC
(`AWS_DEPLOY_ROLE_ARN` como secret del repo) sin llaves AWS de larga duración.


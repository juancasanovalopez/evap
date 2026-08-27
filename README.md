# evap

Aplicación web para estimar la evaporación de una superficie abierta usando
datos meteorológicos horarios. El backend está implementado en Go y sirve la
interfaz SSR, los recursos estáticos y la API desde una única función AWS
Lambda.

## Estado del proyecto

El proyecto está operativo como una aplicación web SSR con autenticación OAuth2
mediante Google y GitHub. La rama que dispara el despliegue automático es
`master`.

## Arquitectura

```text
Navegador
	 |
	 v
API Gateway HTTP API
	 |
	 v
AWS Lambda (Go, provided.al2023, arm64)
	 |             |              |
	 v             v              v
DynamoDB       SSM            Open-Meteo
usuarios       secretos       previsión horaria
```

- `backend/`: aplicación Go, router chi, handlers, autenticación, persistencia,
	plantillas y estáticos.
- `infra/`: Terraform para Lambda, API Gateway, DynamoDB, SSM, CloudWatch y
	el proveedor OIDC de GitHub Actions.
- `.github/workflows/deploy.yml`: pruebas, lint, empaquetado y despliegue.
- DynamoDB usa una tabla on-demand con claves `PK` y `SK` para perfiles OAuth.
- Los secretos de producción se leen desde SSM Parameter Store como
	`SecureString`; no deben almacenarse en Git ni en Terraform con valores reales.

## Requisitos

- Go `1.25.0` o compatible con [backend/go.mod](backend/go.mod).
- Git.
- Docker Desktop para las pruebas con DynamoDB Local.
- AWS CLI y Terraform para desplegar infraestructura.
- Credenciales OAuth de Google y/o GitHub para probar el login completo.

Los comandos `make` requieren GNU Make. También se pueden ejecutar directamente
los comandos Go descritos más abajo.

## Desarrollo local

### Configuración

Copia [backend/.env.example](backend/.env.example) a un archivo local. No
subas ese archivo al repositorio:

```sh
cp backend/.env.example backend/.env
```

Go no carga automáticamente archivos `.env`; exporta las variables en la
terminal antes de arrancar:

```sh
export DYNAMODB_TABLE_NAME=evap_users_local
export DYNAMODB_ENDPOINT=http://localhost:8000
export OAUTH_REDIRECT_BASE_URL=http://localhost:8080
export JWT_SIGNING_KEY=local-only-change-me
export LOCAL_HTTP_ADDR=:8080
export OAUTH_SSM_PREFIX=
export AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1
```

Deja `OAUTH_SSM_PREFIX` vacío en local. Si tiene valor, la aplicación intenta
leer secretos desde AWS SSM. `DYNAMODB_ENDPOINT` lo usan actualmente las
pruebas de integración; el entrypoint `cmd/api` todavía no lo conecta de forma
explícita al cliente DynamoDB, por lo que el arranque manual puede requerir
credenciales AWS y debe considerarse una mejora pendiente.

### DynamoDB Local

Con Docker iniciado:

```sh
docker run --rm --name evap-dynamodb \
	-p 8000:8000 amazon/dynamodb-local:latest `
	-jar DynamoDBLocal.jar -sharedDb -inMemory
```

La tabla de integración se crea automáticamente por los tests. Para una
sesión manual, crea una tabla con `PK` como clave hash y `SK` como clave range
usando AWS CLI contra el endpoint local.

### Arrancar la aplicación

```sh
cd backend
go run ./cmd/api
```

Abre `http://localhost:8080`. El endpoint de salud no requiere autenticación:

```sh
curl --fail http://localhost:8080/health
```

El login OAuth local requiere registrar las URLs callback locales en el
proveedor correspondiente. Actualmente el entrypoint local fija las cookies
como `Secure`, por lo que los navegadores pueden rechazarlas en HTTP. Para
probar OAuth local de extremo a extremo se necesita HTTPS local o hacer
configurable esa opción antes de usar este flujo.

Las pruebas de integración sí pueden aislarse con DynamoDB Local y credenciales
de prueba. El arranque manual no hace cambios en GitHub ni publica nada por sí
mismo, pero no debe considerarse completamente aislado de AWS hasta conectar
`DYNAMODB_ENDPOINT` en `cmd/api`.

## API y rutas

| Método | Ruta | Auth | Descripción |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Liveness para smoke tests y monitorización. |
| `GET` | `/` | Cookie JWT opcional | Dashboard SSR; usuarios anónimos van a `/login`. |
| `GET` | `/login` | No | Página de acceso. |
| `GET` | `/auth/google/login` | No | Inicia OAuth2 con Google. |
| `GET` | `/auth/github/login` | No | Inicia OAuth2 con GitHub. |
| `GET` | `/auth/{provider}/callback` | Cookie de estado | Completa OAuth2 y crea la sesión. |
| `GET` | `/api/v1/private` | JWT | Devuelve el perfil autenticado. |
| `GET` | `/api/v1/simulate` | JWT | Consulta Open-Meteo y calcula evaporación. |

La API acepta el JWT mediante `Authorization: Bearer <token>` o la cookie
HttpOnly `session`. La sesión dura 24 horas y el estado OAuth cinco minutos.

### Parámetros de simulación

`GET /api/v1/simulate` acepta estos parámetros:

| Parámetro | Ejemplo | Default actual |
| --- | --- | --- |
| `area` | `32` | `32.0` |
| `profundidad` | `1.2` | `1.2` |
| `lat` | `40.4167` | `40.4167` |
| `lon` | `-3.7037` | `-3.7037` |
| `fecha_inicio` | `2025-07-15` | `2025-07-15` |
| `fecha_fin` | `2025-07-17` | `2025-07-17` |

La respuesta incluye litros perdidos, milímetros descendidos, metros lineales
y el detalle horario. El cálculo depende de `https://api.open-meteo.com` y
tiene un timeout HTTP de 10 segundos.

## Configuración

| Variable | Local | AWS |
| --- | --- | --- |
| `DYNAMODB_TABLE_NAME` | Nombre de tabla local | Inyectada por Terraform |
| `DYNAMODB_ENDPOINT` | `http://localhost:8000` | Endpoint AWS por defecto |
| `OAUTH_REDIRECT_BASE_URL` | `http://localhost:8080` | URL de API Gateway |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Variables planas | SSM |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | Variables planas | SSM |
| `JWT_SIGNING_KEY` | Variable local | SSM |
| `OAUTH_SSM_PREFIX` | Vacía | Por defecto `/evap/prod` |
| `ALLOWED_CORS_ORIGINS` | Lista separada por comas | Variable Terraform |
| `LOCAL_HTTP_ADDR` | `:8080` | No se usa en Lambda |

Cuando `OAUTH_SSM_PREFIX` está definido, se consultan cinco parámetros bajo
`<prefix>`: los cuatro secretos OAuth y `jwt_signing_key`.

## Pruebas y calidad

Desde `backend/`:

```sh
go test ./...
go test -tags=integration ./test/integration/...
golangci-lint run ./...
```

Las pruebas de integración requieren DynamoDB Local en `localhost:8000` y
pueden recibir otro endpoint con `DYNAMODB_ENDPOINT`. Con GNU Make:

```sh
make test
make test-integration
make lint
```

No ejecutes las pruebas de integración contra una tabla de producción.

## Build y empaquetado

La función usa `provided.al2023` y arquitectura ARM64:

```sh
cd backend
make package
```

El artefacto se genera en `backend/dist/function.zip`. Sin Make, ejecuta desde
`backend/`:

```sh
mkdir -p dist
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/bootstrap ./cmd/api
cd dist && zip -j function.zip bootstrap
```

## Despliegue AWS con Terraform

1. Configura credenciales AWS con permisos suficientes para el bootstrap.
2. Copia `infra/terraform.tfvars.example` a `infra/terraform.tfvars`.
3. Sustituye `github_repository`, `oauth_redirect_base_url` y los orígenes
	 CORS. No guardes secretos OAuth en `terraform.tfvars`.
4. Valida la infraestructura:

	 ```sh
	 cd infra
	 terraform init
	 terraform fmt -check -recursive
	 terraform validate
	 terraform plan
	 ```

5. Construye el paquete y aplica:

	 ```sh
	 cd ../backend
	 make package
	 cd ../infra
	 terraform apply
	 ```

Terraform crea Lambda, API Gateway HTTP API, DynamoDB on-demand, CloudWatch,
parámetros SSM y el rol OIDC de GitHub Actions. Tras el primer `apply`,
rellena los cinco parámetros `SecureString` fuera de Terraform:

```sh
aws ssm put-parameter --name /evap/prod/google_client_id \
	--type SecureString --overwrite --value "..."
```

Obtén la URL pública con `terraform output -raw api_endpoint` y registra las
URLs callback definitivas en Google y GitHub. El estado Terraform es local por
defecto; [infra/backend.tf.example](infra/backend.tf.example) describe la
migración a S3 con locking.

## CI/CD

[.github/workflows/deploy.yml](.github/workflows/deploy.yml) se ejecuta con
cambios en `backend/`, `infra/` o el workflow:

- Pull requests: tests unitarios, integración con DynamoDB Local, lint y
	`terraform plan`.
- Push a `master`: tests, lint, empaquetado, `terraform apply` y smoke test de
	`/health`.
- AWS se autentica mediante GitHub OIDC, sin claves AWS de larga duración.

Requiere el secret `AWS_DEPLOY_ROLE_ARN` y la variable
`OAUTH_REDIRECT_BASE_URL` del repositorio. El rol está restringido a la rama
`master` y al repositorio configurado en Terraform.

## Seguridad y operación

- No subas `.env`, credenciales OAuth, tokens JWT ni valores reales de SSM.
- Usa HTTPS en producción; las cookies de sesión deben ser HttpOnly y Secure.
- Mantén `ALLOWED_CORS_ORIGINS` explícito; no uses `*` con cookies.
- API Gateway aplica throttling de 10 solicitudes por segundo con burst 20.
- El rate limiter interno es por IP y best-effort entre instancias Lambda.
- CloudWatch conserva logs durante 14 días por defecto.
- DynamoDB usa pago por solicitud y PITR está desactivado actualmente.

## Troubleshooting

### `go` o `make` no se reconoce

Instala Go 1.25 y GNU Make, o ejecuta directamente `go test`, `go build` y los
comandos Terraform equivalentes.

### DynamoDB Local no responde

Comprueba que Docker Desktop está iniciado y que el puerto `8000` está libre:

```sh
docker ps
curl --fail http://localhost:8000
```

### El login local vuelve a `/login`

Comprueba las URLs callback y la política de cookies del navegador. La cookie
`Secure` no se guarda en HTTP; usa HTTPS local o habilita una configuración de
cookies inseguras exclusivamente para desarrollo.

### Terraform no encuentra el paquete Lambda

Ejecuta `make package` desde `backend/` y confirma que existe
`backend/dist/function.zip`.

## Mejoras recomendadas

### Prioridad alta

- Hacer configurable `SecureCookies`, con `true` por defecto y una opción
	explícita para desarrollo local.
- Validar al arrancar las variables obligatorias y emitir mensajes claros
	cuando falten secretos, tabla o URL de callback.
- Validar rangos, coordenadas, fechas y límites de tamaño en `/simulate`.
- Revisar `X-Forwarded-For` y confiar únicamente en cabeceras del proxy
	conocido o en límites gestionados por API Gateway.

### Prioridad media

- Sustituir `apigateway:*` en el rol de despliegue por permisos concretos.
- Separar roles de Terraform para `plan` y `apply` si el equipo crece.
- Añadir métricas, alarmas CloudWatch, request IDs correlacionables y un
	dashboard operativo.
- Añadir tests de contrato para configuración, cookies, CORS y simulaciones.
- Activar PITR y backups de DynamoDB según los requisitos de recuperación.

### Prioridad baja

- Añadir un endpoint de readiness separado de `/health` para comprobar
	dependencias cuando sea necesario.
- Migrar el estado Terraform a S3 con locking para evitar colisiones.


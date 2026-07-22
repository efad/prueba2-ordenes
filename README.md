# API de Órdenes - Prueba Técnica 2

API GraphQL de gestión de órdenes de compra construida con Go, gqlgen, PostgreSQL y Clean Architecture.

## Requisitos

- Go 1.25+
- Docker y Docker Compose
- Make (opcional)

## Arranque rápido

```bash
cp .env.example .env
docker compose up --build -d
```

Servicios:

| URL | Descripción |
|-----|-------------|
| `http://localhost:8080/health` | Health check |
| `http://localhost:8080/` | GraphQL Playground |
| `http://localhost:8080/query` | Endpoint GraphQL |

### Nota para WSL (Windows)

Si Docker Engine corre dentro de WSL y `localhost:8080` no responde desde el navegador de Windows, usar la IP de WSL:

```bash
wsl hostname -I
# Abrir http://<primera-ip>:8080/
```

O habilitar en `%USERPROFILE%\.wslconfig`:

```ini
[wsl2]
localhostForwarding=true
networkingMode=mirrored
```

Luego ejecutar `wsl --shutdown`.

Al arrancar, la aplicación ejecuta migraciones goose y carga productos de prueba si la tabla está vacía.

## Variables de entorno

| Variable | Obligatoria | Default | Descripción |
|----------|-------------|---------|-------------|
| `DATABASE_URL` | Sí | - | Conexión PostgreSQL |
| `JWT_SECRET` | Sí | - | Secreto para firmar JWT |
| `JWT_EXPIRATION` | No | `24h` | Duración del token |
| `PORT` | No | `8080` | Puerto HTTP |
| `POSTGRES_USER` | No | `orders` | Usuario Postgres (compose) |
| `POSTGRES_PASSWORD` | No | `orders` | Password Postgres (compose) |
| `POSTGRES_DB` | No | `orders` | Base de datos (compose) |

Ver [.env.example](.env.example).

## Tests

```bash
# Unit tests
go test ./...

# Integración (requiere Postgres)
make postgres-up
make migrate-up
make test-integration
```

## Lint

```bash
go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./...
```

## Ejemplos GraphQL

### 1. Registro

```graphql
mutation {
  register(input: { email: "user@example.com", password: "password1" }) {
    token
  }
}
```

### 2. Login

```graphql
mutation {
  login(input: { email: "user@example.com", password: "password1" }) {
    token
  }
}
```

### 3. Productos (requiere token)

En Playground, agregar HTTP HEADERS:

```json
{ "Authorization": "Bearer TU_TOKEN" }
```

```graphql
query {
  products(page: 1, pageSize: 10, filter: { name: "Teclado", minPrice: 10 }) {
    totalCount
    items { id name price stock }
  }
}
```

### 4. Crear orden

```graphql
mutation {
  createOrder(input: {
    items: [{ productId: "UUID_DEL_PRODUCTO", quantity: 1 }]
  }) {
    id
    total
    status
    items { quantity unitPrice product { name } }
  }
}
```

### 5. Mis órdenes y cancelar

```graphql
query {
  myOrders {
    items { id total status items { product { name } } }
  }
}

mutation {
  cancelOrder(id: "UUID_ORDEN") {
    id
    status
  }
}
```




## Stack

- Go + gqlgen
- PostgreSQL (pgx)
- JWT (HS256)
- goose (migraciones)
- DataLoader (N+1)
- docker-compose

## Decisión de diseño: transacciones

Las operaciones que modifican stock y órdenes se ejecutan dentro de un `TransactionManager` definido en dominio, que abre una transacción PostgreSQL y la propaga por `context` a los repositorios. Así, `createOrder` valida stock, descuenta inventario y persiste la orden de forma atómica: ante cualquier fallo se hace rollback completo. El descuento de stock usa `UPDATE ... WHERE stock >= cantidad` para evitar sobreventa en concurrencia. La cancelación bloquea la orden con `SELECT ... FOR UPDATE`, restaura el stock de cada ítem y marca la orden como cancelada (soft delete).

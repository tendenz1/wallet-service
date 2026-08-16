# Wallet Service

REST-сервис на Go для работы с балансом кошелька.

Стек: Go, PostgreSQL, pgx, Viper, Docker Compose.

## Запуск

```bash
docker compose up --build
```

Тестовый кошелёк:

```text
00000000-0000-0000-0000-000000000001
```

Начальный баланс:

```text
100000
```

## API

Изменение баланса:

```http
POST /api/v1/wallet
```

```json
{
  "walletId": "00000000-0000-0000-0000-000000000001",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

Получение баланса:

```http
GET /api/v1/wallets/{WALLET_UUID}
```

Поддерживаемые операции:

```text
DEPOSIT
WITHDRAW
```

## Тесты

```bash
go test ./...
```

Integration-тесты:

```powershell
$env:TEST_DATABASE_URL = 'postgres://wallet:wallet@localhost:15432/wallet?sslmode=disable'
go test -tags=integration ./internal/repository -v
```

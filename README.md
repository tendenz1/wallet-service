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

## Конкурентность

Атомарный `UPDATE` в PostgreSQL исключает гонку read-modify-write и потерю обновлений.
При `WITHDRAW` достаточность средств проверяется в том же `UPDATE`.
Поэтому баланс кошелька не может стать отрицательным.

## Тесты

Unit-тесты запускаются командой:

```bash
go test ./...
```

Integration-тесты проверяют работу с реальным PostgreSQL, включая конкурентные `Deposit` и `Withdraw`, и требуют запущенного Docker Compose:

```powershell
$env:TEST_DATABASE_URL = 'postgres://wallet:wallet@localhost:15432/wallet?sslmode=disable'
go test -tags=integration ./internal/repository -v
```

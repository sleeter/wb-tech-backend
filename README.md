# WB-tech задание L0

## Описание
В данном репозитории представлен сервис, который получает данные о заказе с помощью брокера сообщений, записывает их в базу даннных и кэш. Сервис позволяет получить данные о заказе по его идентификатору.

## Решение
Сервис написан на языке Golang с помощью gin-gonic, pgx, migrate, viper, squirrel, nats-io, а также базовых библиотек.
Для хранения используется база данных PostgreSQL, для удобного хранения модель была разбита на несколько таблиц: deliveries, payments, items, orders.

## Деплой
Для запуска проекта необходимо выполнить команду:

```docker-compose up```

Для отправки сообщения через publisher необходимо выполнить команду:

```go run cmd/pub/main.go```

Чтобы остановить сервис небходимо выполнить команду:

```docker-compose down```

## Пример запроса
```localhost:8080/order?order_uid=b563feb7b2b84b6test```
![GET_order_example](https://github.com/sleeter/wb-tech-backend/raw/master/pic/GET_order_example.png)

```localhost:8080/orders```
![GET_orders_example](https://github.com/sleeter/wb-tech-backend/pic/GET_orders_example.png)
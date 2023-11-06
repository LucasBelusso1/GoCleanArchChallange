# GoCleanArchChallange

Challange from the post graduation GoExpert.

To start the project (with docker and docker compose installed), run these commands:

```SHELL
$ docker compose up -d #Setup docker containers with Mysql and RabbitMQ
```
```SHELL
$ make migrate # Run migrations to create necessary tables
```
```SHELL
$ cd cmd/ordersystem # Go to main directory of the project
```
```SHELL
$ go run main.go wire_gen.go # Run the project
```


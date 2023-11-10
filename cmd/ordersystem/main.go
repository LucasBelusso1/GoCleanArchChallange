package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/LucasBelusso1/GoCleanArchChallange/configs"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/event/handler"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/infra/database"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/infra/graph"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/infra/grpc/pb"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/infra/grpc/service"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/infra/web/webserver"
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/usecase"
	"github.com/LucasBelusso1/GoCleanArchChallange/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	conf, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(conf.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", conf.DBUser, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rabbitMQChannel := getRabbitMQChannel(conf.RabbitMqUser, conf.RabbitMqPassword, conf.RabbitMqHost, conf.RabbitMqPort)

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)

	//TODO: Alterar para utilizar DI.
	orderRepository := database.NewOrderRepository(db)
	listOrderUseCase := usecase.NewListOrderUseCase(orderRepository)

	webserver := webserver.NewWebServer(conf.WebServerPort)
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	webserver.AddHandler(webserver.NewHttpHandler("post", "/order", webOrderHandler.Create))
	webserver.AddHandler(webserver.NewHttpHandler("get", "/orders", webOrderHandler.List))
	fmt.Println("Starting web server on port", conf.WebServerPort)
	go webserver.Start()

	grpcServer := grpc.NewServer()
	createOrderService := service.NewOrderService(*createOrderUseCase, *listOrderUseCase)
	pb.RegisterOrderServiceServer(grpcServer, createOrderService)
	reflection.Register(grpcServer)

	fmt.Println("Starting gRPC server on port", conf.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", conf.GRPCServerPort))
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(lis)

	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", conf.GraphQLServerPort)
	http.ListenAndServe(":"+conf.GraphQLServerPort, nil)
}

func getRabbitMQChannel(user, passwd, host, port string) *amqp.Channel {
	ampqUri := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, passwd, host, port)
	conn, err := amqp.Dial(ampqUri)
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}

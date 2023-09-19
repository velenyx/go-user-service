package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"some/internal/config"
	"some/internal/user"
	"some/internal/user/db"
	"some/pkg/client/mongodb"
	"some/pkg/logging"
	"time"
)

func main() {
	logger := logging.GetLogger()
	logger.Info("Create router")
	router := httprouter.New()

	cfg := config.GetConfig()

	cfgMongo := cfg.MongoDB
	mongoDBClient, err := mongodb.NewClient(context.Background(), cfgMongo.URI, cfgMongo.Username, cfgMongo.Password)
	if err != nil {
		panic(err)
	}

	databaseName := "feedback"
	database := mongoDBClient.Database(databaseName)
	storage := db.NewStorage(database, cfgMongo.Collection, logger)

	users, err := storage.FindAll(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(users)

	user1 := user.User{
		ID:           "",
		Email:        "myemail@example.com",
		Username:     "velenyx",
		PasswordHash: "asd90uu21083ui",
	}
	user1ID, err := storage.Create(context.Background(), &user1)
	if err != nil {
		panic(err)
	}
	logger.Info(user1ID)

	user2 := user.User{
		ID:           "",
		Email:        "moreMail@example.com",
		Username:     "Artur",
		PasswordHash: "SN)!*W#JIWEHJ*)WDIGHshd0",
	}
	user2ID, err := storage.Create(context.Background(), &user2)
	if err != nil {
		panic(err)
	}
	logger.Info(user2ID)

	user2Found, err := storage.FindOne(context.Background(), user2ID)
	if err != nil {
		panic(err)
	}
	fmt.Println(user2Found)

	user2Found.Email = "moreMail123123@here.com"
	err = storage.Update(context.Background(), user2Found)
	if err != nil {
		panic(err)
	}

	err = storage.Delete(context.Background(), user2ID)
	if err != nil {
		panic(err)
	}

	_, err = storage.FindOne(context.Background(), user2ID)
	if err != nil {
		panic(err)
	}

	logger.Info("Register user handler")
	handler := user.NewHandler(logger)
	handler.Register(router)

	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger := logging.GetLogger()
	logger.Info("start application")

	var listener net.Listener
	var listenErr error

	if cfg.Listen.Type == "sock" {
		logger.Info("detect app path")
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("create socket")
		socketPath := path.Join(appDir, "app.sock")

		logger.Info("listen unix socket")
		listener, listenErr = net.Listen("unix", socketPath)
		logger.Infof("server is listening unix socket: %s", socketPath)
	} else {
		logger.Info("listen tcp")
		listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port))
		logger.Infof("start server on %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)
	}

	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	server := &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}

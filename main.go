package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/controller"
	events "github.com/ipaas-org/ipaas-backend/handlers/containerEvents"
	"github.com/ipaas-org/ipaas-backend/handlers/httpserver"
	"github.com/ipaas-org/ipaas-backend/handlers/rabbitmq"
	"github.com/ipaas-org/ipaas-backend/pkg/logger"
	"github.com/ipaas-org/ipaas-backend/repo/mock"
	mongoRepo "github.com/ipaas-org/ipaas-backend/repo/mongo"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	gracefulShutdownTimeout = 15 * time.Second
)

const (
	StartHTTPHandler int = iota + 1
	StartRMQHandler
	StartDatabaseCleanupHandler
	StartContainerEventHandler
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}

	l := logger.NewLogger(conf.Log.Level, conf.Log.Type)
	l.Debug("initizalized logger")

	ctx, cancel := context.WithCancel(context.Background())

	c := controller.NewController(ctx, conf, l)
	l.Debugf("conf: %+v\n", conf)

	switch conf.Database.Driver {
	case "mock":
		l.Info("using mock database")
		c.UserRepo = mock.NewUserRepoer()
		c.TokenRepo = mock.NewTokenRepoer()
		c.StateRepo = mock.NewStateRepoer()
		c.ApplicationRepo = mock.NewApplicationRepoer()

	case "mongo":
		l.Info("using mongo database")

		l.Debug("connecting to database")
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Database.URI))
		if err != nil {
			l.Fatalf("main - mongo.Connect - error connecting to database: %s", err.Error())
		}
		cancel()

		l.Debug("connecting to user collection")
		userCollection := client.Database("ipaas").Collection("user")
		userRepo := mongoRepo.NewUserRepoer(userCollection)
		c.UserRepo = userRepo

		l.Debug("connecting to token collection")
		tokenCollection := client.Database("ipaas").Collection("token")
		tokenRepo := mongoRepo.NewTokenRepoer(tokenCollection)
		c.TokenRepo = tokenRepo

		l.Debug("connecting to state collection")
		stateCollection := client.Database("ipaas").Collection("state")
		stateRepo := mongoRepo.NewStateRepoer(stateCollection)
		c.StateRepo = stateRepo

		l.Debug("connecting to application collection")
		applicationCollection := client.Database("ipaas").Collection("application")
		applicationRepo := mongoRepo.NewApplicationRepoer(applicationCollection)
		c.ApplicationRepo = applicationRepo

		l.Debug("connecting to template collection")
		templateCollection := client.Database("ipaas").Collection("templates")
		templateRepo := mongoRepo.NewTemplateRepoer(templateCollection)
		c.TemplateRepo = templateRepo
	default:
		l.Fatalf("main - unknown database driver: %s", conf.Database.Driver)
	}

	e := echo.New()
	httpHandler := httpserver.InitRouter(e, l, c, conf)

	rmq := rabbitmq.NewRabbitMQ(conf.RMQ.URI, conf.RMQ.RequestQueue, conf.RMQ.ResponseQueue, c, l)
	if err := rmq.Connect(); err != nil {
		l.Fatalf("main - rmq.Connect - error connecting to rabbitmq: %s", err.Error())
	}
	rmq.Close()

	containerEventHandler, err := events.NewContainerEventHandler(ctx, c, l)
	if err != nil {
		l.Fatalf("main - events.NewContainerEventHandler - error creating container event handler: %s", err.Error())
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGTERM)
	RoutineMonitor := make(chan int, 100)
	RoutineMonitor <- StartHTTPHandler
	RoutineMonitor <- StartRMQHandler
	RoutineMonitor <- StartContainerEventHandler

	for {
		select {
		case i := <-interrupt:
			l.Info("main - signal: " + i.String())
			l.Info("main - canceling context")
			cancel()
			gracefulTimer := time.Tick(gracefulShutdownTimeout)
			//TODO: this is wrong, it needs to check both rmq and httphandler are done
			select {
			case <-gracefulTimer:
				l.Info("main - graceful shutdown timeout reached, exiting with status 1")
				os.Exit(1)
			case <-rmq.Done:
				l.Info("main - rabbitmq finished")
			case <-httpHandler.Done:
				l.Info("main - http handler finished")
			case <-containerEventHandler.Done:
				l.Info("main - container event handler finished")
			}
			//returns 0 because the shutdown was successful
			os.Exit(0)
		case err = <-rmq.Error:
			l.Errorf("rabbitmq handler error: %v", err)
		default:
		}

		select {
		case ID := <-RoutineMonitor:
			l.Infof("Starting Routine: %d", ID)
			switch ID {
			case StartRMQHandler:
				go rmq.Start(ctx, StartRMQHandler, RoutineMonitor)
			case StartHTTPHandler:
				go httpserver.StartRouter(ctx, httpHandler, conf, StartHTTPHandler, RoutineMonitor)
			case StartContainerEventHandler:
				go events.StartConainerEventHandler(ctx, containerEventHandler, StartContainerEventHandler, RoutineMonitor)
			default:
			}
		default:
		}

		time.Sleep(10 * time.Millisecond)
	}
}

/*
!PAGES ENDPOINTS:
*endpoints for public pages:
/static -> static files
/ -> homepage
/login -> login page
/{studentID} -> public student page
/{studentID}/{appID} -> public app page with info of it

*endpoints for student pages:
/user/ -> user area
/user/application/new -> page to create a new application
/user/database/new -> page to create a new database
! not implemented as pages  /user/application/new -> new application page
! not implemented as pages  /user/database/new -> new database page

!API ENDPOINTS:
*public api endpoints:
/api/oauth -> oauth endpoint (generate the oauth url)
/api/logout -> destroy the session and redirets to login
/api/tokens/new -> get a new token pair from a refresh token
/api/{studentID}/all -> get all the public applications of a student
/api/{studentID}/{appID} -> get the info of a public application

*user api endpoints:
/api/user/ -> get the info of the user
/api/user/getApps/{type} -> get all the applications of a user (private or public) with the type of application (database, web, all, updatable)

*container api endpoints:
/api/container/delete/{containerID} -> delete a container
/api/container/publish/{containerID} -> publish a container
/api/container/revoke/{containerID} -> revoke a container

*api endpoints for database:
/api/db/new -> create a new database
! not implemented /api/db/export/{containerID}/{dbName} -> export a database

*api endpoints for applications:
/api/app/new -> create a new application
/api/app/update/{containerID} -> update an application if the repo is changed

func main() {
	mainRouter := mux.NewRouter()
	mainRouter.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("static/"))))

	mainRouter.HandleFunc("/", handler.UserPageHandler)
	mainRouter.HandleFunc("/login", handler.LoginPageHandler)
	mainRouter.HandleFunc("/mock", handler.MockRegisterUserPageHandler)
	mainRouter.HandleFunc("/{studentID}", handler.PublicStudentPageHandler)
	// mainRouter.HandleFunc("/{studentID}/{appID}", handler.PublicAppPageHandler)
	//homepage for the logged user
	mainRouter.HandleFunc("/user/", handler.UserPageHandler).Methods("GET")

	//user's router with access token middleware
	userAreaRouter := mainRouter.PathPrefix("/user").Subrouter()
	//set middleware on user area router
	userAreaRouter.Use(handler.TokensMiddleware)
	//page to create a new application
	userAreaRouter.HandleFunc("/application/new", handler.NewAppPageHandler).Methods("GET")
	//page to create a new database
	userAreaRouter.HandleFunc("/database/new", handler.NewDatabasePageHandler).Methods("GET")

	//!API HANDLERS
	api := mainRouter.PathPrefix("/api").Subrouter()

	//! PUBLIC HANDLERS
	api.HandleFunc("/mock/create", handler.MockRegisterUserHandler).Methods("POST")
	api.HandleFunc("/oauth", handler.OauthHandler).Methods("GET")
	api.HandleFunc("/logout", handler.LogoutHandler).Methods("GET")
	api.HandleFunc("/oauth/check/{randomID}", handler.CheckOauthState).Methods("GET")
	api.HandleFunc("/tokens/new", handler.NewTokenPairFromRefreshTokenHandler).Methods("GET")
	api.HandleFunc("/{studentID}/all", handler.GetAllApplicationsOfStudentPublic).Methods("GET")
	// api.HandleFunc("/{studentID}/{appID}", handler.GetInfoApplication).Methods("GET")

	//! USER HANDLERS
	userApiRouter := api.PathPrefix("/user").Subrouter()
	userApiRouter.Use(handler.TokensMiddleware)
	//get the user data
	//still kinda don't know what to do with this one, will probably return the homepage
	userApiRouter.HandleFunc("/", handler.LoginHandler).Methods("GET")
	//validate a GitHub repo and (if valid) returns the branches
	userApiRouter.HandleFunc("/validate", handler.ValidGithubUrlAndGetBranchesHandler).Methods("POST")
	//get all the applications (even the private one) must define the type (database, web, all)
	userApiRouter.HandleFunc("/getApps/{type}", handler.GetAllApplicationsOfStudentPrivate).Methods("GET")
	//update an application
	userApiRouter.HandleFunc("/application/update/{containerID}", handler.UpdateApplicationHandler).Methods("GET")

	//! CONTAINER HANDLERS
	containerApiRouter := api.PathPrefix("/container").Subrouter()
	containerApiRouter.Use(handler.TokensMiddleware)
	//delete a container
	containerApiRouter.HandleFunc("/delete/{containerID}", handler.DeleteApplicationHandler).Methods("DELETE")
	//publish a container
	containerApiRouter.HandleFunc("/publish/{containerID}", handler.PublishApplicationHandler).Methods("GET")
	//revoke a container
	containerApiRouter.HandleFunc("/revoke/{containerID}", handler.RevokeApplicationHandler).Methods("GET")

	//! DBaaS HANDLERS
	//DBaaS router (subrouter of user area router so it has access token middleware)
	dbApiRouter := api.PathPrefix("/db").Subrouter()
	dbApiRouter.Use(handler.TokensMiddleware)
	//let the user create a new database
	dbApiRouter.HandleFunc("/new", handler.NewDBHandler).Methods("POST")
	//export a database
	//dbApiRouter.HandleFunc("/export/{containerID}/{dbName}")

	//! APPLICATIONS HANDLERS
	//application router, it's the main part of the application
	appApiRouter := api.PathPrefix("/app").Subrouter()
	appApiRouter.Use(handler.TokensMiddleware)
	appApiRouter.HandleFunc("/new", handler.NewApplicationHandler).Methods("POST")
	appApiRouter.HandleFunc("/update/{containerID}", handler.UpdateApplicationHandler).Methods("POST")

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST"})

	//start event handler
	log.Println("starting event handler")
	go handler.cc.EventHandler()

	go RunExecutor(
		oauthStateCleaningInterval,
		refreshTokenCleaningInterval,
		usersCleaningInterval,
		pollingIDsCleaningInterval,
		executeCleaning,
	)

	go CheckExpiries(executeCleaning)

	log.Println("starting the server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(mainRouter)))
}
*/

package server

import (
	"avanoo_cd/deploy"
	"avanoo_cd/utils"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

type AppServer struct {
	httpServer   *http.Server
	addr         string
	ErrorChannel chan bool
}

func (app *AppServer) InitServer() {
	router := mux.NewRouter()
	app.httpServer = &http.Server{
		Addr:    app.addr,
		Handler: router,
	}

	basicHandler := alice.New(utils.LoggingHandler)
	commonHandlers := alice.New(utils.LoggingHandler, utils.RecoverHandler)

	router.MethodNotAllowedHandler = basicHandler.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	router.NotFoundHandler = basicHandler.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	router.Handle("/domain", commonHandlers.ThenFunc(deploy.ManageDomain)).Methods("POST", "OPTIONS")
	router.Handle("/updateDomainBranch", commonHandlers.ThenFunc(deploy.UpdateDomainBranch)).Methods("POST", "OPTIONS")
	router.Handle("/domains", commonHandlers.ThenFunc(deploy.ListDomains)).Methods("GET", "OPTIONS")
	router.Handle("/domain/{domain}", commonHandlers.ThenFunc(deploy.DetailDomain)).Methods("GET", "OPTIONS")
	router.Handle("/domain/{domain}", commonHandlers.ThenFunc(deploy.DeleteDomain)).Methods("DELETE", "OPTIONS")
	router.Handle("/webhook", commonHandlers.ThenFunc(deploy.ListenWebHookEvent)).Methods("POST", "OPTIONS")
	router.Handle("/builds", commonHandlers.ThenFunc(deploy.ListBuilds)).Methods("GET", "OPTIONS")
	//router.Handle("/docs", httpSwagger.WrapHandler).Methods("GET", "OPTIONS")

	if app.httpServer != nil {
		go func() {

			err := app.httpServer.ListenAndServe()
			log.Printf("HTTP Server Listen and Serve TLS:%v\n", err.Error())
			app.ErrorChannel <- true
		}()
	}
}

func (app *AppServer) CloseApplication() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(ctx); err != nil {
			log.Printf("Main Shutdown:%v\n", err.Error())
		}
	}
}

func CreateAppServer(addr string) (*AppServer, error) {
	appServer := AppServer{}
	appServer.addr = addr
	return &appServer, nil
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/goava/di"
)

// StdLogger is a standard logger with implementation via log.
type StdLogger struct {
}

// NewStdLogger creates std logger.
func NewStdLogger() *StdLogger {
	return &StdLogger{}
}

func (l *StdLogger) Logf(format string, values ...interface{}) {
	log.Printf(format, values...)
}

func main() {
	var ctx context.Context
	c, err := di.New(
		di.Provide(NewStdLogger, di.As(new(di.Logger))),
		di.Provide(NewContext),  // provide application context
		di.Provide(NewServer),   // provide http server
		di.Provide(NewServeMux), // provide http serve mux
		// controllers
		di.Provide(NewOrderController, di.As(new(Controller))), // provide order controller
		di.Provide(NewUserController, di.As(new(Controller))),  // provide user controller
		// invokes
		di.Invoke(StartServer),
		// resolves
		di.Resolve(&ctx),
	)
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
	c.Cleanup()
}

// StartServer starts http server.
func StartServer(server *http.Server) {
	go func() {
		log.Println("start listen")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("listen error:", err)
		}
	}()
}

// NewContext creates new application context.
func NewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
		<-stop
		cancel()
	}()
	return ctx
}

// NewServer creates a http server with provided mux as handler.
func NewServer(mux *http.ServeMux) (*http.Server, func()) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	cleanup := func() {
		if err := server.Close(); err != nil {
			log.Println("server close error:", err)
		}
		log.Println("server closed")
	}
	return server, cleanup
}

// NewServeMux creates a new http serve mux.
func NewServeMux(controllers []Controller) *http.ServeMux {
	mux := &http.ServeMux{}
	for _, controller := range controllers {
		controller.RegisterRoutes(mux)
	}
	return mux
}

// Controller is an interface that can register its routes.
type Controller interface {
	RegisterRoutes(mux *http.ServeMux)
}

// OrderController is a http controller for orders.
type OrderController struct{}

// NewOrderController creates a auth http controller.
func NewOrderController() *OrderController {
	return &OrderController{}
}

// RegisterRoutes is a Controller interface implementation.
func (a *OrderController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/orders", a.RetrieveOrders)
}

// Retrieve loads orders and writes it to the writer.
func (a *OrderController) RetrieveOrders(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("Orders"))
}

// UserController is a http endpoint for a user.
type UserController struct{}

// NewUserController creates a user http endpoint.
func NewUserController() *UserController {
	return &UserController{}
}

// RegisterRoutes is a Controller interface implementation.
func (e *UserController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users", e.RetrieveUsers)
}

// Retrieve loads users and writes it using the writer.
func (e *UserController) RetrieveUsers(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("Users"))
}

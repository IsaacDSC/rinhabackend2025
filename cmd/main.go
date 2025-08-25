package main

import (
	"context"
	"database/sql"
	"github.com/IsaacDSC/rinhabackend2025/internal/appcfg"
	"github.com/IsaacDSC/rinhabackend2025/internal/payhealth"
	"github.com/IsaacDSC/rinhabackend2025/internal/payprocess"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystate"
	"github.com/IsaacDSC/rinhabackend2025/internal/paystore"
	"github.com/IsaacDSC/rinhabackend2025/internal/rpay"
	"github.com/IsaacDSC/rinhabackend2025/internal/wpay"
	"github.com/IsaacDSC/rinhabackend2025/pkg/handle"
	"github.com/IsaacDSC/rinhabackend2025/pkg/middleware"
	"github.com/IsaacDSC/workqueue"
	"github.com/IsaacDSC/workqueue/SDK"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	ctx := context.Background()
	env := appcfg.Get()

	redisClient := redis.NewClient(&redis.Options{
		//Addr: "redis:6379",
		Addr: env.RedisUrl,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	//"idsc:admin@tcp(mysql:3306)/rinhabackend"
	db, err := sql.Open("mysql", env.DatabaseUrl)
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	opts := workqueue.NewOptsBuilder().
		WithQueueType("internal.critical").
		WithMaxRetries(5).
		WithRetention(workqueue.NewDuration("168h")). // 7 days
		WithUniqueTTL(workqueue.NewDuration("1m")).
		Build()

	//"payment-processor-default:8080"
	//"payment-processor-fallback:8080"
	dp := payprocess.NewPaymentProcessor("default", &url.URL{Scheme: "http", Host: env.ProcessorDefaultUrl})
	fp := payprocess.NewPaymentProcessor("fallback", &url.URL{Scheme: "http", Host: env.ProcessorFallbackUrl})

	state := paystate.NewState(redisClient, dp, fp)
	go payhealth.StartJob(ctx, state, dp, fp)

	store := paystore.NewMySQLStore(db)
	//"http://localhost:8080"
	producer := SDK.NewProducer(env.GQueueUrl, "your-token", opts)

	mux := http.NewServeMux()
	handlers := []handle.HandleHTTP{
		rpay.GetHandleHTTP(store),
		wpay.CmdPaymentProcessor(store, wpay.NewCmdPaymentProcessor(producer)),
		wpay.EventPaymentReceived(state, wpay.NewEventPaymentReceived(producer)),
		wpay.EventPaymentProcessed(store, wpay.NewEventPaymentProcessed(producer)),
	}

	for _, h := range handlers {
		mux.HandleFunc(h.Path, h.Handle)
	}

	loggedMux := middleware.LoggingMiddleware(mux)
	recoveryMux := middleware.RecoveryMiddleware(loggedMux)

	log.Print("Listening on port 3333")
	if err := http.ListenAndServe(":3333", recoveryMux); err != nil {
		log.Fatal(err)
	}

}

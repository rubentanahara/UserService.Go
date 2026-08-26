package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rubentanahara/user_service/internal/logging"
	"github.com/rubentanahara/user_service/internal/ratelimit"
)

const (
	maxBodyBytes       = 1 << 20
	requestTimeout     = 5 * time.Second
	shutdownTimeout    = 10 * time.Second
	readHeaderTimeout  = 5 * time.Second
	readTimeout        = 10 * time.Second
	writeTimeout       = 10 * time.Second
	idleTimeout        = 60 * time.Second
	loginRatePerSecond = 5
	loginRateBurst     = 10
	defaultMaxPoolSize = 100
)

func main() {
	ctx := context.Background()

	srv, cleanup, err := InitializeServer(ctx)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}
	defer cleanup()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(bodySizeLimit(maxBodyBytes))
	r.Use(logging.Middleware(srv.Logger))
	r.Use(timeout(requestTimeout))

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		if err := srv.Mongo.Ping(c.Request.Context(), nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv.Handler.Register(r, ratelimit.PerIP(loginRatePerSecond, loginRateBurst))

	httpServer := &http.Server{
		Addr:              ":" + string(srv.Port),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func bodySizeLimit(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}

func timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

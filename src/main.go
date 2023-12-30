package main

import (
	"context"
	"fmt"
	"ginx/logx"
	"ginx/middleware"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func getHostName() string {
	cmd := exec.Command("uname", "-n") // Linux或MacOS上的命令为 "uname -n"
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(output)
}

func main() {
	hostName := getHostName()
	logx.InitFile("gin_logrus/logs", "server")
	router := gin.New()
	router.Use(middleware.LogMiddleware())
	router.GET("/testFuncPanic", func(ctx *gin.Context) {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					var error = fmt.Sprintf("error: %v\n", err)
					logrus.Error(error)
				}
			}()
			panic("error")
		}()
	})
	router.GET("/testPanic", func(ctx *gin.Context) {
		var message = fmt.Sprintf("%s Application is about to groove panic", hostName)
		ctx.JSON(200, gin.H{"message": message})
		logrus.Panic(message)
	})
	router.GET("/testExited", func(ctx *gin.Context) {
		var message = fmt.Sprintf("%s Application is about to exited", hostName)
		ctx.JSON(200, gin.H{"message": message})
		logrus.Fatal(message)
	})
	router.GET("/hello", func(ctx *gin.Context) {
		var message = fmt.Sprintf("Hello %s !", hostName)
		logrus.Infoln(message)
		ctx.JSON(200, gin.H{"message": message})
	})

	server := &http.Server{
		Addr:    ":8857",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			var message = fmt.Sprintf("listen: %s\n", err)
			logrus.Panic(message)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Infoln("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Shutdown Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		logrus.Infoln("Timeout Of 1 Seconds.")
	}
	logrus.Infoln("Server exiting")
}

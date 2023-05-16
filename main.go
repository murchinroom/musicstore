package main

import (
	"context"
	"flag"
	"musicstore/audiofilestore"
	"musicstore/metadata"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cdfmlr/crud/config"
	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/router"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var logger = log.ZoneLogger("musicstore")

var (
	configFile = flag.String("config", "config.yaml", "config file path")
	dryRun     = flag.Bool("dry-run", false, "print config and exit")
	corsEnable = flag.Bool("cors", false, "enable cors")
)

func main() {
	flag.Parse()
	cfg := loadConfig(*configFile)
	srv := startServices(cfg)
	gracefulShoutdown(srv)
}

func loadConfig(configFile string) *MusicstoreConfig {
	var cfg MusicstoreConfig
	config.Init(&cfg, config.FromFile(configFile))

	logger.Info("config loaded.")
	cfg.Write(os.Stdout)

	if *dryRun {
		os.Exit(0)
	}

	return &cfg
}

func startServices(cfg *MusicstoreConfig) *http.Server {
	logger.Info("starting musicstore...")

	r := router.NewRouter()

	// CORS here is not needed, murecom-gw4reader now proxies audio files requests.
	// duplicate CORS headers will cause problems.
	if *corsEnable {
		logger.Info("CORS is enabled.")
		r.Use(cors.Default())
	}

	// so, the odd thing here is that, we ListenAndServe first,
	// and then register routes (by metadata.Start & startAudioFileStore).
	//
	// this is because, to LoadFromDir (a.k.a. AddTracksFromDir) in
	// startAudioFileStore(), we need expose the uri to audio files,
	// so that emomusic can download and analyze them.
	srv := startHttpServer(cfg.HttpListenAddr, r)

	if cfg.Emomusic.Server != "" {
		os.Setenv("EMOMUSIC_SERVER", cfg.Emomusic.Server)
	}

	metadata.Start(cfg.Metadata.DB, r)

	for _, afsCfg := range cfg.AudioFileStores {
		if err := startAudioFileStore(afsCfg, r); err != nil {
			logger.Fatalf("startAudioFileStore failed: %v", err)
		}
	}

	return srv
}

func corsSetting(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func startHttpServer(addr string, r http.Handler) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	logger.Infof("server started at %s", addr)

	return srv
}

func startAudioFileStore(afsCfg AudioFileStoreConfig, r gin.IRouter) error {
	afs := audiofilestore.NewAudioFileStore(
		afsCfg.Name, afsCfg.FileDir, afsCfg.BaseUrl, afsCfg.EnableEmomusic, r)

	if afsCfg.LoadFromDir {
		if err := afs.AddTracksFromDir(); err != nil {
			return err
		}
	}
	return nil
}

func gracefulShoutdown(srv *http.Server) {
	// https://gin-gonic.com/docs/examples/graceful-restart-or-stop/

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 200 ms.
	select {
	case <-ctx.Done():
		logger.Println("timeout of 200 ms.")
	}
	logger.Println("Server exiting")
}

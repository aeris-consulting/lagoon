package main

import (
	"context"
	"encoding/base64"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"lagoon/api"
	"lagoon/datasource"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "lagoon/datasource/redis"
)

const contextPath = "/lagoon"

var rootCmd = &cobra.Command{
	Use:   "lagoon",
	Short: "Lagoon is a GUI to visualize data from various middleware",
	Long: `A friendly GUI to explore and work on data
                provided by Redis and others.
                Complete documentation is available at https://github.com/ericjesse/lagoon`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			yamlConfig []byte
			err        error
		)

		if *configurationFlags.base64 != "" {
			yamlConfig, err = base64.StdEncoding.DecodeString(*(configurationFlags.base64))
		} else if *configurationFlags.file != "" {
			if _, err := os.Stat(*configurationFlags.file); err == nil {
				yamlConfig, err = ioutil.ReadFile(*(configurationFlags.file))
			}
		}

		if len(yamlConfig) > 0 {
			err = yaml.Unmarshal(yamlConfig, &configuration)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			for _, ds := range configuration.Datasources {
				log.Printf("Creating data source %s", ds.Name)
				dsInfos, _ := api.CreateDataSourceFromDescriptor(ds, true)
				api.DataSourcesHeaders[dsInfos.Id] = dsInfos
			}
		}

		// Listen and Server in 0.0.0.0:port
		if configuration.Port == 0 {
			configuration.Port = 4000
		}

		router := setupRouter()

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(configuration.Port),
			Handler: router,
		}

		go func() {
			// service connections
			log.Printf("Starting Lagoon listening %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server start failed: %s\n", err)
			}
		}()

		// Waiting for a signal to stop the server.
		quit := make(chan os.Signal)
		// SIGKILL but can't be catch and is therefore ignored.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")
		api.CloseAllDataSources()
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %s\n", err)
		}
		// Waiting for the timeout.
		select {
		case <-ctx.Done():
		}
	},
}

var (
	configurationFlags struct {
		base64 *string
		file   *string
	}

	configuration struct {
		Port        int                               `yaml:"port"`
		Datasources []datasource.DataSourceDescriptor `yaml:"datasources"`
	}

	debug *bool
)

func init() {
	configurationFlags.base64 = rootCmd.PersistentFlags().StringP("base64-configuration", "b", "", "Full YAML configuration as base64 string")
	configurationFlags.file = rootCmd.PersistentFlags().StringP("configuration-file", "c", "lagoon.yml", "Path of the YAML configuration file")
	debug = rootCmd.PersistentFlags().BoolP("debug", "d", false, "Start Lagoon in debug mode")
}

func setupRouter() *gin.Engine {
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true

	r.Use(cors.Default())

	// Create a data source
	r.POST(contextPath+"/datasource", func(c *gin.Context) {
		api.CreateNewDataSource(c)
	})
	r.PATCH(contextPath+"/datasource", func(c *gin.Context) {
		api.UpdateDataSource(c)
	})
	r.DELETE(contextPath+"/datasource/:DataSourceId", func(c *gin.Context) {
		api.DeleteDataSource(c)
	})
	r.GET(contextPath+"/datasource", func(c *gin.Context) {
		api.GetDataSources(c)
	})

	// list entry points
	r.GET(contextPath+"/data/:DataSourceId/entrypoint", func(c *gin.Context) {
		api.ListEntryPoints(c)
	})

	r.GET(contextPath+"/data/:DataSourceId/infos", func(c *gin.Context) {
		api.GetInfos(c)
	})

	r.GET(contextPath+"/data/:DataSourceId/state", func(c *gin.Context) {
		api.GetState(c)
	})

	r.GET(contextPath+"/data/:DataSourceId/entrypoint/:entrypoint/info", func(c *gin.Context) {
		api.GetEntryPointInfos(c)
	})

	r.GET(contextPath+"/data/:DataSourceId/entrypoint/:entrypoint/content", func(c *gin.Context) {
		api.GetEntryPointContent(c)
	})

	r.DELETE(contextPath+"/data/:DataSourceId/entrypoint/:entrypoint", func(c *gin.Context) {
		api.DeleteEntryPoint(c)
	})

	r.DELETE(contextPath+"/data/:DataSourceId/entrypoint/:entrypoint/children", func(c *gin.Context) {
		api.DeleteEntryPointChildren(c)
	})

	r.POST(contextPath+"/data/:DataSourceId/command", func(c *gin.Context) {
		api.ExecuteCommand(c)
	})

	// Consume web-socket.
	r.GET(contextPath+"/ws/:wsUuid", func(c *gin.Context) {
		api.ReadChannelContentAndSendToWebSocket(c)
	})

	r.Static(contextPath+"/ui", "./ui/dist")

	r.GET(contextPath+"/", func(context *gin.Context) {
		log.Printf("Context: %s\n", context.Request.RequestURI)
		context.Redirect(http.StatusMovedPermanently, contextPath+"/ui")
	})

	r.GET("/", func(context *gin.Context) {
		log.Printf("Context: %s\n", context.Request.RequestURI)
		context.Redirect(http.StatusMovedPermanently, contextPath+"/ui")
	})

	return r
}

func main() {
	rootCmd.Execute()
}

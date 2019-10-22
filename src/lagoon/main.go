package main

import (
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
	"strconv"

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
				dsInfos, _ := api.CreateDataSourceFromDescriptor(ds)
				api.DataSourcesHeaders = append(api.DataSourcesHeaders, dsInfos)
			}
		}

		// Listen and Server in 0.0.0.0:port
		if configuration.Port == 0 {
			configuration.Port = 4000
		}

		r := setupRouter()

		log.Printf("Starting Lagoon on port %d", configuration.Port)
		r.Run(":" + strconv.Itoa(configuration.Port))
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
	r.PUT(contextPath+"/datasource", func(c *gin.Context) {
		api.CreateNewDataSource(c)
	})
	r.GET(contextPath+"/datasource", func(c *gin.Context) {
		log.Printf("Current data sources: %v \n", api.DataSourcesHeaders)
		c.JSON(http.StatusOK, gin.H{"datasources": api.DataSourcesHeaders})
	})

	// list entry points
	r.GET(contextPath+"/data/:DataSourceUuid/entrypoint", func(c *gin.Context) {
		api.ListEntryPoints(c)
	})

	r.GET(contextPath+"/data/:DataSourceUuid/entrypoint/:entrypoint/info", func(c *gin.Context) {
		api.GetEntryPointInfos(c)
	})

	r.GET(contextPath+"/data/:DataSourceUuid/entrypoint/:entrypoint/content", func(c *gin.Context) {
		api.GetEntryPointContent(c)
	})

	r.DELETE(contextPath+"/data/:DataSourceUuid/entrypoint/:entrypoint", func(c *gin.Context) {
		api.DeleteEntryPoint(c)
	})

	r.DELETE(contextPath+"/data/:DataSourceUuid/entrypoint/:entrypoint/children", func(c *gin.Context) {
		api.DeleteEntryPointChildren(c)
	})

	// list entry points
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

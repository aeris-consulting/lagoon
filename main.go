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
)

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
				dsUuid, _ := api.CreateDatasource(&ds)
				dsInfos := api.DataSourceInfos{
					Uuid:        dsUuid,
					Vendor:      ds.Vendor,
					Name:        ds.Name,
					Description: ds.Description,
				}
				api.DataSources = append(api.DataSources, dsInfos)
			}
		}

		r := setupRouter()
		// Listen and Server in 0.0.0.0:port
		if configuration.Port == 0 {
			configuration.Port = 4000
		}
		r.Run(":" + strconv.Itoa(configuration.Port))
	},
}

var (
	configurationFlags struct {
		base64 *string
		file   *string
	}

	configuration struct {
		Port        int                     `yaml:"port"`
		Datasources []datasource.DataSource `yaml:"datasources"`
	}
)

func init() {
	configurationFlags.base64 = rootCmd.PersistentFlags().StringP("base64-configuration", "b", "", "Full YAML configuration as base64 string")
	configurationFlags.file = rootCmd.PersistentFlags().StringP("configuration-file", "c", "lagoon.yml", "Path of the YAML configuration file")
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.Use(cors.Default())

	r.Static("/ui", "./ui/dist")

	// Create a data source
	r.PUT("/datasource", func(c *gin.Context) {
		api.CreateNewDataSource(c)
	})
	r.GET("/datasource", func(c *gin.Context) {
		log.Printf("Current data sources: %v \n", api.DataSources)
		c.JSON(http.StatusOK, gin.H{"datasources": api.DataSources})
	})

	// list entry points
	r.GET("/data/:DataSourceUuid/entrypoint", func(c *gin.Context) {
		api.ListEntryPoints(c)
	})

	r.GET("/data/:DataSourceUuid/entrypoint/:entrypoint/info", func(c *gin.Context) {
		api.GetEntryPointInfos(c)
	})

	r.GET("/data/:DataSourceUuid/entrypoint/:entrypoint/content", func(c *gin.Context) {
		api.GetEntryPointContent(c)
	})

	r.DELETE("/data/:DataSourceUuid/entrypoint/:entrypoint", func(c *gin.Context) {
		api.DeleteEntryPoint(c)
	})

	r.DELETE("/data/:DataSourceUuid/entrypoint/:entrypoint/children", func(c *gin.Context) {
		api.DeleteEntryPointChildren(c)
	})

	// list entry points
	r.GET("/ws/:wsUuid", func(c *gin.Context) {
		api.ReadChannelContentAndSendToWebSocket(c)
	})

	// Ping test
	r.GET("/ping/:DataSourceUuid", func(c *gin.Context) {
		_, ok := api.FindDatasource(c)
		if ok {
			dataSourceUuid := c.Params.ByName("DataSourceUuid")
			log.Printf("Datasource %s pinged\n", dataSourceUuid)
			c.String(http.StatusOK, "pong")
		}
	})

	r.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/ui")
	})

	return r
}

func main() {
	rootCmd.Execute()
}

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/twinj/uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"lagoon/datasource"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const MaxLevel = ^uint(0)

type dataSourceUuid string

type dataSourceInfos struct {
	Uuid        dataSourceUuid `json:"uuid" binding:"required"`
	Vendor      string         `json:"vendor" binding:"required"`
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
}

var clients = make(map[dataSourceUuid]datasource.Datasource)

var webSocketChannels = make(map[string]chan datasource.DataBatch)
var webSocketErrorChannels = make(map[string]chan error)

var dataSources []dataSourceInfos

var webSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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
				dsUuid, _ := createDatasource(&ds)
				dsInfos := dataSourceInfos{
					Uuid:        dsUuid,
					Vendor:      ds.Vendor,
					Name:        ds.Name,
					Description: ds.Description,
				}
				dataSources = append(dataSources, dsInfos)
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
		createNewDataSource(c)
	})
	r.GET("/datasource", func(c *gin.Context) {
		log.Printf("Current data sources: %v \n", dataSources)
		c.JSON(http.StatusOK, gin.H{"datasources": dataSources})
	})

	// list entry points
	r.GET("/data/:dataSourceUuid/entrypoint", func(c *gin.Context) {
		listEntryPoints(c)
	})

	r.GET("/data/:dataSourceUuid/entrypoint/:entrypoint/info", func(c *gin.Context) {
		getEntryPointInfos(c)
	})

	r.GET("/data/:dataSourceUuid/entrypoint/:entrypoint/content", func(c *gin.Context) {
		getEntryPointContent(c)
	})

	r.DELETE("/data/:dataSourceUuid/entrypoint/:entrypoint", func(c *gin.Context) {
		deleteEntryPoint(c)
	})

	r.DELETE("/data/:dataSourceUuid/entrypoint/:entrypoint/children", func(c *gin.Context) {
		deleteEntryPointChildren(c)
	})

	// list entry points
	r.GET("/ws/:wsUuid", func(c *gin.Context) {
		readChannelContentAndSendToWebSocket(c)
	})

	// Ping test
	r.GET("/ping/:dataSourceUuid", func(c *gin.Context) {
		_, ok := findDatasource(c)
		if ok {
			dataSourceUuid := c.Params.ByName("dataSourceUuid")
			log.Printf("Datasource %s pinged\n", dataSourceUuid)
			c.String(http.StatusOK, "pong")
		}
	})

	r.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/ui")
	})

	return r
}

func createNewDataSource(c *gin.Context) {
	var dataSource datasource.DataSource
	if c.Bind(&dataSource) == nil {
		log.Printf("Creating data source %v\n", dataSource)
		dataSourceUuid, err := createDatasource(&dataSource)

		if err == nil {
			dsInfos := dataSourceInfos{
				Uuid:        dataSourceUuid,
				Vendor:      dataSource.Vendor,
				Name:        dataSource.Name,
				Description: dataSource.Description,
			}
			dataSources = append(dataSources, dsInfos)
			log.Printf("Current data sources: %v \n", dataSources)
			c.JSON(http.StatusOK, gin.H{"dataSourceUuid": dataSourceUuid})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func createDatasource(source *datasource.DataSource) (dataSourceUuid, error) {
	var err error

	var dataSource datasource.Datasource
	vendor := strings.TrimSpace(strings.ToLower(source.Vendor))
	switch vendor {
	case "redis":
		dataSource, err = createRedisClient(source)
	default:
		err = errors.New(fmt.Sprintf("Vendor %s is unknown", vendor))
		log.Printf("ERROR: %s\n", err.Error())
	}

	var dSUuid dataSourceUuid
	if err == nil {
		err = dataSource.Open()
		if err == nil {
			if source.Uuid != "" {
				dSUuid = dataSourceUuid(source.Uuid)
			} else {
				dSUuid = dataSourceUuid(uuid.NewV4().String())
			}
			clients[dSUuid] = dataSource
		}
	}

	return dSUuid, err
}

func createRedisClient(source *datasource.DataSource) (datasource.Datasource, error) {
	datasource := datasource.RedisClient{
		Datasource: source,
	}
	err := datasource.Open()

	return &datasource, err
}

func listEntryPoints(c *gin.Context) {
	cli, ok := findDatasource(c)
	if ok {
		filter := getFilter(c)
		minLevel := getMinLevel(c)
		maxLevel := getMaxLevel(c)
		entrypointsChannel := make(chan datasource.DataBatch, datasource.SwitchToWsBarrier)
		status, err := cli.ListEntryPoints(filter, entrypointsChannel, minLevel, maxLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		} else if status == datasource.Moved {
			wsUuid := uuid.NewV4().String()
			webSocketChannels[wsUuid] = entrypointsChannel
			c.JSON(http.StatusAccepted, gin.H{"link": fmt.Sprintf("/ws/%s", wsUuid)})

		} else if status == datasource.Completed {
			dataBatch := <-entrypointsChannel
			close(entrypointsChannel)
			c.JSON(http.StatusOK, gin.H{"size": dataBatch.Size, "data": dataBatch.Data})
		}
	}
}

func getFilter(c *gin.Context) string {
	filter, exists := c.GetQuery("filter")
	if !exists {
		filter = "*"
	}
	return filter
}

func getMinLevel(c *gin.Context) uint {
	levelParam, exists := c.GetQuery("min")
	if exists {
		minLevel, error := strconv.Atoi(levelParam)
		if error == nil {
			return uint(minLevel)
		}
	}
	return 0
}

func getMaxLevel(c *gin.Context) uint {
	levelParam, exists := c.GetQuery("max")
	if exists {
		minLevel, error := strconv.Atoi(levelParam)
		if error == nil {
			return uint(minLevel)
		}
	}
	return MaxLevel
}

func getEntryPointInfos(c *gin.Context) {
	cli, ok := findDatasource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		infos, err := cli.GetEntryPointInfos(entrypoint)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"type": datasource.EntryPointTypesAsString[infos.Type], "length": infos.Length})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func getEntryPointContent(c *gin.Context) {
	cli, ok := findDatasource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		filter := getFilter(c)
		dataChannel := make(chan datasource.DataBatch, datasource.SwitchToWsBarrier)
		status, err := cli.GetContent(entrypoint, filter, dataChannel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else if status == datasource.Moved {
			wsUuid := uuid.NewV4().String()
			webSocketChannels[wsUuid] = dataChannel
			c.JSON(http.StatusAccepted, gin.H{"link": fmt.Sprintf("/ws/%s", wsUuid)})

		} else if status == datasource.Completed {
			dataBatch := <-dataChannel
			close(dataChannel)
			c.JSON(http.StatusOK, gin.H{"size": dataBatch.Size, "data": dataBatch.Data})
		}
	}
}

func deleteEntryPoint(c *gin.Context) {
	cli, ok := findDatasource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		err := cli.DeleteEntrypoint(entrypoint)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Entry point was deleted"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func deleteEntryPointChildren(c *gin.Context) {
	cli, ok := findDatasource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		errorChannel := make(chan error, datasource.SwitchToWsBarrier)
		_, err := cli.DeleteEntrypointChidren(entrypoint, errorChannel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			wsUuid := uuid.NewV4().String()
			webSocketErrorChannels[wsUuid] = errorChannel
			c.JSON(http.StatusAccepted, gin.H{"link": fmt.Sprintf("/ws/%s", wsUuid)})
		}
	}
}

func findDatasource(c *gin.Context) (datasource.Datasource, bool) {
	datasourceUuid := dataSourceUuid(c.Params.ByName("dataSourceUuid"))
	cli, ok := clients[datasourceUuid]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Datasource with UUID %s was not found", datasourceUuid)})
	}
	return cli, ok
}

func readChannelContentAndSendToWebSocket(c *gin.Context) {
	var (
		errorChannel chan error
	)

	wsUuid := c.Params.ByName("wsUuid")
	dataChannel, ok := webSocketChannels[wsUuid]
	if !ok {
		errorChannel, ok = webSocketErrorChannels[wsUuid]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Web socket channel wuth UUID %s was not found", wsUuid)})
			return
		} else {
			delete(webSocketErrorChannels, wsUuid)
		}
	} else {
		delete(webSocketChannels, wsUuid)
	}

	upgrader := websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin:       func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %+v\n", err)
		return
	}

	log.Printf("Reading channel data for %s\n", wsUuid)
	if dataChannel != nil {
		for data := range dataChannel {
			// Write messages of the channel
			json, err := json.Marshal(data)
			if err == nil {
				err = conn.WriteMessage(websocket.TextMessage, json)
			} else {
				log.Printf("ERROR while converting %v: %s\n", data, err.Error())
			}
		}
	}

	if errorChannel != nil {
		for error := range errorChannel {
			// Write messages of the channel
			json, err := json.Marshal(error.Error())
			if err == nil {
				err = conn.WriteMessage(websocket.TextMessage, json)
			}
		}
	}
	log.Printf("Stop reading channel data for %s\n", wsUuid)
}

func main() {
	rootCmd.Execute()
}

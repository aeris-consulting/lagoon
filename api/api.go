package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
	"lagoon/datasource"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const MaxLevel = ^uint(0)

type DataSourceUuid string

type DataSourceInfos struct {
	Uuid        DataSourceUuid `json:"uuid" binding:"required"`
	Vendor      string         `json:"vendor" binding:"required"`
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
}

var clients = make(map[DataSourceUuid]datasource.Datasource)

var webSocketChannels = make(map[string]chan datasource.DataBatch)
var webSocketErrorChannels = make(map[string]chan error)

var DataSources []DataSourceInfos

var webSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func CreateNewDataSource(c *gin.Context) {
	var dataSource datasource.DataSource
	if c.Bind(&dataSource) == nil {
		log.Printf("Creating data source %v\n", dataSource)
		dataSourceUuid, err := CreateDatasource(&dataSource)

		if err == nil {
			dsInfos := DataSourceInfos{
				Uuid:        dataSourceUuid,
				Vendor:      dataSource.Vendor,
				Name:        dataSource.Name,
				Description: dataSource.Description,
			}
			DataSources = append(DataSources, dsInfos)
			log.Printf("Current data sources: %v \n", DataSources)
			c.JSON(http.StatusOK, gin.H{"DataSourceUuid": dataSourceUuid})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func CreateDatasource(source *datasource.DataSource) (DataSourceUuid, error) {
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

	var dSUuid DataSourceUuid
	if err == nil {
		err = dataSource.Open()
		if err == nil {
			if source.Uuid != "" {
				dSUuid = DataSourceUuid(source.Uuid)
			} else {
				dSUuid = DataSourceUuid(uuid.NewV4().String())
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

func ListEntryPoints(c *gin.Context) {
	cli, ok := FindDatasource(c)
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

func GetEntryPointInfos(c *gin.Context) {
	cli, ok := FindDatasource(c)
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

func GetEntryPointContent(c *gin.Context) {
	cli, ok := FindDatasource(c)
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

func DeleteEntryPoint(c *gin.Context) {
	cli, ok := FindDatasource(c)
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

func DeleteEntryPointChildren(c *gin.Context) {
	cli, ok := FindDatasource(c)
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

func FindDatasource(c *gin.Context) (datasource.Datasource, bool) {
	datasourceUuid := DataSourceUuid(c.Params.ByName("DataSourceUuid"))
	cli, ok := clients[datasourceUuid]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Datasource with UUID %s was not found", datasourceUuid)})
	}
	return cli, ok
}

func ReadChannelContentAndSendToWebSocket(c *gin.Context) {
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

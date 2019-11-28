package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
	"lagoon/datasource"
	"log"
	"net/http"
	"strconv"
	"time"
)

type DataSourceHeader struct {
	Uuid        datasource.DataSourceUuid `json:"uuid" binding:"required"`
	Vendor      string                    `json:"vendor" binding:"required"`
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
}

type CommandRequest struct {
	Args   []interface{} `json:"args" binding:"required`
	NodeID string        `json:"nodeId"`
}

var DataSourcesHeaders []DataSourceHeader

var dataSources = make(map[datasource.DataSourceUuid]datasource.DataSource)
var webSocketChannels = make(map[string]chan datasource.DataBatch)
var webSocketErrorChannels = make(map[string]chan error)

func CreateNewDataSource(c *gin.Context) {
	var dataSourceDescriptor datasource.DataSourceDescriptor
	if err := c.Bind(&dataSourceDescriptor); err == nil {
		datasource, err := CreateDataSourceFromDescriptor(dataSourceDescriptor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			DataSourcesHeaders = append(DataSourcesHeaders, datasource)
			log.Printf("Current data sources: %v \n", DataSourcesHeaders)
			c.JSON(http.StatusOK, gin.H{"DataSourceUuid": datasource.Uuid})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func CreateDataSourceFromDescriptor(dataSourceDescriptor datasource.DataSourceDescriptor) (DataSourceHeader, error) {
	log.Printf("Creating data source %v\n", dataSourceDescriptor)
	dataSource, err := datasource.CreateDataSource(&dataSourceDescriptor)

	if err == nil {
		var dataSourceUuid datasource.DataSourceUuid
		if dataSourceDescriptor.Uuid != "" {
			dataSourceUuid = datasource.DataSourceUuid(dataSourceDescriptor.Uuid)
		} else {
			dataSourceUuid = datasource.DataSourceUuid(uuid.NewV4().String())
		}
		dataSources[dataSourceUuid] = dataSource

		dsInfos := DataSourceHeader{
			Uuid:        dataSourceUuid,
			Vendor:      dataSourceDescriptor.Vendor,
			Name:        dataSourceDescriptor.Name,
			Description: dataSourceDescriptor.Description,
		}
		return dsInfos, nil
	}
	return DataSourceHeader{}, err
}

func ListEntryPoints(c *gin.Context) {
	ds, ok := findDataSource(c)
	if ok {
		filter := getFilter(c)
		minLevel := getMinLevel(c)
		maxLevel := getMaxLevel(c)
		entrypointsChannel := make(chan datasource.DataBatch, datasource.SwitchToWsBarrier)
		status, err := ds.ListEntryPoints(filter, entrypointsChannel, minLevel, maxLevel)
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
	return datasource.MaxLevel
}

func GetEntryPointInfos(c *gin.Context) {
	ds, ok := findDataSource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		infos, err := ds.GetEntryPointInfos(entrypoint)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"type": datasource.EntryPointTypesAsString[infos.Type], "length": infos.Length, "timeToLive": infos.TimeToLive / time.Millisecond})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func GetEntryPointContent(c *gin.Context) {
	ds, ok := findDataSource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		filter := getFilter(c)
		dataChannel := make(chan datasource.DataBatch, datasource.SwitchToWsBarrier)
		status, err := ds.GetContent(entrypoint, filter, dataChannel)
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
	ds, ok := findDataSource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		err := ds.DeleteEntrypoint(entrypoint)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Entry point was deleted"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func DeleteEntryPointChildren(c *gin.Context) {
	ds, ok := findDataSource(c)
	if ok {
		entrypoint := datasource.EntryPoint(c.Params.ByName("entrypoint"))
		errorChannel := make(chan error, datasource.SwitchToWsBarrier)
		_, err := ds.DeleteEntrypointChildren(entrypoint, errorChannel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			wsUuid := uuid.NewV4().String()
			webSocketErrorChannels[wsUuid] = errorChannel
			c.JSON(http.StatusAccepted, gin.H{"link": fmt.Sprintf("/ws/%s", wsUuid)})
		}
	}
}

func findDataSource(c *gin.Context) (datasource.DataSource, bool) {
	datasourceUuid := datasource.DataSourceUuid(c.Params.ByName("DataSourceUuid"))
	ds, ok := dataSources[datasourceUuid]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("DataSource with UUID %s was not found", datasourceUuid)})
	}
	return ds, ok
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
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin:       func(r *http.Request) bool { return true },
	}

	// Force headers for calls behind reverse-proxy.
	c.Request.Header.Set("Connection", "upgrade")
	c.Request.Header.Set("Upgrade", "websocket")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v\n", err)
		return
	}
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("Websocket %v closed with code %v and reason %v\n", wsUuid, code, text)
		return nil
	})

	defer func() {
		go func() {
			time.Sleep(30 * time.Second)
			log.Printf("Websocket %v expired\n", wsUuid)
			conn.Close()
		}()
	}()

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
			// Write errors of the channel
			json, err := json.Marshal(error.Error())
			if err == nil {
				err = conn.WriteMessage(websocket.TextMessage, json)
			}
		}
	}

	log.Printf("Stop reading channel data for %s\n", wsUuid)
}

func ExecuteCommand(c *gin.Context) {
	var commandRequest CommandRequest
	if c.Bind(&commandRequest) == nil {
		ds, ok := findDataSource(c)
		if ok {
			message, err := ds.ExecuteCommand(commandRequest.Args, commandRequest.NodeID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"data": message})
			}
		}
	}
}

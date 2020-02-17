package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
	"lagoon/datasource"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

const webSocketBatchSize = uint64(100)

type DataSourceHeader struct {
	Id          datasource.DataSourceId `json:"id" binding:"required"`
	Vendor      string                  `json:"vendor" binding:"required"`
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	ReadOnly    bool                    `json:"readonly"`
}

type CommandRequest struct {
	Args   []interface{} `json:"args" binding:"required`
	NodeID string        `json:"nodeId"`
}

var DataSourcesHeaders = make(map[datasource.DataSourceId]DataSourceHeader)

var dataSources = make(map[datasource.DataSourceId]datasource.DataSource)
var webSocketChannels = make(map[string]chan datasource.DataBatch)
var webSocketErrorChannels = make(map[string]chan error)

func CloseDataSources() {
	log.Println("Closing all data sources...")
	for _, ds := range dataSources {
		ds.Close()
	}
	log.Println("Data sources closed")
}

func CreateNewDataSource(c *gin.Context) {
	// TODO Validate: https://gin-gonic.com/docs/examples/custom-validators/
	var dataSourceDescriptor datasource.DataSourceDescriptor
	if err := c.Bind(&dataSourceDescriptor); err == nil {
		datasource, err := CreateDataSourceFromDescriptor(dataSourceDescriptor, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			DataSourcesHeaders[datasource.Id] = datasource
			log.Printf("Current data sources: %v \n", DataSourcesHeaders)
			c.JSON(http.StatusOK, gin.H{"dataSourceId": datasource.Id})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func UpdateDataSource(c *gin.Context) {
	var dataSourceDescriptor datasource.DataSourceDescriptor
	if err := c.Bind(&dataSourceDescriptor); err == nil {
		dataSourceId := datasource.DataSourceId(dataSourceDescriptor.Id)
		existingDataSource, exists := dataSources[dataSourceId]
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("The datasource with id '%s' does not exists", dataSourceId)})
			return
		}
		existingDataSource.Close()

		datasource, err := CreateDataSourceFromDescriptor(dataSourceDescriptor, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			// Replace the existing data source.
			DataSourcesHeaders[datasource.Id] = datasource
			log.Printf("Current data sources: %v \n", DataSourcesHeaders)
			c.Status(http.StatusNoContent)
			c.Writer.WriteHeaderNow()
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func CreateDataSourceFromDescriptor(dataSourceDescriptor datasource.DataSourceDescriptor, new bool) (DataSourceHeader, error) {
	log.Printf("Creating data source %v\n", dataSourceDescriptor)
	dataSource, err := datasource.CreateDataSource(&dataSourceDescriptor)

	if err == nil {
		var dataSourceId datasource.DataSourceId
		if dataSourceDescriptor.Id != "" {
			dataSourceId = datasource.DataSourceId(dataSourceDescriptor.Id)
			if _, exists := dataSources[dataSourceId]; exists && new {
				return DataSourceHeader{}, errors.New(fmt.Sprintf("The datasource with id '%s' already exists", dataSourceId))
			}
		} else {
			dataSourceId = datasource.DataSourceId(uuid.NewV4().String())
		}
		dataSources[dataSourceId] = dataSource

		dsInfos := DataSourceHeader{
			Id:          dataSourceId,
			Vendor:      dataSourceDescriptor.Vendor,
			Name:        dataSourceDescriptor.Name,
			Description: dataSourceDescriptor.Description,
			ReadOnly:    dataSourceDescriptor.ReadOnly,
		}
		return dsInfos, nil
	}
	return DataSourceHeader{}, err
}

func GetInfos(c *gin.Context) {
	ds, ok := findDataSource(c)
	if ok {
		infos, err := ds.GetInfos()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		} else {
			c.JSON(http.StatusOK, gin.H{"infos": infos})
		}
	}
}

func GetState(c *gin.Context) {
	ds, ok := findDataSource(c)
	if ok {
		status, err := ds.GetStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		} else {
			c.JSON(http.StatusOK, gin.H{"status": status})
		}
	}
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
	datasourceUuid := datasource.DataSourceId(c.Params.ByName("DataSourceId"))
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
			time.Sleep(10 * time.Second)
			log.Printf("Websocket %v expired\n", wsUuid)
			conn.Close()
		}()
	}()

	log.Printf("Reading channel data for %s\n", wsUuid)
	if dataChannel != nil {
		for data := range dataChannel {
			for _, dataItem := range splitBatch(data) {
				err = conn.WriteJSON(dataItem)
				if err != nil {
					log.Printf("ERROR while converting %v: %s\n", dataItem, err.Error())
				}
			}
		}
	}

	if errorChannel != nil {
		for error := range errorChannel {
			err = conn.WriteJSON(error.Error())
			if err != nil {
				log.Printf("ERROR while sending %v\n", err.Error())
			}
		}
	}
	log.Printf("Stop reading channel data for %s\n", wsUuid)
}

func splitBatch(dataToSplit datasource.DataBatch) []datasource.DataBatch {
	if dataToSplit.Size <= webSocketBatchSize {
		return []datasource.DataBatch{dataToSplit}
	} else {
		result := []datasource.DataBatch{}
		for i := uint64(0); i < dataToSplit.Size; i += webSocketBatchSize {
			d := datasource.DataBatch{}
			start := i
			end := uint64(math.Min(float64(i+webSocketBatchSize), float64(dataToSplit.Size)))
			for j := start; j < end; j++ {
				d.Data = append(d.Data, dataToSplit.Data[j])
			}
			d.Size = uint64(len(d.Data))
			result = append(result, d)
		}
		return result
	}
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

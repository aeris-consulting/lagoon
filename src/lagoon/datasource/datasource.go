package datasource

import (
	"errors"
	"log"
	"time"
)

type DataSourceId string

type ActionStatus uint8

type EntryPointType uint8

const MaxLevel = ^uint(0)

const (
	None      ActionStatus = 0
	Completed ActionStatus = 1
	Moved     ActionStatus = 2

	Value     EntryPointType = 1
	Set       EntryPointType = 2
	ScoredSet EntryPointType = 3
	List      EntryPointType = 4
	Hash      EntryPointType = 5
	Stream    EntryPointType = 6

	SwitchToWsBarrier uint8 = 20
)

var EntryPointTypesAsString = map[EntryPointType]string{
	Value:     "VALUE",
	Set:       "SET",
	ScoredSet: "SCORED_SET",
	List:      "LIST",
	Hash:      "HASH",
	Stream:    "STREAM",
}

type DataBatch struct {
	Size uint64        `json:"size" binding:"required"`
	Data []interface{} `json:"data" binding:"required"`
}

type DataSourceDescriptor struct {
	Id            string            `json:"id" yaml:"id"`
	Vendor        string            `json:"vendor" yaml:"vendor" binding:"required"`
	Name          string            `json:"name" yaml:"name" binding:"required"`
	Description   string            `json:"description"yaml:"description"`
	Bootstrap     string            `json:"bootstrap" yaml:"bootstrap" binding:"required"`
	ReadOnly      bool              `json:"readonly" yaml:"readonly"`
	User          string            `json:"user" yaml:"user"`
	Password      string            `json:"password" yaml:"password"`
	Configuration map[string]string `json:"configuration" yaml:"configuration`
}

type EntryPoint string

type EntryPointNode struct {
	Path       EntryPoint `json:"path" binding:"required"`
	HasContent bool       `json:"hasContent" binding:"required"`
	Length     uint64     `json:"length" binding:"required"`
}

type EntryPointInfos struct {
	Type       EntryPointType `json:"type" binding:"required"`
	Length     uint64         `json:"length" binding:"required"`
	TimeToLive time.Duration  `json:"timeToLive" binding:"required"`
}

type SingleValue interface{}

type StreamInfos struct {
}

type Filter struct {
}

type ClusterNode struct {
	Id      string   `json:"id"`
	Server  string   `json:"server"`
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	Masters []string `json:"masters"`
}

type Cluster struct {
	Nodes []ClusterNode `json:"nodes"`
}

type StateSection struct {
	Name   string                 `json:"name"`
	Values map[string]interface{} `json:"values"`
}

type NodeState struct {
	NodeId        string         `json:"nodeId"`
	StateSections []StateSection `json:"sections"`
}

type ClusterState struct {
	Timestamp     time.Time      `json:"timestamp"`
	NodeStates    []NodeState    `json:"nodeStates"`
	StateSections []StateSection `json:"sections"`
}

var (
	ErrUnkownDatasource = errors.New("the specified kind of datasource is not known")
	vendors             = []Vendor{}
)

// DeclareImplementation allows to the vendor plugins to register themselves at initialization time.
func DeclareImplementation(vendor Vendor) {
	vendors = append(vendors, vendor)
}

// CreateDataSource is a general purpose function to create a data source from a descriptor, when its vendore is supported.
func CreateDataSource(source *DataSourceDescriptor) (DataSource, error) {
	var err error

	var dataSource DataSource
	for _, vendor := range vendors {
		if vendor.Accept(source) {
			dataSource, err = vendor.CreateDataSource(source)
		}
	}

	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	} else if dataSource == nil {
		err = ErrUnkownDatasource
		log.Printf("ERROR: %s\n", err.Error())
	} else {

		err = dataSource.Open()
	}
	return dataSource, err
}

// Vendor represents a parent interface for providers of data sources.
type Vendor interface {
	// Accept checks if the vendor is able to create a data source for the descriptor passed as parameter.
	Accept(source *DataSourceDescriptor) bool

	// CreateDataSource creates the actual data source from the descriptor passed as parameter.
	CreateDataSource(source *DataSourceDescriptor) (DataSource, error)
}

// DataSource provides a common interface for all kind of supported Lagoon datasource.
// It might be that some implementations of DataSource do not support all the functions.
type DataSource interface {
	// Open creates the datasource and opens connections to the server.
	Open() error

	// Close closes the connections to the server, as well as all the streams.
	Close()

	// ListEntryPoints provides the full list of Redis keys, Kafka and RabbitMQ topics in the channel entrypoints.
	// In order to provide a more flexible way of listing them, a pattern can be passed, with * and ? als wildcards.
	ListEntryPoints(filter string, entrypoints chan<- DataBatch, minTreeLevel uint, maxTreeLevel uint) (ActionStatus, error)

	// GetEntryPointInfos returns the available details of the entrypoint: type, size...
	GetEntryPointInfos(entryPointValue EntryPoint) (EntryPointInfos, error)

	// GetValue returns the unique value when entryPointValue is attached to only one value, like string values in Redis.
	GetContent(entryPointValue EntryPoint, filter string, content chan<- DataBatch) (ActionStatus, error)

	DeleteEntrypoint(entryPointValue EntryPoint) error

	DeleteEntrypointChildren(entryPointValue EntryPoint, errorChannel chan<- error) (ActionStatus, error)

	// OpenStream consumes a stream or topic and add the accepted values to the channel.
	Consume(entryPointValue EntryPoint, values chan<- DataBatch, filter Filter, fromBeginning bool) (ActionStatus, error)

	// ExecuteCommand executes a native command and returns the result.
	ExecuteCommand(args []interface{}, nodeID string) (interface{}, error)

	// GetInfos provides essential information about the data source.
	GetInfos() (Cluster, error)

	// GetStatus provides essential status and health information about the data source.
	GetStatus() (ClusterState, error)
}

package datasource

type ActionStatus uint8

type EntryPointType uint8

const (
	None      ActionStatus = 0
	Completed ActionStatus = 1
	Moved     ActionStatus = 2

	Value     EntryPointType = 1
	Set       EntryPointType = 2
	SortedSet EntryPointType = 3
	List      EntryPointType = 4
	Hash      EntryPointType = 5
	Stream    EntryPointType = 6

	SwitchToWsBarrier uint8 = 20
)

var EntryPointTypesAsString = map[EntryPointType]string{
	Value:     "VALUE",
	Set:       "SET",
	SortedSet: "SORTED_SET",
	List:      "LIST",
	Hash:      "HASH",
	Stream:    "STREAM",
}

type DataBatch struct {
	Size uint64        `json:"size" binding:"required"`
	Data []interface{} `json:"data" binding:"required"`
}

type DataSource struct {
	Uuid          string            `json:"uuid" yaml:"uuid"`
	Vendor        string            `json:"vendor" yaml:"vendor" binding:"required"`
	Name          string            `json:"name" yaml:"name" binding:"required"`
	Description   string            `json:"description"yaml:"description"`
	Bootstrap     string            `json:"bootstrap" yaml:"bootstrap" binding:"required"`
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
	TimeToLive int64          `json:"timeToLive"` // Time to live of the entry, in seconds.
}

type EntryPointContent struct {
	Type       string                  `json:"type"`
	TimeToLive int64                   `json:"timeToLive"`
	Value      interface{}             `json:"value"`
	Hash       map[string]interface{}  `json:"hash"`
	Values     map[float64]interface{} `json:"values"`
}
type SingleValue interface{}

type StreamInfos struct{}

type Filter struct{}

// Datasource provides a common interface for all kind of supported Lagoon datasource.
// It might be that some implementations of Datasource do not support all the functions.
type Datasource interface {
	// Open creates the datasource and opens connections to the server.
	Open() error

	// Close closes the connections to the server, as well as all the streams.
	Close()

	// ListEntryPoints provides the full list of Redis keys, Kafka and RabbitMQ topics in the channel entrypoints.
	// In order to provide a more flexible way of listing them, a pattern can be passed, with * and ? als wildcards.
	ListEntryPoints(filter string, entrypoints chan<- DataBatch, minTreeLevel uint, maxTreeLevel uint) (ActionStatus, error)

	// GetEntryPointInfos returns the available details of the entrypoint: type, size...
	GetEntryPointInfos(entryPointValue EntryPoint) (EntryPointInfos, error)

	// GetContent returns the unique value when entryPointValue is attached to only one value, like string values in Redis.
	GetContent(entryPointValue EntryPoint, filter string, content chan<- DataBatch) (ActionStatus, error)

	// SetContent adds or updates the content of an entrypoint.
	SetContent(entryPointValue EntryPoint, content EntryPointContent) error

	// DeleteEntrypoint deletes a unique entrypoint content, preserving its children.
	DeleteEntrypoint(entryPointValue EntryPoint) error

	// DeleteEntrypointChidren deletes the content of the entrypoint and all its children.
	DeleteEntrypointChidren(entryPointValue EntryPoint, errorChannel chan<- error) (ActionStatus, error)

	// Consume reads a stream or topic and continually add the accepted values to the channel.
	Consume(entryPointValue EntryPoint, values chan<- DataBatch, filter Filter, fromBeginning bool) (ActionStatus, error)

	// GetSupportedTypes provides the types of entry points that this kind of Datasource can support for creation and display.
	GetSupportedTypes() []EntryPointType
}

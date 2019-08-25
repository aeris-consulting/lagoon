package datasource

type ActionStatus uint8

type EntryPointType uint8

const (
	None      ActionStatus = 0
	Completed ActionStatus = 1
	Moved     ActionStatus = 2

	Value  EntryPointType = 1
	Set    EntryPointType = 2
	List   EntryPointType = 3
	Hash   EntryPointType = 4
	Stream EntryPointType = 5

	SwitchToWsBarrier uint8 = 20
)

var EntryPointTypesAsString = map[EntryPointType]string{
	Value:  "VALUE",
	Set:    "SET",
	List:   "LIST",
	Hash:   "HASH",
	Stream: "STREAM",
}

type DataBatch struct {
	Size uint64        `json:"size" binding:"required"`
	Data []interface{} `json:"data" binding:"required"`
}

type DataSource struct {
	Vendor        string            `json:"vendor" binding:"required"`
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	Bootstrap     string            `json:"bootstrap" binding:"required"`
	User          string            `json:"user"`
	Password      string            `json:"password"`
	Configuration map[string]string `json:"configuration"`
}

type EntryPoint string

type EntryPointNode struct {
	Path       EntryPoint `json:"path" binding:"required"`
	HasContent bool       `json:"hasContent" binding:"required"`
	Length     uint64     `json:"length" binding:"required"`
}

type EntryPointInfos struct {
	Type   EntryPointType `json:"type" binding:"required"`
	Length uint64         `json:"length" binding:"required"`
}

type SingleValue interface{}

type StreamInfos struct {
}

type Filter struct {
}

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

	// GetValue returns the unique value when entryPointValue is attached to only one value, like string values in Redis.
	GetContent(entryPointValue EntryPoint, filter string, content chan<- DataBatch) (ActionStatus, error)

	DeleteEntrypoint(entryPointValue EntryPoint) error

	DeleteEntrypointChidren(entryPointValue EntryPoint, errorChannel chan<- error) (ActionStatus, error)

	// OpenStream consumes a stream or topic and add the accepted values to the channel.
	Consume(entryPointValue EntryPoint, values chan<- DataBatch, filter Filter, fromBeginning bool) (ActionStatus, error)
}

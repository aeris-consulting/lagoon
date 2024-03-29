package redis

// https://github.com/go-redis/redis/blob/master/example_test.go
import (
	"context"
	tls2 "crypto/tls"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"lagoon/datasource"
	"log"
	"reflect"
	regexp2 "regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const scanSize = int64(1000)

var invalidProtocolError struct {
	protocol string
}

type RedisClient struct {
	datasource       *datasource.DataSourceDescriptor
	client           redis.Cmdable
	readOnlyCommands []string
}

type RedisVendor struct {
}

type SortedSetValues struct {
	Score  float64  `json:"score"`
	Values []string `json:"values"`
}

type HashValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func init() {
	datasource.DeclareImplementation(&RedisVendor{})
}

const (
	pathSeparator         = rune(':')
	pathSeparatorAsString = ":"
	openingBracket        = rune('{')
	closingBracket        = rune('}')
)

func split(key string) (uint, []string) {
	tokenCount := uint(0)
	openBrackets := 0
	var token = make([]rune, len(key))
	runeIndexInToken := 0
	buffer := []string{}
	for _, r := range []rune(key) {
		// We split with pathSeparator only when no bracket is open.
		if r == pathSeparator && openBrackets == 0 {
			buffer = append(buffer, string(token[:runeIndexInToken]))
			tokenCount = tokenCount + 1
			runeIndexInToken = 0
		} else {
			token[runeIndexInToken] = r
			runeIndexInToken = runeIndexInToken + 1
			if r == openingBracket {
				openBrackets = openBrackets + 1
			} else if r == closingBracket && openBrackets > 0 {
				openBrackets = openBrackets - 1
			}
		}
	}
	if runeIndexInToken > 0 {
		buffer = append(buffer, string(token[:runeIndexInToken]))
		tokenCount = tokenCount + 1
	}
	return tokenCount, buffer
}

func (c *RedisVendor) Accept(source *datasource.DataSourceDescriptor) bool {
	return "redis" == strings.TrimSpace(strings.ToLower(source.Vendor))
}

func (c *RedisVendor) CreateDataSource(source *datasource.DataSourceDescriptor) (datasource.DataSource, error) {
	datasource := RedisClient{
		datasource: source,
	}
	err := datasource.Open()

	return &datasource, err
}

func (c *RedisClient) GetInfos() (datasource.Cluster, error) {
	switch c.client.(type) {
	case *redis.ClusterClient:
		client := c.client.(*redis.ClusterClient)
		return getClusterInfos(client)
	default:
		return datasource.Cluster{}, nil
	}
}

// getClusterInfos returns the list of nodes of the server.
func getClusterInfos(c *redis.ClusterClient) (datasource.Cluster, error) {
	result := []datasource.ClusterNode{}

	clusterNodes, err := c.ClusterNodes(context.Background()).Result()
	if err != nil {
		return datasource.Cluster{}, err
	}

	for _, i := range strings.Split(clusterNodes, "\n") {
		if i != "" {
			values := strings.Split(i, " ")
			node := datasource.ClusterNode{}
			node.Id = values[0]                          // Id of the node.
			node.Server = values[1]                      // Announced IP and normal port
			node.Name = strings.Split(values[1], "@")[0] // Announced IP and normal port and cluster bus port
			role := strings.Split(values[2], ",")        // Role of the node.
			node.Role = role[len(role)-1]
			if node.Role == "slave" { // When role is slave, next value is the ID of the master.
				node.Masters = []string{values[3]}
			}
			result = append(result, node)
		}
	}
	return datasource.Cluster{result}, nil
}

func (c *RedisClient) GetStatus() (datasource.ClusterState, error) {
	switch c.client.(type) {
	case *redis.ClusterClient:
		client := c.client.(*redis.ClusterClient)
		return getClusterStatus(client)
	default:
		result := datasource.ClusterState{
			Timestamp:  time.Now(),
			NodeStates: []datasource.NodeState{},
		}

		nodeState := datasource.NodeState{
			NodeId:        "",
			StateSections: []datasource.StateSection{},
		}
		client := c.client.(*redis.Client)
		err := getNodeInfo(client, &nodeState)
		if err != nil {
			return result, nil
		}
		result.NodeStates = append(result.NodeStates, nodeState)
		return result, nil
	}
}

// getClusterStatus returns the overall status of the cluster and the individual of each node.
func getClusterStatus(c *redis.ClusterClient) (datasource.ClusterState, error) {
	result := datasource.ClusterState{
		Timestamp:     time.Now(),
		NodeStates:    []datasource.NodeState{},
		StateSections: []datasource.StateSection{},
	}

	// First collect the infos at the cluster level.
	clusterInfos, err := c.ClusterInfo(context.Background()).Result()
	if err != nil {
		return result, err
	}

	clusterSection := datasource.StateSection{
		Name:   "Cluster",
		Values: make(map[string]interface{}),
	}
	for _, i := range strings.Split(clusterInfos, "\r\n") {
		convertClusterInfoAndPutValue(i, clusterSection.Values)
	}
	result.StateSections = append(result.StateSections, clusterSection)

	// Then collect at the node level.
	err = c.ForEachShard(context.Background(), func(ctx context.Context, nodeClient *redis.Client) error {
		cmd := nodeClient.Do(ctx, "cluster", "myid")
		if cmd.Err() == nil {
			nodeState := datasource.NodeState{
				NodeId:        cmd.String(),
				StateSections: []datasource.StateSection{},
			}
			err := getNodeInfo(nodeClient, &nodeState)
			if err != nil {
				return err
			}
			result.NodeStates = append(result.NodeStates, nodeState)
		}
		return err
	})
	return result, err
}

func getNodeInfo(nodeClient *redis.Client, nodeState *datasource.NodeState) error {
	nodeInfos, err := nodeClient.Info(context.Background()).Result()
	if err != nil {
		return err
	}
	section := datasource.StateSection{}
	for _, v := range strings.Split(nodeInfos, "\r\n") {
		if v != "" {
			if v[0] == '#' {
				if section.Name != "" {
					nodeState.StateSections = append(nodeState.StateSections, section)
				}
				section = datasource.StateSection{
					Name:   strings.SplitN(v, " ", 2)[1],
					Values: make(map[string]interface{}),
				}
			} else {
				convertClusterInfoAndPutValue(v, section.Values)
			}
		}
	}
	if section.Name != "" {
		nodeState.StateSections = append(nodeState.StateSections, section)
	}
	return nil
}

func convertClusterInfoAndPutValue(v string, section map[string]interface{}) {
	values := strings.Split(v, ":")
	if len(values) >= 2 { // Eliminate prefixes like "cluster info:"
		numValue, err := strconv.Atoi(values[len(values)-1])
		if err == nil {
			section[values[len(values)-2]] = numValue
		} else {
			section[values[len(values)-2]] = values[len(values)-1]
		}
	}
}

func (c *RedisClient) Open() error {
	err := c.createConnection()
	if err == nil {
		pong, err := c.client.Ping(context.Background()).Result()
		if err == nil {
			log.Printf("Connection status of ping: %v\n", pong)
		} else {
			log.Printf("ERROR When pinging: %s\n", err.Error())
		}
	}
	return err
}

func (c *RedisClient) initReadonlyCommands() {
	if c.datasource.ReadOnly && len(c.readOnlyCommands) == 0 {
		for _, v := range c.client.Command(context.Background()).Val() {
			if v.ReadOnly {
				c.readOnlyCommands = append(c.readOnlyCommands, strings.ToLower(v.Name))
			}
		}
		sort.Strings(c.readOnlyCommands)
	}
}

func (c *RedisClient) createConnection() error {
	var bootstrap = c.datasource.Bootstrap
	var parts = strings.Split(bootstrap, "://")
	switch parts[0] {
	case "cluster":
		return c.createClusterConnection(parts[1])
	case "sentinel":
		return c.createSentinelConnection(parts[1])
	case "redis":
		return c.createRedisConnection(parts[1])
	}
	return errors.New(fmt.Sprintf("Protocol %s is unkown for Redis", parts[0]))
}

func (c *RedisClient) createClusterConnection(url string) error {
	defaultOptions := redis.ClusterOptions{}

	var e error
	var user string
	if c.datasource.User != "" {
		user = c.datasource.User
	} else {
		user = defaultOptions.Username
	}
	var password string
	if c.datasource.Password != "" {
		password = c.datasource.Password
	} else {
		password = defaultOptions.Password
	}
	readTimeout := defaultOptions.ReadTimeout
	if _, ok := c.datasource.Configuration["readTimeout"]; ok {
		timeout, err := strconv.Atoi(c.datasource.Configuration["readTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	writeTimeout := defaultOptions.WriteTimeout
	if _, ok := c.datasource.Configuration["writeTimeout"]; ok {
		timeout, err := strconv.Atoi(c.datasource.Configuration["writeTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}
	maxIdleConns := defaultOptions.MaxIdleConns
	if _, ok := c.datasource.Configuration["maxIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["maxIdleConns"])
		if err != nil {
			maxIdleConns = c
		}
	}
	poolSize := defaultOptions.PoolSize
	if _, ok := c.datasource.Configuration["poolSize"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["poolSize"])
		if err != nil {
			poolSize = c
		}
	}
	readonly := defaultOptions.ReadOnly
	if _, ok := c.datasource.Configuration["readonly"]; ok {
		c, err := strconv.ParseBool(c.datasource.Configuration["readonly"])
		if err != nil {
			readonly = c
		}
	}
	routeRandomly := defaultOptions.RouteRandomly
	if _, ok := c.datasource.Configuration["routeRandomly"]; ok {
		c, err := strconv.ParseBool(c.datasource.Configuration["routeRandomly"])
		if err != nil {
			routeRandomly = c
		}
	}
	tls := false
	if _, ok := c.datasource.Configuration["tls-enabled"]; ok {
		c, err := strconv.ParseBool(c.datasource.Configuration["tls-enabled"])
		if err != nil {
			tls = c
		}
	}
	var tlsConfig tls2.Config
	if tls {
		insecureSkipVerify := false
		if _, ok := c.datasource.Configuration["insecureSkipVerify"]; ok {
			c, err := strconv.ParseBool(c.datasource.Configuration["insecureSkipVerify"])
			if err != nil {
				insecureSkipVerify = c
			}
		}
		tlsConfig = tls2.Config{
			InsecureSkipVerify: insecureSkipVerify,
		}
	}

	opts := redis.ClusterOptions{
		Addrs:         strings.Split(url, ","),
		ReadOnly:      readonly,
		RouteRandomly: routeRandomly,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Printf("Connected to the cluster %v \n", strings.Split(url, ","))
			return nil
		},
		Username:     user,
		Password:     password,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxIdleConns: maxIdleConns,
		TLSConfig:    &tlsConfig,
	}
	client := redis.NewClusterClient(&opts)
	c.client = client
	log.Printf("Connection to the cluster %v was created\n", strings.Split(url, ","))

	return e
}

func (c *RedisClient) createSentinelConnection(url string) error {
	defaultOptions := redis.FailoverOptions{}

	var e error
	var password string

	var user string
	if c.datasource.User != "" {
		user = c.datasource.User
	} else {
		user = defaultOptions.Username
	}
	if c.datasource.Password != "" {
		password = c.datasource.Password
	} else {
		password = defaultOptions.Password
	}

	readTimeout := defaultOptions.ReadTimeout
	if _, ok := c.datasource.Configuration["readTimeout"]; ok {
		timeout, err := strconv.Atoi(c.datasource.Configuration["readTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}

	writeTimeout := defaultOptions.WriteTimeout
	if _, ok := c.datasource.Configuration["writeTimeout"]; ok {
		timeout, err := strconv.Atoi(c.datasource.Configuration["writeTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}

	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}
	maxIdleConns := defaultOptions.MaxIdleConns
	if _, ok := c.datasource.Configuration["maxIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["maxIdleConns"])
		if err != nil {
			maxIdleConns = c
		}
	}
	poolSize := defaultOptions.PoolSize
	if _, ok := c.datasource.Configuration["poolSize"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["poolSize"])
		if err != nil {
			poolSize = c
		}
	}
	db := defaultOptions.DB
	if _, ok := c.datasource.Configuration["db"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["db"])
		if err != nil {
			db = c
		}
	}
	tls := false
	if _, ok := c.datasource.Configuration["tls-enabled"]; ok {
		c, err := strconv.ParseBool(c.datasource.Configuration["tls-enabled"])
		if err != nil {
			tls = c
		}
	}
	var tlsConfig tls2.Config
	if tls {
		insecureSkipVerify := false
		if _, ok := c.datasource.Configuration["insecureSkipVerify"]; ok {
			c, err := strconv.ParseBool(c.datasource.Configuration["insecureSkipVerify"])
			if err != nil {
				insecureSkipVerify = c
			}
		}
		tlsConfig = tls2.Config{
			InsecureSkipVerify: insecureSkipVerify,
		}
	}

	opts := redis.FailoverOptions{
		MasterName:    c.datasource.Configuration["master"],
		SentinelAddrs: strings.Split(url, ","),
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Printf("Connected to the sentinels %v \n", strings.Split(url, ","))
			return nil
		},
		Username:     user,
		Password:     password,
		DB:           db,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxIdleConns: maxIdleConns,
		TLSConfig:    &tlsConfig,
	}
	c.client = redis.NewFailoverClient(&opts)
	log.Printf("Connection to the sentinels %v was created \n", strings.Split(url, ","))
	return e
}

func (c *RedisClient) createRedisConnection(url string) error {
	defaultOptions := redis.Options{}

	var e error
	var user string
	if c.datasource.User != "" {
		user = c.datasource.User
	} else {
		user = defaultOptions.Username
	}
	var password string
	if c.datasource.Password != "" {
		password = c.datasource.Password
	} else {
		password = defaultOptions.Password
	}
	readTimeout := defaultOptions.ReadTimeout
	if _, ok := c.datasource.Configuration["readTimeout"]; ok {
		timeout, err := strconv.Atoi(c.datasource.Configuration["readTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	writeTimeout := defaultOptions.WriteTimeout
	if _, ok := c.datasource.Configuration["writeTimeout"]; ok {
		timeout, err := strconv.Atoi(c.datasource.Configuration["writeTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}
	maxIdleConns := defaultOptions.MaxIdleConns
	if _, ok := c.datasource.Configuration["maxIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["maxIdleConns"])
		if err != nil {
			maxIdleConns = c
		}
	}
	poolSize := defaultOptions.PoolSize
	if _, ok := c.datasource.Configuration["poolSize"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["poolSize"])
		if err != nil {
			poolSize = c
		}
	}
	db := defaultOptions.DB
	if _, ok := c.datasource.Configuration["db"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["db"])
		if err != nil {
			db = c
		}
	}
	tls := false
	if _, ok := c.datasource.Configuration["tls-enabled"]; ok {
		c, err := strconv.ParseBool(c.datasource.Configuration["tls-enabled"])
		if err != nil {
			tls = c
		}
	}
	var tlsConfig tls2.Config
	if tls {
		insecureSkipVerify := false
		if _, ok := c.datasource.Configuration["insecureSkipVerify"]; ok {
			c, err := strconv.ParseBool(c.datasource.Configuration["insecureSkipVerify"])
			if err != nil {
				insecureSkipVerify = c
			}
		}
		tlsConfig = tls2.Config{
			InsecureSkipVerify: insecureSkipVerify,
		}
	}

	opts := redis.Options{
		Addr:       url,
		ClientName: "lagoon",
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Printf("Connected to the redis server %v \n", strings.Split(url, ","))
			return nil
		},
		Username:     user,
		Password:     password,
		DB:           db,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxIdleConns: maxIdleConns,
		TLSConfig:    &tlsConfig,
	}
	c.client = redis.NewClient(&opts)
	return e
}

func (c *RedisClient) Close() {
	switch v := c.client.(type) {
	case *redis.Client:
		v.Close()
	case *redis.ClusterClient:
		v.Close()
	}
}

func (c *RedisClient) ListEntryPoints(filter string, entrypointsChannel chan<- datasource.DataBatch, minTreeLevel uint, maxTreeLevel uint) (datasource.ActionStatus, error) {
	// TODO Add list of the channels
	// https://stackoverflow.com/questions/8165188/redis-command-to-get-all-available-channels-for-pub-sub

	var (
		err          error
		actionStatus datasource.ActionStatus
	)

	err = c.client.Ping(context.Background()).Err()
	if err == nil {
		go c.extractEntryPointsWithLevels(err, filter, minTreeLevel, maxTreeLevel, entrypointsChannel)
		actionStatus = datasource.Moved
	}
	return actionStatus, err
}

func (c *RedisClient) extractEntryPointsWithLevels(err error, filter string, minTreeLevel uint, maxTreeLevel uint, entrypointsChannel chan<- datasource.DataBatch) {
	filterTokens := strings.Split(filter, ",")
	scanFilter := filterTokens[0]
	var regexFilter *regexp2.Regexp
	if len(filterTokens) > 1 {
		regexFilter, _ = regexp2.Compile(filterTokens[1])
	}

	scannedKeyCount, entrypoints, err := c.scanAllNodes(scanFilter, regexFilter, minTreeLevel, maxTreeLevel)
	if err != nil {
		log.Printf("ERROR while scanning: %s\n", err.Error())
	} else {
		log.Printf("Number of scanned keys: %d\n", scannedKeyCount)
		var orderedKeys []string
		for e, _ := range entrypoints {
			orderedKeys = append(orderedKeys, e)
		}
		sort.Strings(orderedKeys)
		var valuesToSend []interface{}
		var node *datasource.EntryPointNode
		for _, e := range orderedKeys {
			node = entrypoints[e]
			node.Path = datasource.EntryPoint(e)
			valuesToSend = append(valuesToSend, node)

			// Push messages when valuesToSend is equal to the scan size.
			if int64(len(valuesToSend)) == scanSize {
				c.sendValuesToChannel(valuesToSend, entrypointsChannel)
				valuesToSend = nil
			}
		}
		// After the loop, there might be residual values.
		if len(valuesToSend) > 0 {
			c.sendValuesToChannel(valuesToSend, entrypointsChannel)
		}
	}

	close(entrypointsChannel)
}

func (c *RedisClient) scanAllNodes(scanFilter string, regexFilter *regexp2.Regexp, minTreeLevel uint, maxTreeLevel uint) (int, map[string]*datasource.EntryPointNode, error) {
	var (
		err             error
		scannedKeyCount int
	)

	entrypoints := make(map[string]*datasource.EntryPointNode)
	switch client := c.client.(type) {
	case *redis.ClusterClient:
		mutex := sync.Mutex{}
		loopError := client.ForEachMaster(context.Background(), func(ctx context.Context, node *redis.Client) error {
			id := node.Do(ctx, "cluster", "myid").Val()
			log.Printf("Scanning keys on master node %+v\n", id)
			count, err := c.scanOneNode(node, false, scanFilter, regexFilter, minTreeLevel, maxTreeLevel, entrypoints, func() { mutex.Lock() }, func() { mutex.Unlock() })
			scannedKeyCount = scannedKeyCount + count
			return err
		})

		if loopError != nil {
			err = loopError
		}
	default:
		scannedKeyCount, err = c.scanOneNode(c.client, false, scanFilter, regexFilter, minTreeLevel, maxTreeLevel, entrypoints, func() {}, func() {})
	}
	return scannedKeyCount, entrypoints, err
}

func (c *RedisClient) scanOneNode(scanningRedisClient redis.Cmdable, validateOwnership bool, scanFilter string, regexFilter *regexp2.Regexp, minTreeLevel uint, maxTreeLevel uint, entrypoints map[string]*datasource.EntryPointNode, acquireMutex func(), releaseMutex func()) (int, error) {
	var (
		cursor          uint64
		keys            []string
		entrypoint      string
		err             error
		scannedKeyCount int
	)
	excludedKeys := make(map[string]bool)

	for err == nil {
		keys, cursor, err = scanningRedisClient.Scan(context.Background(), cursor, scanFilter, scanSize).Result()
		if err == nil {
			scannedKeyCount = scannedKeyCount + len(keys)
			for _, key := range keys {
				if regexFilter != nil && !regexFilter.Match([]byte(key)) {
					excludedKeys[key] = true
					continue
				}

				if validateOwnership {
					// If the node belongs to a cluster, we validate the key exists and ignore it otherwise.
					nodeType, err := scanningRedisClient.Type(context.Background(), key).Result()
					if "none" == strings.ToLower(nodeType) || err != nil {
						continue
					}
				}

				tokenCount, tokens := split(key)
				if tokenCount > minTreeLevel {
					entrypoint = ""
					// Complete path and save the number of children
					acquireMutex()

					// Create the entrypoint prefix containing the ignored levels of trees.
					entryPointPrefix := ""
					if minTreeLevel > 0 && minTreeLevel < tokenCount {
						for level := uint(0); level < minTreeLevel; level++ {
							if entryPointPrefix == "" {
								entryPointPrefix = tokens[level]
							} else {
								entryPointPrefix += pathSeparatorAsString + tokens[level]
							}
						}
						entryPointPrefix += pathSeparatorAsString
					}

					for level := minTreeLevel; level <= maxTreeLevel && level < tokenCount; level++ {
						if entrypoint == "" {
							entrypoint = tokens[level]
						} else {
							entrypoint += pathSeparatorAsString + tokens[level]
						}
						existingNode, exists := entrypoints[entrypoint]

						if level < tokenCount-1 {
							if exists {
								existingNode.Length = existingNode.Length + 1
							} else {
								parentHasContent := false
								if regexFilter != nil {
									_, parentHasContent = excludedKeys[entryPointPrefix+entrypoint]
								}
								entrypoints[entrypoint] = &datasource.EntryPointNode{
									Length:     1,
									HasContent: parentHasContent,
									Path:       datasource.EntryPoint(entrypoint),
								}
							}
						} else {
							if exists {
								existingNode.HasContent = true
							} else {
								entrypoints[entrypoint] = &datasource.EntryPointNode{
									Length:     0,
									HasContent: true,
									Path:       datasource.EntryPoint(entrypoint),
								}
							}
						}
					}
					releaseMutex()
				}
			}

			// End of the scanning.
			if cursor == 0 {
				break
			}
		}
	}
	return scannedKeyCount, err
}

func (c *RedisClient) scan(filter string, dataChannel chan<- datasource.DataBatch, scanFn func(cursor uint64, match string, count int64) *redis.ScanCmd, formatFn func(values []string) interface{}) (datasource.ActionStatus, error) {
	var (
		err          error
		actionStatus datasource.ActionStatus
	)

	err = c.client.Ping(context.Background()).Err()
	if err != nil {
		return actionStatus, err
	} else {
		var cursor uint64

		var values []string
		var err error

		// Read first "pages" until the channel is full.
		for len(dataChannel) < cap(dataChannel) {
			values, cursor, err = scanFn(cursor, filter, scanSize).Result()
			if err == nil {
				c.sendValuesToChannel(formatFn(values), dataChannel)
				if cursor == 0 {
					actionStatus = datasource.Completed
					break
				}
			} else {
				return actionStatus, err
			}
		}

		if cursor > 0 {
			// There are more pages to read, the result will be got using a web-socket.
			actionStatus = datasource.Moved

			go func() {
				for cursor != 0 {
					values, cursor, err = scanFn(cursor, filter, scanSize).Result()
					c.sendValuesToChannel(formatFn(values), dataChannel)

					if err != nil {
						log.Printf("ERROR: %s\n", err.Error())
						cursor = 0
					} else {
						log.Printf("Cursor: %d, values length: %d\n", cursor, len(values))

					}
				}

				close(dataChannel)
				log.Println("Leaving the reading routine")
			}()
		}

	}
	return actionStatus, err
}

func (c *RedisClient) fullScan(filter string, scanFn func(cursor uint64, match string, count int64) *redis.ScanCmd, appendFn func([]interface{}, []string) []interface{}) (datasource.DataBatch, error) {
	var (
		err    error
		result datasource.DataBatch
	)

	err = c.client.Ping(context.Background()).Err()
	if err == nil {
		var cursor uint64

		var allValues []interface{}
		var values []string
		var err error

		// Read first "pages" until the channel is full.
		scanned := false
		for !scanned || cursor != 0 {
			values, cursor, err = scanFn(cursor, filter, scanSize).Result()
			if err == nil {
				allValues = appendFn(allValues, values)
			} else {
				return result, err
			}
			scanned = true
		}
		result.Data = allValues
		result.Size = uint64(len(allValues))
	}
	return result, err
}

func (c *RedisClient) sendValuesToChannel(values interface{}, target chan<- datasource.DataBatch) {
	var data []interface{}
	var stringData []string
	var ok bool

	rt := reflect.ValueOf(values)

	switch rt.Kind() {
	case reflect.Slice:
		data, ok = values.([]interface{})
		if !ok {
			stringData, ok = values.([]string)
			if ok {
				for _, value := range stringData {
					data = append(data, value)
				}
			} else {
				data = append(data, values)
			}
		}
	default:
		data = append(data, values)
	}

	if len(data) > 0 {
		target <- datasource.DataBatch{
			Size: uint64(len(data)),
			Data: data,
		}
	}
}

func (c *RedisClient) GetEntryPointInfos(entryPointValue datasource.EntryPoint) (datasource.EntryPointInfos, error) {
	key := string(entryPointValue)

	var (
		keyType string
		err     error
	)

	keyType, err = c.client.Type(context.Background(), key).Result()
	var infos datasource.EntryPointInfos

	if err == nil {
		var result datasource.EntryPointType
		length := uint64(0)
		timeToLive := time.Duration(-1)
		t := strings.ToLower(keyType)
		switch t {
		case "string":
			result = datasource.Value
			length = uint64(c.client.StrLen(context.Background(), key).Val())
			timeToLive = c.client.PTTL(context.Background(), key).Val()
		case "set":
			result = datasource.Set
			length = uint64(c.client.SCard(context.Background(), key).Val())
			timeToLive = c.client.PTTL(context.Background(), key).Val()
		case "zset":
			result = datasource.ScoredSet
			length = uint64(c.client.ZCard(context.Background(), key).Val())
			timeToLive = c.client.PTTL(context.Background(), key).Val()
		case "list":
			result = datasource.List
			length = uint64(c.client.LLen(context.Background(), key).Val())
			timeToLive = c.client.PTTL(context.Background(), key).Val()
		case "hash":
			result = datasource.Hash
			length = uint64(c.client.HLen(context.Background(), key).Val())
			timeToLive = c.client.PTTL(context.Background(), key).Val()
		case "stream":
			result = datasource.Stream
			length = uint64(c.client.XLen(context.Background(), key).Val())
		case "none":
			err = errors.New(fmt.Sprintf("Entrypoint %s was not found", entryPointValue))
		default:
			err = errors.New(fmt.Sprintf("Type %s is unsupported", t))
		}
		infos = datasource.EntryPointInfos{
			Type:       result,
			Length:     length,
			TimeToLive: timeToLive,
		}
	}

	return infos, err
}

func (c *RedisClient) DeleteEntrypoint(entryPointValue datasource.EntryPoint) error {
	if c.datasource.ReadOnly {
		return errors.New("the data source can be only read")
	}
	return c.client.Del(context.Background(), string(entryPointValue)).Err()
}

func (c *RedisClient) DeleteEntrypointChildren(entryPointValue datasource.EntryPoint, errorChannel chan<- error) (datasource.ActionStatus, error) {
	var (
		err          error
		actionStatus datasource.ActionStatus
	)
	if c.datasource.ReadOnly {
		return actionStatus, errors.New("the data source can be only read")
	}

	scanFilter := string(entryPointValue) + ":*"

	err = c.client.Ping(context.Background()).Err()
	if err != nil {
		return actionStatus, err
	} else {
		go func() {
			defer close(errorChannel)

			_, entrypoints, err := c.scanAllNodes(scanFilter, nil, 0, datasource.MaxLevel)
			if err != nil {
				errorChannel <- err
			}
			// Exclude the parent endpoint which should have been added.
			delete(entrypoints, string(entryPointValue))

			total := int64(0)
			keys := []string{}
			for k, v := range entrypoints {
				if v.HasContent {
					keys = append(keys, k)
				}
			}
			log.Printf("%d entries have to be deleted\n", len(keys))

			switch client := c.client.(type) {
			case *redis.ClusterClient:
				// On a cluster, keys have to be deleted one by one, or by groups only if all the elements of the group belongs to the same slot.
				for _, k := range keys {
					log.Printf("Deleting %s...\n", k)
					count, err := client.Del(context.Background(), k).Result()
					total = total + count
					if err != nil {
						log.Printf("ERROR while deleting %s: %s\n", k, err.Error())
						errorChannel <- err
					} else if count > 1 && total%scanSize == 0 {
						log.Printf("%d entries were deleted so far\n", total)
					}
				}
			default:
				total, err = client.Del(context.Background(), keys...).Result()
				if err != nil {
					log.Printf("ERROR while deleting keys %s: %s\n", scanFilter, err.Error())
					errorChannel <- err
				}
			}
			log.Printf("A total of %d entries were deleted\n", total)
		}()
	}
	return datasource.Moved, err
}

func (c *RedisClient) GetContent(entryPointValue datasource.EntryPoint, filter string, contentChannel chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	var (
		err    error
		result datasource.DataBatch
	)

	key := string(entryPointValue)
	statusCmd := c.client.Type(context.Background(), key)
	err = statusCmd.Err()

	if err == nil {
		t := strings.ToLower(statusCmd.Val())
		switch t {
		case "string":
			value, err := c.getValue(entryPointValue)
			if err == nil {
				result = datasource.DataBatch{
					Size: 1,
					Data: []interface{}{value},
				}
			}
		case "set":
			result, err = c.getSetValues(entryPointValue, filter)
		case "zset":
			result, err = c.getZSetValues(entryPointValue, filter)
		case "list":
			result, err = c.getListValues(entryPointValue, filter)
		case "hash":
			result, err = c.getFullHash(entryPointValue, filter)
		case "stream":
			// TODO
			err = errors.New(fmt.Sprintf("Type %s is unsupported", t))
		case "none":
			err = errors.New(fmt.Sprintf("Entrypoint %s was not found", entryPointValue))
		default:
			err = errors.New(fmt.Sprintf("Type %s is unsupported", t))
		}
	}
	if err == nil {
		contentChannel <- result
	}

	return datasource.Completed, err
}

func (c *RedisClient) getValue(entryPointValue datasource.EntryPoint) (datasource.SingleValue, error) {
	key := string(entryPointValue)
	result := c.client.Get(context.Background(), key)
	return result.Val(), result.Err()
}

func (c *RedisClient) getSetValues(entryPointValue datasource.EntryPoint, filter string) (datasource.DataBatch, error) {
	return c.fullScan(filter, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.SScan(context.Background(), string(entryPointValue), cursor, match, count)
	}, func(allValues []interface{}, values []string) (result []interface{}) {
		result = allValues
		for _, v := range values {
			result = append(result, v)
		}
		// Sort the values.
		sort.Slice(result, func(i, j int) bool {
			return result[i].(string) < result[j].(string)
		})
		return
	})
}

func (c *RedisClient) getZSetValues(entryPointValue datasource.EntryPoint, filter string) (datasource.DataBatch, error) {
	return c.fullScan(filter, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.ZScan(context.Background(), string(entryPointValue), cursor, match, count)
	}, func(allValues []interface{}, values []string) (result []interface{}) {
		scoredValuesMap := make(map[float64]SortedSetValues)

		var previousValue SortedSetValues
		for _, val := range allValues {
			previousValue = val.(SortedSetValues)
			scoredValuesMap[previousValue.Score] = previousValue
		}

		// Result slice contains a sequence of "value score value score...".
		// First group the values by score.
		for i := 0; i < len(values); i = i + 2 {
			score, err := strconv.ParseFloat(values[i+1], 64)
			if err == nil {
				scoredValues, exists := scoredValuesMap[score]
				if exists {
					scoredValues.Values = append(scoredValues.Values, values[i])
				} else {
					scoredValues = SortedSetValues{Score: score, Values: []string{values[i]}}
				}
				scoredValuesMap[score] = scoredValues
			}
		}

		for _, value := range scoredValuesMap {
			// Sort the values for each score.
			sort.Strings(value.Values)
			result = append(result, value)
		}
		// Sort the total results by score.
		sort.Slice(result, func(i, j int) bool {
			return result[i].(SortedSetValues).Score < result[j].(SortedSetValues).Score
		})
		return
	})
}

func (c *RedisClient) getListValues(entryPointValue datasource.EntryPoint, filter string) (datasource.DataBatch, error) {
	var result datasource.DataBatch
	values, err := c.client.LRange(context.Background(), string(entryPointValue), 0, -1).Result()
	if err == nil {
		if len(values) > 0 {
			var regexp *regexp2.Regexp
			if filter != "" {
				regexp = regexp2.MustCompile(strings.ReplaceAll(filter, "*", ".*"))
			}
			var sendableValues []interface{}
			for _, value := range values {
				if regexp == nil || regexp.Match([]byte(value)) {
					sendableValues = append(sendableValues, value)
				}
			}
			result = datasource.DataBatch{
				Size: uint64(len(sendableValues)),
				Data: sendableValues,
			}
		}
	}
	return result, err
}

func (c *RedisClient) getFullHash(entryPointValue datasource.EntryPoint, filter string) (datasource.DataBatch, error) {
	return c.fullScan(filter, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.HScan(context.Background(), string(entryPointValue), cursor, match, count)
	}, func(allValues []interface{}, values []string) (result []interface{}) {
		result = allValues
		for i := 0; i < len(values); i = i + 2 {
			result = append(result, HashValue{values[i], values[i+1]})
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i].(HashValue).Key < result[j].(HashValue).Key
		})
		return
	})
}

func (c *RedisClient) getStream(entryPointValue datasource.EntryPoint, filter string, target chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	messages, err := c.client.XRange(context.Background(), string(entryPointValue), "-", "+").Result()
	if err == nil {
		if len(messages) > 0 {
			dataBatch := datasource.DataBatch{
				Size: uint64(len(target)),
			}
			for _, message := range messages {
				dataBatch.Data = append(dataBatch.Data, message)
			}
			target <- dataBatch
		}
		// End of the stream.
		target <- datasource.DataBatch{
			Size: 0,
		}
		return datasource.Completed, nil
	}
	return datasource.None, err
}

func (c *RedisClient) Consume(entryPointValue datasource.EntryPoint, target chan<- datasource.DataBatch, filter datasource.Filter, fromBeginning bool) (datasource.ActionStatus, error) {

	panic("Implement me!")

	/*
		streamCmd := c.datasource.XReadStreams(string(entryPointValue))
		c.datasource.XRead()
		var status ActionStatus
		err := streamCmd.Err()
		if err == nil {
			go func() {
				for c.activeStreams {
					// TODO
				}
			}()
		}
		return status, err
	*/
}

func (c *RedisClient) ExecuteCommand(args []interface{}, nodeID string) (interface{}, error) {
	if len(args) > 0 && c.datasource.ReadOnly && !c.isClusterReadonlyCommand(args) {
		cmd, ok := args[0].(string)
		if ok {
			c.initReadonlyCommands()
			if !c.isReadOnlyCommand(cmd) {
				return nil, errors.New(fmt.Sprintf("the data source %s can only be read", c.datasource.Id))
			}
		}
	}

	cmd := redis.NewCmd(context.Background(), args...)
	c.processCmd(cmd, nodeID)
	return cmd.Result()
}

func (c *RedisClient) isClusterReadonlyCommand(args []interface{}) bool {
	if len(args) >= 2 {
		cmd, ok := args[0].(string)
		if ok && strings.ToLower(cmd) == "cluster" {
			cmd, ok = args[1].(string)
			cmd = strings.ToLower(cmd)
			return ok && (cmd == "info" || cmd == "getkeysinslot" || cmd == "keyslot" || cmd == "myid" || cmd == "nodes" || cmd == "replicas" || cmd == "slaves" || cmd == "slots")
		}
	}
	return false
}

func (c *RedisClient) isReadOnlyCommand(cmd string) bool {
	lowerCmd := strings.ToLower(cmd)
	for i := range c.readOnlyCommands {
		if c.readOnlyCommands[i] == lowerCmd {
			return true
		}
	}
	return false
}

func (c *RedisClient) processCmd(cmd redis.Cmder, nodeID string) {
	switch v := c.client.(type) {
	case *redis.Client:
		v.Process(context.Background(), cmd)
	case *redis.ClusterClient:
		if nodeID == "" {
			v.Process(context.Background(), cmd)
		} else {
			v.ForEachShard(context.Background(), func(ctx context.Context, client *redis.Client) error {
				myIDResult, err := client.Do(ctx, "cluster", "myid").Result()
				if err == nil {
					myID := (myIDResult).(string)
					if myID == nodeID {
						client.Process(ctx, cmd)
						return err
					}
				}
				return err
			})
		}
	}
}

package redis

// https://github.com/go-redis/redis/blob/master/example_test.go
import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
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

func init() {
	datasource.DeclareImplementation(&RedisVendor{})
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

func (c *RedisClient) GetInfos() (interface{}, error) {
	switch c.client.(type) {
	case *redis.ClusterClient:
		return c.client.ClusterInfo().Result()
	default:
		return c.client.Info().Result()
	}
}

func (c *RedisClient) GetStatus() (interface{}, error) {
	return c.GetInfos()
}

func (c *RedisClient) Open() error {
	err := c.createConnection()
	if err == nil {
		pong, err := c.client.Ping().Result()
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
		for _, v := range c.client.Command().Val() {
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
	maxConnAge := defaultOptions.MaxConnAge
	if _, ok := c.datasource.Configuration["maxConnAge"]; ok {
		age, err := strconv.Atoi(c.datasource.Configuration["maxConnAge"])
		if err == nil {
			maxConnAge = time.Duration(age) * time.Minute
		}
	}
	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}

	opts := redis.ClusterOptions{
		Addrs:         strings.Split(url, ","),
		Password:      password,
		ReadTimeout:   readTimeout,
		WriteTimeout:  writeTimeout,
		MaxConnAge:    maxConnAge,
		MinIdleConns:  minIdleConns,
		ReadOnly:      false,
		RouteRandomly: false,
		OnConnect: func(conn *redis.Conn) error {
			log.Printf("Connected to the cluster %v \n", strings.Split(url, ","))
			return nil
		},
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

	maxConnAge := defaultOptions.MaxConnAge
	if _, ok := c.datasource.Configuration["maxConnAge"]; ok {
		age, err := strconv.Atoi(c.datasource.Configuration["maxConnAge"])
		if err == nil {
			maxConnAge = time.Duration(age) * time.Minute
		}
	}

	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}

	opts := redis.FailoverOptions{
		MasterName:    c.datasource.Configuration["master"],
		SentinelAddrs: strings.Split(url, ","),
		Password:      password,
		ReadTimeout:   readTimeout,
		WriteTimeout:  writeTimeout,
		MaxConnAge:    maxConnAge,
		MinIdleConns:  minIdleConns,
		OnConnect: func(conn *redis.Conn) error {
			log.Printf("Connected to the sentinels %v \n", strings.Split(url, ","))
			return nil
		},
	}
	c.client = redis.NewFailoverClient(&opts)
	log.Printf("Connection to the sentinels %v was created \n", strings.Split(url, ","))
	return e
}

func (c *RedisClient) createRedisConnection(url string) error {
	defaultOptions := redis.Options{}

	var e error
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
	maxConnAge := defaultOptions.MaxConnAge
	if _, ok := c.datasource.Configuration["maxConnAge"]; ok {
		age, err := strconv.Atoi(c.datasource.Configuration["maxConnAge"])
		if err == nil {
			maxConnAge = time.Duration(age) * time.Minute
		}
	}
	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}
	opts := redis.Options{
		Addr:         url,
		Password:     password,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		MaxConnAge:   maxConnAge,
		MinIdleConns: minIdleConns,
		OnConnect: func(conn *redis.Conn) error {
			log.Printf("Connected to the redis server %v \n", strings.Split(url, ","))
			return nil
		},
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

	err = c.client.Ping().Err()
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
				valuesToSend = valuesToSend[:0]
			}
		}
		// After the loop, there might be residual values.
		c.sendValuesToChannel(valuesToSend, entrypointsChannel)
	}

	// End of the stream.
	entrypointsChannel <- datasource.DataBatch{}
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
		loopError := client.ForEachNode(func(node *redis.Client) error {
			roleResult, err := node.Do("role").Result()
			if err == nil {
				role := (roleResult.([]interface{})[0]).(string)
				if "master" == strings.ToLower(role) {
					id := node.Do("cluster", "myid").Val()
					log.Printf("Scanning keys on master node %+v\n", id)
					count, err := c.scanOneNode(node, false, scanFilter, regexFilter, minTreeLevel, maxTreeLevel, entrypoints, func() { mutex.Lock() }, func() { mutex.Unlock() })
					scannedKeyCount = scannedKeyCount + count
					return err
				}
			}
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
		tokens          []string
		entrypoint      string
		err             error
		scannedKeyCount int
	)

	excludedKeys := make(map[string]bool)

	for err == nil {
		keys, cursor, err = scanningRedisClient.Scan(cursor, scanFilter, scanSize).Result()
		if err == nil {
			scannedKeyCount = scannedKeyCount + len(keys)
			for _, key := range keys {
				if regexFilter != nil && !regexFilter.Match([]byte(key)) {
					excludedKeys[key] = true
					continue
				}

				if validateOwnership {
					// If the node belongs to a cluster, we validate the key exists and ignore it otherwise.
					nodeType, err := scanningRedisClient.Type(key).Result()
					if "none" == strings.ToLower(nodeType) || err != nil {
						continue
					}
				}

				tokens = strings.Split(key, ":")
				if uint(len(tokens)) > minTreeLevel {
					entrypoint = ""
					// Complete path and save the number of children
					acquireMutex()

					// Create the entrypoint prefix containing the ignored levels of trees.
					entryPointPrefix := ""
					if minTreeLevel > 0 && minTreeLevel < uint(len(tokens)) {
						for level := uint(0); level < minTreeLevel; level++ {
							if entryPointPrefix == "" {
								entryPointPrefix = tokens[level]
							} else {
								entryPointPrefix += ":" + tokens[level]
							}
						}
						entryPointPrefix += ":"
					}

					for level := minTreeLevel; level <= maxTreeLevel && level < uint(len(tokens)); level++ {
						if entrypoint == "" {
							entrypoint = tokens[level]
						} else {
							entrypoint += ":" + tokens[level]
						}
						existingNode, exists := entrypoints[entrypoint]

						if level < uint(len(tokens)-1) {
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

	err = c.client.Ping().Err()
	if err != nil {
		return actionStatus, err
	} else {
		var cursor uint64

		var keys []string
		var err error

		// Read first "pages" until the channel is full.
		for i := 0; i < cap(dataChannel); i++ {
			keys, cursor, err = scanFn(cursor, filter, scanSize).Result()
			if err == nil {
				c.sendValuesToChannel(formatFn(keys), dataChannel)
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
					keys, cursor, err = scanFn(cursor, filter, scanSize).Result()
					c.sendValuesToChannel(formatFn(keys), dataChannel)

					if err != nil {
						log.Printf("ERROR: %s\n", err.Error())
						cursor = 0
					} else {
						log.Printf("Cursor: %d, keys length: %d\n", cursor, len(keys))

					}
				}

				// End of the stream.
				dataChannel <- datasource.DataBatch{}
				close(dataChannel)
				log.Println("Leaving the reading routine")
			}()
		}

	}
	return actionStatus, err
}

// identityFormat is a formatter function which returns the exact received values.
func (c *RedisClient) identityFormat(values []string) interface{} {
	return values
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

	keyType, err = c.client.Type(key).Result()
	var infos datasource.EntryPointInfos

	if err == nil {
		var result datasource.EntryPointType
		length := uint64(0)
		timeToLive := time.Duration(-1)
		t := strings.ToLower(keyType)
		switch t {
		case "string":
			result = datasource.Value
			length = uint64(c.client.StrLen(key).Val())
			timeToLive = c.client.PTTL(key).Val()
		case "set":
			result = datasource.Set
			length = uint64(c.client.SCard(key).Val())
			timeToLive = c.client.PTTL(key).Val()
		case "zset":
			result = datasource.ScoredSet
			length = uint64(c.client.ZCard(key).Val())
			timeToLive = c.client.PTTL(key).Val()
		case "list":
			result = datasource.List
			length = uint64(c.client.LLen(key).Val())
			timeToLive = c.client.PTTL(key).Val()
		case "hash":
			result = datasource.Hash
			length = uint64(c.client.HLen(key).Val())
			timeToLive = c.client.PTTL(key).Val()
		case "stream":
			result = datasource.Stream
			length = uint64(c.client.XLen(key).Val())
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
	return c.client.Del(string(entryPointValue)).Err()
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

	err = c.client.Ping().Err()
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
					count, err := client.Del(k).Result()
					total = total + count
					if err != nil {
						log.Printf("ERROR while deleting %s: %s\n", k, err.Error())
						errorChannel <- err
					} else if count > 1 && total%scanSize == 0 {
						log.Printf("%d entries were deleted so far\n", total)
					}
				}
			default:
				total, err = client.Del(keys...).Result()
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

func (c *RedisClient) GetContent(entryPointValue datasource.EntryPoint, filter string, content chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	var (
		err          error
		actionStatus datasource.ActionStatus
	)

	key := string(entryPointValue)
	statusCmd := c.client.Type(key)
	err = statusCmd.Err()

	if err == nil {
		t := strings.ToLower(statusCmd.Val())
		switch t {
		case "string":
			value, err := c.getValue(entryPointValue)
			if err == nil {
				content <- datasource.DataBatch{
					Size: 1,
					Data: []interface{}{value},
				}
			}
			actionStatus = datasource.Completed
		case "set":
			return c.getSetValues(entryPointValue, filter, content)
		case "zset":
			return c.getZSetValues(entryPointValue, filter, content)
		case "list":
			return c.getListValues(entryPointValue, filter, content)
		case "hash":
			return c.getFullHash(entryPointValue, filter, content)
		case "stream":
			// TODO
		case "none":
			err = errors.New(fmt.Sprintf("Entrypoint %s was not found", entryPointValue))
		default:
			err = errors.New(fmt.Sprintf("Type %s is unsupported", t))
		}
	}

	return actionStatus, err
}

func (c *RedisClient) getValue(entryPointValue datasource.EntryPoint) (datasource.SingleValue, error) {
	key := string(entryPointValue)
	result := c.client.Get(key)
	return result.Val(), result.Err()
}

func (c *RedisClient) getSetValues(entryPointValue datasource.EntryPoint, filter string, target chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	return c.scan(filter, target, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.SScan(string(entryPointValue), cursor, match, count)
	}, c.identityFormat)
}

func (c *RedisClient) getZSetValues(entryPointValue datasource.EntryPoint, filter string, target chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	return c.scan(filter, target, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.ZScan(string(entryPointValue), cursor, match, count)
	}, func(values []string) interface{} {
		result := make(map[float64][]string)
		for i := 0; i < len(values); i = i + 2 {
			score, err := strconv.ParseFloat(values[i+1], 64)
			if err == nil {
				scoredValues, ok := result[score]
				if ok {
					scoredValues = append(scoredValues, values[i])
				} else {
					scoredValues = []string{values[i]}
				}
				result[score] = scoredValues
			}
		}
		return result
	})
}

func (c *RedisClient) getListValues(entryPointValue datasource.EntryPoint, filter string, target chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	values, err := c.client.LRange(string(entryPointValue), 0, -1).Result()
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
			target <- datasource.DataBatch{
				Size: uint64(len(sendableValues)),
				Data: sendableValues,
			}
		}

		// End of the stream.
		target <- datasource.DataBatch{
			Size: 0,
		}
		return datasource.Completed, nil
	}
	return datasource.None, err
}

func (c *RedisClient) getFullHash(entryPointValue datasource.EntryPoint, filter string, target chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	return c.scan(filter, target, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.HScan(string(entryPointValue), cursor, match, count)
	}, func(values []string) interface{} {
		result := make(map[string]string)
		for i := 0; i < len(values); i = i + 2 {
			result[values[i]] = values[i+1]
		}
		return result
	})
}

func (c *RedisClient) getStream(entryPointValue datasource.EntryPoint, filter string, target chan<- datasource.DataBatch) (datasource.ActionStatus, error) {
	messages, err := c.client.XRange(string(entryPointValue), "-", "+").Result()
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

	cmd := redis.NewCmd(args...)
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
		v.Process(cmd)
	case *redis.ClusterClient:
		if nodeID == "" {
			v.Process(cmd)
		} else {
			v.ForEachNode(func(client *redis.Client) error {
				myIDResult, err := client.Do("CLUSTER", "MYID").Result()
				if err == nil {
					myID := (myIDResult).(string)
					if myID == nodeID {
						client.Process(cmd)
						return err
					}
				}
				return err
			})
		}
	}
}

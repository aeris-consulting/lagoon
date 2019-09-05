package datasource

// https://github.com/go-redis/redis/blob/master/example_test.go
import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
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
	Datasource    *DataSource
	client        redis.Cmdable
	activeStreams bool
}

func (c *RedisClient) GetSupportedTypes() []EntryPointType {
	return []EntryPointType{Value, Set, SortedSet, List, Hash, Stream}
}

func (c *RedisClient) Open() error {
	err := c.createConnection()
	if err == nil {
		pong, err := c.client.Ping().Result()
		if err == nil {
			log.Printf("Connection status of ping: %v\n", pong)
			c.activeStreams = true
		} else {
			log.Printf("ERROR When pinging: %s\n", err.Error())
		}
	}
	return err
}

func (c *RedisClient) createConnection() error {
	var bootstrap = c.Datasource.Bootstrap
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
	if c.Datasource.Password != "" {
		password = c.Datasource.Password
	} else {
		password = defaultOptions.Password
	}
	readTimeout := defaultOptions.ReadTimeout
	if _, ok := c.Datasource.Configuration["readTimeout"]; ok {
		timeout, err := strconv.Atoi(c.Datasource.Configuration["readTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	writeTimeout := defaultOptions.WriteTimeout
	if _, ok := c.Datasource.Configuration["writeTimeout"]; ok {
		timeout, err := strconv.Atoi(c.Datasource.Configuration["writeTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	maxConnAge := defaultOptions.MaxConnAge
	if _, ok := c.Datasource.Configuration["maxConnAge"]; ok {
		age, err := strconv.Atoi(c.Datasource.Configuration["maxConnAge"])
		if err == nil {
			maxConnAge = time.Duration(age) * time.Minute
		}
	}
	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.Datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.Datasource.Configuration["minIdleConns"])
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
	var sentinelPassword string
	var password string
	var ok bool

	if c.Datasource.Password != "" {
		password = c.Datasource.Password
	} else {
		password = defaultOptions.Password
	}

	if sentinelPassword, ok = c.Datasource.Configuration["sentinelPassword"]; !ok {
		sentinelPassword = defaultOptions.SentinelPassword
	}

	readTimeout := defaultOptions.ReadTimeout
	if _, ok := c.Datasource.Configuration["readTimeout"]; ok {
		timeout, err := strconv.Atoi(c.Datasource.Configuration["readTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}

	writeTimeout := defaultOptions.WriteTimeout
	if _, ok := c.Datasource.Configuration["writeTimeout"]; ok {
		timeout, err := strconv.Atoi(c.Datasource.Configuration["writeTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}

	maxConnAge := defaultOptions.MaxConnAge
	if _, ok := c.Datasource.Configuration["maxConnAge"]; ok {
		age, err := strconv.Atoi(c.Datasource.Configuration["maxConnAge"])
		if err == nil {
			maxConnAge = time.Duration(age) * time.Minute
		}
	}

	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.Datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.Datasource.Configuration["minIdleConns"])
		if err != nil {
			minIdleConns = c
		}
	}

	opts := redis.FailoverOptions{
		MasterName:       c.Datasource.Configuration["master"],
		SentinelAddrs:    strings.Split(url, ","),
		SentinelPassword: sentinelPassword,
		Password:         password,
		ReadTimeout:      readTimeout,
		WriteTimeout:     writeTimeout,
		MaxConnAge:       maxConnAge,
		MinIdleConns:     minIdleConns,
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
	if c.Datasource.Password != "" {
		password = c.Datasource.Password
	} else {
		password = defaultOptions.Password
	}
	readTimeout := defaultOptions.ReadTimeout
	if _, ok := c.Datasource.Configuration["readTimeout"]; ok {
		timeout, err := strconv.Atoi(c.Datasource.Configuration["readTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	writeTimeout := defaultOptions.WriteTimeout
	if _, ok := c.Datasource.Configuration["writeTimeout"]; ok {
		timeout, err := strconv.Atoi(c.Datasource.Configuration["writeTimeout"])
		if err == nil {
			readTimeout = time.Duration(timeout) * time.Second
		}
	}
	maxConnAge := defaultOptions.MaxConnAge
	if _, ok := c.Datasource.Configuration["maxConnAge"]; ok {
		age, err := strconv.Atoi(c.Datasource.Configuration["maxConnAge"])
		if err == nil {
			maxConnAge = time.Duration(age) * time.Minute
		}
	}
	minIdleConns := defaultOptions.MinIdleConns
	if _, ok := c.Datasource.Configuration["minIdleConns"]; ok {
		c, err := strconv.Atoi(c.Datasource.Configuration["minIdleConns"])
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

func (c *RedisClient) ListEntryPoints(filter string, entrypointsChannel chan<- DataBatch, minTreeLevel uint, maxTreeLevel uint) (ActionStatus, error) {
	// TODO Add list of the channels
	// https://stackoverflow.com/questions/8165188/redis-command-to-get-all-available-channels-for-pub-sub

	var (
		err          error
		actionStatus ActionStatus
	)

	err = c.client.Ping().Err()
	if err == nil {
		go c.extractEntryPointsWithLevels(err, filter, minTreeLevel, maxTreeLevel, entrypointsChannel)
		actionStatus = Moved
	}
	return actionStatus, err
}

func (c *RedisClient) extractEntryPointsWithLevels(err error, filter string, minTreeLevel uint, maxTreeLevel uint, entrypointsChannel chan<- DataBatch) {
	var entrypoints = make(map[string]*EntryPointNode)

	filterTokens := strings.Split(filter, ",")
	scanFilter := filterTokens[0]
	var regexFilter *regexp2.Regexp
	if len(filterTokens) > 1 {
		regexFilter, _ = regexp2.Compile(filterTokens[1])
	}

	var scannedKeyCount int
	switch client := c.client.(type) {
	case *redis.ClusterClient:
		mutex := sync.Mutex{}
		loopError := client.ForEachNode(func(client *redis.Client) error {
			roleResult, err := client.Do("ROLE").Result()
			if err == nil {
				role := (roleResult.([]interface{})[0]).(string)
				if "master" == strings.ToLower(role) {
					log.Printf("Scanning keys on %v\n", roleResult)
					err, count := c.scanKeysOnNode(client, scanFilter, regexFilter, minTreeLevel, maxTreeLevel, entrypoints, func() { mutex.Lock() }, func() { mutex.Unlock() })
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
		err, scannedKeyCount = c.scanKeysOnNode(c.client, scanFilter, regexFilter, minTreeLevel, maxTreeLevel, entrypoints, func() {}, func() {})
	}

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
		var node *EntryPointNode
		for _, e := range orderedKeys {
			node = entrypoints[e]
			node.Path = EntryPoint(e)
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
	entrypointsChannel <- DataBatch{}
	close(entrypointsChannel)
}

func (c *RedisClient) scanKeysOnNode(redisClient redis.Cmdable, scanFilter string, regexFilter *regexp2.Regexp, minTreeLevel uint, maxTreeLevel uint, entrypoints map[string]*EntryPointNode, acquireMutex func(), releaseMutex func()) (error, int) {
	var (
		cursor          uint64
		keys            []string
		tokens          []string
		entrypoint      string
		err             error
		scannedKeyCount int
	)

	for err == nil {
		keys, cursor, err = redisClient.Scan(cursor, scanFilter, scanSize).Result()
		if err == nil {
			scannedKeyCount = scannedKeyCount + len(keys)
			for _, key := range keys {
				if regexFilter != nil && !regexFilter.Match([]byte(key)) {
					continue
				}
				tokens = strings.Split(key, ":")
				if uint(len(tokens)) > minTreeLevel {
					entrypoint = ""
					// Complete path and save the number of children
					acquireMutex()
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
								entrypoints[entrypoint] = &EntryPointNode{Length: 1, HasContent: false}
							}
						} else {
							if exists {
								existingNode.HasContent = true
							} else {
								entrypoints[entrypoint] = &EntryPointNode{Length: 0, HasContent: true}
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
	return err, scannedKeyCount
}

func (c *RedisClient) scan(filter string, dataChannel chan<- DataBatch, scanFn func(cursor uint64, match string, count int64) *redis.ScanCmd, formatFn func(values []string) interface{}) (ActionStatus, error) {
	var (
		err          error
		actionStatus ActionStatus
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
					actionStatus = Completed
					break
				}
			} else {
				return actionStatus, err
			}
		}

		if cursor > 0 {
			// There are more pages to read, the result will be got using a web-socket.
			actionStatus = Moved

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
				dataChannel <- DataBatch{}
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

func (c *RedisClient) sendValuesToChannel(values interface{}, target chan<- DataBatch) {
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
		target <- DataBatch{
			Size: uint64(len(data)),
			Data: data,
		}
	}
}

func (c *RedisClient) GetEntryPointInfos(entryPointValue EntryPoint) (EntryPointInfos, error) {
	key := string(entryPointValue)

	var (
		keyType string
		err     error
	)
	keyType, err = c.client.Type(key).Result()
	var infos EntryPointInfos

	if err == nil {
		var (
			result     EntryPointType
			length     uint64
			timeToLive time.Duration
			ttlErr     error
		)

		t := strings.ToLower(keyType)
		switch t {
		case "string":
			result = Value
			length = uint64(c.client.StrLen(key).Val())
		case "set":
			result = Set
			length = uint64(c.client.SCard(key).Val())
		case "zset":
			result = SortedSet
			length = uint64(c.client.ZCard(key).Val())
		case "list":
			result = List
			length = uint64(c.client.LLen(key).Val())
		case "hash":
			result = Hash
			length = uint64(c.client.HLen(key).Val())
		case "stream":
			result = Stream
			length = uint64(c.client.XLen(key).Val())
		case "none":
			err = errors.New(fmt.Sprintf("Entrypoint %s was not found", entryPointValue))
		default:
			err = errors.New(fmt.Sprintf("Type %s is unsupported", t))
		}
		infos = EntryPointInfos{
			Type:   result,
			Length: length,
		}
		timeToLive, ttlErr = c.client.PTTL(key).Result()
		if ttlErr != nil {
			infos.TimeToLive = int64(timeToLive / time.Second)
		} else {
			infos.TimeToLive = -1
		}
	}

	return infos, err
}

func (c *RedisClient) DeleteEntrypoint(entryPointValue EntryPoint) error {
	return c.client.Unlink(string(entryPointValue)).Err()
}

func (c *RedisClient) DeleteEntrypointChidren(entryPointValue EntryPoint, errorChannel chan<- error) (ActionStatus, error) {
	var (
		err          error
		actionStatus ActionStatus
	)

	err = c.client.Ping().Err()
	if err != nil {
		return actionStatus, err
	} else {
		var cursor uint64
		var children []string
		var keys []string
		var err error

		go func() {
			for {
				keys, cursor, err = c.client.Scan(cursor, string(entryPointValue)+":*", scanSize).Result()
				if err != nil {
					children = append(children, keys...)
					if cursor == 0 {
						break
					}
				}
			}
			c.client.Unlink(children...)
			close(errorChannel)
		}()
	}
	return Moved, err
}

func (c *RedisClient) GetContent(entryPointValue EntryPoint, filter string, content chan<- DataBatch) (ActionStatus, error) {
	var (
		err          error
		actionStatus ActionStatus
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
				content <- DataBatch{
					Size: 1,
					Data: []interface{}{value},
				}
			}
			actionStatus = Completed
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

func (c *RedisClient) SetContent(entryPoint EntryPoint, content EntryPointContent) error {
	var (
		entrypointType EntryPointType
		err            error
	)
	for k, v := range EntryPointTypesAsString {
		if content.Type == v {
			entrypointType = k
			break
		}
	}

	key := string(entryPoint)

	switch entrypointType {
	case Value:
		c.client.Set(key, content.Value, time.Second*time.Duration(content.TimeToLive))
	case Set:
		var (
			valuesToRemove []interface{}
			valuesToAdd    []interface{}
		)
		for k, v := range content.Values {
			if k < 0 {
				valuesToRemove = append(valuesToRemove, v)
			} else {
				valuesToAdd = append(valuesToAdd, v)
			}
		}
		c.client.SRem(key, valuesToRemove...)
		c.client.SAdd(key, valuesToAdd...)
	case SortedSet:
		var (
			valuesToRemove []interface{}
			valuesToAdd    []*redis.Z
		)
		for k, v := range content.Values {
			if k < 0 {
				valuesToRemove = append(valuesToRemove, v)
			} else {
				valuesToAdd = append(valuesToAdd, &redis.Z{
					Score:  k,
					Member: v,
				})
			}
		}
		c.client.ZRem(key, valuesToRemove...)
		c.client.ZAddCh(key, valuesToAdd...)
	case List:
		// TODO
	case Hash:
		c.client.Del(key)
		c.client.HMSet(key, content.Hash)
	case Stream:
		// TODO
	default:
		err = errors.New(fmt.Sprintf("Type %s is unsupported", entrypointType))
	}

	return err
}

func (c *RedisClient) getValue(entryPointValue EntryPoint) (SingleValue, error) {
	key := string(entryPointValue)
	result := c.client.Get(key)
	return result.String(), result.Err()
}

func (c *RedisClient) getSetValues(entryPointValue EntryPoint, filter string, target chan<- DataBatch) (ActionStatus, error) {
	return c.scan(filter, target, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.SScan(string(entryPointValue), cursor, match, count)
	}, c.identityFormat)
}

func (c *RedisClient) getZSetValues(entryPointValue EntryPoint, filter string, target chan<- DataBatch) (ActionStatus, error) {
	return c.scan(filter, target, func(cursor uint64, match string, count int64) *redis.ScanCmd {
		return c.client.ZScan(string(entryPointValue), cursor, match, count)
	}, c.identityFormat)
}

func (c *RedisClient) getListValues(entryPointValue EntryPoint, filter string, target chan<- DataBatch) (ActionStatus, error) {
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
			target <- DataBatch{
				Size: 0,
				Data: sendableValues,
			}
		}

		// End of the stream.
		target <- DataBatch{
			Size: 0,
		}
		return Completed, nil
	}
	return None, err
}

func (c *RedisClient) getFullHash(entryPointValue EntryPoint, filter string, target chan<- DataBatch) (ActionStatus, error) {
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

func (c *RedisClient) getStream(entryPointValue EntryPoint, filter string, target chan<- DataBatch) (ActionStatus, error) {
	messages, err := c.client.XRange(string(entryPointValue), "-", "+").Result()
	if err == nil {
		if len(messages) > 0 {
			dataBatch := DataBatch{
				Size: uint64(len(target)),
			}
			for _, message := range messages {
				dataBatch.Data = append(dataBatch.Data, message)
			}
			target <- dataBatch
		}
		// End of the stream.
		target <- DataBatch{
			Size: 0,
		}
		return Completed, nil
	}
	return None, err
}

func (c *RedisClient) Consume(entryPointValue EntryPoint, target chan<- DataBatch, filter Filter, fromBeginning bool) (ActionStatus, error) {

	panic("Implement me!")

	/*
		streamCmd := c.datasource.XReadStreams(string(entryPointValue))
		c.datasource.XReadGroup()
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

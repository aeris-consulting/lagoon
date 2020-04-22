package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"lagoon/datasource"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// https://stackoverflow.com/questions/23729790/how-can-i-do-test-setup-using-the-testing-package-in-go
// https://golang.org/pkg/testing/

var (
	redisIp   string
	redisPort int
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:5-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
			Cmd:          []string{"redis-server"},
		},
		Started: true,
	})
	if err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
	redisIp, err = redisC.Host(ctx)
	if err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
	containerPort, err := redisC.MappedPort(ctx, "6379/tcp")
	if err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
	redisPort, err = strconv.Atoi(containerPort.Port())
	if err != nil {
		logrus.Println(err)
		os.Exit(1)
	}
	defer redisC.Terminate(ctx)

	os.Exit(m.Run())
}

func TestRedisClient_OpenAndCloseWithPassword(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	client.client.ConfigSet("requirepass", "this-is-my-password")

	// when
	_, err = client.ExecuteCommand([]interface{}{"info"}, "")
	assert.NotNil(t, err)

	clientWithPassword := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
			Password:  "this-is-my-password",
		},
	}
	defer func() {
		clientWithPassword.client.ConfigSet("requirepass", "")
		clientWithPassword.Close()
		client.Close()
	}()

	// then
	err = clientWithPassword.Open()
	assert.Nil(t, err)
	result, err := client.ExecuteCommand([]interface{}{"info"}, "")
	// No action is possible any more without password.
	assert.NotNil(t, err)

	// then
	result, err = clientWithPassword.ExecuteCommand([]interface{}{"info"}, "")
	// The action is possible with password.
	assert.Nil(t, err)

	t.Logf("Redis Info: %+v\n", result)
}

func TestRedisClient_OpenAndClose(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	// when
	err := client.Open()
	defer client.Close()

	// then
	assert.Nil(t, err)

	result, err := client.GetInfos()
	assert.Nil(t, err)

	t.Logf("Redis Info: %+v\n", result)
}

func TestRedisClient_ListAllEntryPoints(t *testing.T) {
	// given
	testData := []struct {
		key   string
		count int
	}{
		{key: "group-atom", count: 200},
		{key: "group-atic", count: 200},
		{key: "group-artic", count: 200},
		{key: "group-bolton", count: 200},
	}

	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	for _, d := range testData {
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(s, s, time.Minute)
		}
	}

	// List all the entry points in several parts.
	dataChannel := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.ListEntryPoints("*", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	assert.Equal(t, datasource.Moved, actionStatus)

	actualData := []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 804, len(actualData))

	// List all the entry points at once.
	dataChannel = make(chan datasource.DataBatch, 100)
	actionStatus, err = client.ListEntryPoints("*", dataChannel, 0, 0)
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 4, len(actualData)) // There are four root nodes only.

	// List all the entry points at once.
	dataChannel = make(chan datasource.DataBatch, scanSize)
	actionStatus, err = client.ListEntryPoints("*", dataChannel, 0, 1)
	assert.Nil(t, err)
	assert.Equal(t, datasource.Moved, actionStatus)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 804, len(actualData)) // There are four root nodes plus all the leaves.
}

func TestRedisClient_ListEntryPointsWithOneFilter(t *testing.T) {
	// given
	testData := []struct {
		key   string
		count int
	}{
		{key: "group-atom", count: 200},
		{key: "group-atic", count: 200},
		{key: "group-artic", count: 200},
		{key: "group-bolton", count: 200},
	}

	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	for _, d := range testData {
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(s, s, time.Minute)
		}
	}

	dataChannel := make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-*", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	actualData := []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 804, len(actualData)) // There are four root nodes plus all the leaves.

	dataChannel = make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-at*", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 402, len(actualData)) // There are two root nodes plus all their leaves.

	dataChannel = make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-a*t*", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 603, len(actualData)) // There are three root nodes plus all their leaves.

	dataChannel = make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-a*t*", dataChannel, 0, 0)

	// then
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 3, len(actualData)) // There are three root nodes only.
}

func TestRedisClient_ListEntryPointsWithTwoFilters(t *testing.T) {
	// given
	testData := []struct {
		key   string
		count int
	}{
		{key: "group-atom", count: 200},
		{key: "group-atic", count: 200},
		{key: "group-artic", count: 200},
		{key: "group-bolton", count: 200},
	}

	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	for _, d := range testData {
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(s, s, time.Minute)
		}
	}

	dataChannel := make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-*, *bolt*", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	actualData := []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 201, len(actualData)) // There is one root node plus all its leaves.

	dataChannel = make(chan datasource.DataBatch, 100)
	// when
	_, err = client.ListEntryPoints("group-*, *", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 804, len(actualData)) // There are four root node and their leaves.

	dataChannel = make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-*, *bolt*", dataChannel, 0, 1)

	// then
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 201, len(actualData)) // There is one root node only.

	dataChannel = make(chan datasource.DataBatch, 100)

	// when
	_, err = client.ListEntryPoints("group-*, *bolt*", dataChannel, 0, 0)

	// then
	assert.Nil(t, err)
	actualData = []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	assert.Equal(t, 1, len(actualData)) // There is one root node only.
}

func TestRedisClient_GetEntryPointInfosForValue(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	client.client.Set("my-string", "my-value", time.Minute)

	// when
	infos, err := client.GetEntryPointInfos("my-string")

	// then
	assert.Nil(t, err)
	expected := datasource.EntryPointInfos{
		Type:       datasource.Value,
		Length:     uint64(len("my-value")),
		TimeToLive: time.Minute,
	}
	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Persist("my-string")

	// when
	infos, err = client.GetEntryPointInfos("my-string")

	// then
	assert.Nil(t, err)
	expected.TimeToLive = -1 * time.Second
	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Set("my-integer", 12345, -1)

	// when
	infos, err = client.GetEntryPointInfos("my-string")

	// then
	assert.Nil(t, err)
	expected.Length = 8
	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Set("my-integer", true, -1)

	// when
	infos, err = client.GetEntryPointInfos("my-string")

	// then
	assert.Nil(t, err)
	expected.Length = 8
	assertWithTimeToLive(t, expected, infos)
}

func assertWithTimeToLive(t *testing.T, expected datasource.EntryPointInfos, actual datasource.EntryPointInfos) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Length, actual.Length)
	assert.InDelta(t, expected.TimeToLive, actual.TimeToLive, float64(time.Second))
}

func TestRedisClient_GetEntryPointInfosForHash(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := map[string]interface{}{
		"my-field-1": "my-value",
		"my-field-2": 1234,
		"my-field-3": true,
	}
	client.client.HMSet("my-hash", values)
	client.client.Expire("my-hash", time.Minute)

	// when
	infos, err := client.GetEntryPointInfos("my-hash")

	// then
	assert.Nil(t, err)
	expected := datasource.EntryPointInfos{
		Type:       datasource.Hash,
		Length:     3,
		TimeToLive: time.Minute,
	}

	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Persist("my-hash")

	// when
	infos, err = client.GetEntryPointInfos("my-hash")

	// then
	assert.Nil(t, err)
	expected.TimeToLive = -1 * time.Second

	assertWithTimeToLive(t, expected, infos)
}

func TestRedisClient_GetEntryPointInfosForSet(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.SAdd("my-set", values...)
	client.client.Expire("my-set", time.Minute)

	// when
	infos, err := client.GetEntryPointInfos("my-set")

	// then
	assert.Nil(t, err)
	expected := datasource.EntryPointInfos{
		Type:       datasource.Set,
		Length:     3,
		TimeToLive: time.Minute,
	}

	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Persist("my-set")

	// when
	infos, err = client.GetEntryPointInfos("my-set")

	// then
	assert.Nil(t, err)
	expected.TimeToLive = -1 * time.Second

	assertWithTimeToLive(t, expected, infos)
}

func TestRedisClient_GetEntryPointInfosForList(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.LPush("my-list", values...)
	client.client.Expire("my-list", time.Minute)

	// when
	infos, err := client.GetEntryPointInfos("my-list")

	// then
	assert.Nil(t, err)
	expected := datasource.EntryPointInfos{
		Type:       datasource.List,
		Length:     3,
		TimeToLive: time.Minute,
	}

	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Persist("my-list")

	// when
	infos, err = client.GetEntryPointInfos("my-list")

	// then
	assert.Nil(t, err)
	expected.TimeToLive = -1 * time.Second

	assertWithTimeToLive(t, expected, infos)
}

func TestRedisClient_GetEntryPointInfosForOrderedSet(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := []redis.Z{
		redis.Z{
			Score:  0.5,
			Member: "my-first-value",
		},
		redis.Z{
			Score:  0.5,
			Member: "my-second-value",
		},
		redis.Z{
			Score:  1.5,
			Member: "my-third-value",
		},
		redis.Z{
			Score:  1.5,
			Member: 1234,
		},
		redis.Z{
			Score:  14.5,
			Member: 12654,
		},
		redis.Z{
			Score:  14.5,
			Member: 562763.76,
		},
	}
	err = client.client.ZAdd("my-zset", values...).Err()
	assert.Nil(t, err)
	client.client.Expire("my-zset", time.Minute)

	// when
	infos, err := client.GetEntryPointInfos("my-zset")

	// then
	assert.Nil(t, err)
	expected := datasource.EntryPointInfos{
		Type:       datasource.ScoredSet,
		Length:     6,
		TimeToLive: time.Minute,
	}

	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Persist("my-zset")

	// when
	infos, err = client.GetEntryPointInfos("my-zset")

	// then
	assert.Nil(t, err)
	expected.TimeToLive = -1 * time.Second

	assertWithTimeToLive(t, expected, infos)
}

func TestRedisClient_GetEntryPointInfosForMissingKey(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	// when
	_, err = client.GetEntryPointInfos("unknown-key")

	// then
	assert.NotNil(t, err, "An error is expected here, because the key does not exist")
}

func TestRedisClient_GetContentForValue(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	client.client.Set("my-string", "my-value", -1)
	data := make(chan datasource.DataBatch, 1)

	// when
	actionStatus, err := client.GetContent("my-string", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	expected := datasource.DataBatch{
		Size: 1,
		Data: []interface{}{"my-value"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected content is %+v, but actual is %+v", expected, result)
	}

	// given
	client.client.Set("my-integer", 12345, -1)
	data = make(chan datasource.DataBatch, 1)

	// when
	actionStatus, err = client.GetContent("my-integer", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result = <-data
	expected = datasource.DataBatch{
		Size: 1,
		Data: []interface{}{"12345"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected content is %+v, but actual is %+v", expected, result)
	}

	// givem
	client.client.Set("my-boolean", true, -1)
	data = make(chan datasource.DataBatch, 1)

	// when
	actionStatus, err = client.GetContent("my-boolean", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result = <-data
	expected = datasource.DataBatch{
		Size: 1,
		Data: []interface{}{"1"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Expected content is %+v, but actual is %+v", expected, result)
	}

	// given
	client.client.Set("my-float", 1234.567, -1)
	data = make(chan datasource.DataBatch, 1)

	// when
	actionStatus, err = client.GetContent("my-float", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result = <-data
	expected = datasource.DataBatch{
		Size: 1,
		Data: []interface{}{"1234.567"},
	}
	assert.Equal(t, expected, result)
}

func TestRedisClient_GetContentForHash(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := map[string]interface{}{
		"my-field-1": "my-value",
		"my-field-2": 1234,
		"my-field-3": true,
	}
	client.client.HMSet("my-hash", values)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-hash", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	expected := datasource.DataBatch{
		Size: 1,
		Data: []interface{}{
			map[string]string{
				"my-field-1": "my-value",
				"my-field-2": "1234",
				"my-field-3": "1",
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestRedisClient_GetContentForSet(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.SAdd("my-set", values...)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-set", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	expected := datasource.DataBatch{
		Size: 3,
		Data: []interface{}{
			"1",
			"1234",
			"my-value",
		},
	}
	assert.Equal(t, expected.Size, result.Size)
	assert.True(t, sliceUnorderedEqual(result.Data, expected.Data), "Expected content is %+v, but actual is %+v", expected.Data, result.Data)
}

func TestRedisClient_GetContentForList(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.LPush("my-list", values...)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-list", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	expected := datasource.DataBatch{
		Size: 3,
		Data: []interface{}{
			"my-value",
			"1",
			"1234",
		},
	}

	assert.Equal(t, expected.Size, result.Size)
	assert.True(t, sliceUnorderedEqual(result.Data, expected.Data), "Expected content is %+v, but actual is %+v", expected.Data, result.Data)
}

func TestRedisClient_GetContentForOrderedSet(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	values := []redis.Z{
		{
			Score:  0.5,
			Member: "my-first-value",
		},
		{
			Score:  1.5,
			Member: 1234,
		},
		{
			Score:  14.5,
			Member: 12654,
		},
		{
			Score:  2.5,
			Member: "my-third-value",
		},
		{
			Score:  0.5,
			Member: "my-second-value",
		},
		{
			Score:  231.5,
			Member: 562763.76,
		},
	}
	err = client.client.ZAdd("my-zset", values...).Err()
	assert.Nil(t, err)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-zset", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	expected := datasource.DataBatch{
		Size: 1,
		Data: []interface{}{
			map[float64][]string{
				0.5:   {"my-first-value", "my-second-value"},
				1.5:   {"1234"},
				2.5:   {"my-third-value"},
				14.5:  {"12654"},
				231.5: {"562763.76"},
			},
		},
	}
	assert.Equal(t, expected, result)
}

func TestRedisClient_GetContentForMissingKey(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	data := make(chan datasource.DataBatch, 100)

	// when
	_, err = client.GetContent("unknown-key", "", data)

	// then
	assert.NotNil(t, err, "An error is expected here, because the key does not exist")
}

func TestRedisClient_DeleteEntrypointChildren(t *testing.T) {
	// given
	testData := []struct {
		key   string
		count int
	}{
		{key: "group-atom", count: 200},
		{key: "group-atic", count: 200},
		{key: "group-artic", count: 200},
		{key: "group-bolton", count: 200},
	}

	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	for _, d := range testData {
		// In the case, also create the parent to check it is not deleted.
		client.client.Set(d.key, d.key, time.Minute)
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(s, s, time.Minute)
		}
	}
	// Check the number of keys before.
	keys, err := client.client.Keys("*").Result()
	assert.Nil(t, err)
	if len(keys) != 804 { // There are three root nodes plus all their leaves.
		t.Fatalf("Expected result size is %d, but actual is %d", 804, len(keys))
	}

	errorChannel := make(chan error, 10)

	// when
	actionStatus, err := client.DeleteEntrypointChildren("group-atic", errorChannel)

	// then
	assert.Nil(t, err)
	assert.Equal(t, datasource.Moved, actionStatus)

	// Wait for error channel to be closed.
	err = <-errorChannel
	assert.Nil(t, err)

	keys, err = client.client.Keys("*").Result()
	assert.Nil(t, err)
	if len(keys) != 604 { // There are still four root nodes plus all the leaves of only three of them.
		t.Fatalf("Expected result size is %d, but actual is %d", 604, len(keys))
	}
}

func TestRedisClient_DeleteEntrypointChildrenInReadOnlyMode(t *testing.T) {
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
			ReadOnly:  true,
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	errorChannel := make(chan error, 10)

	// when
	_, err = client.DeleteEntrypointChildren("any", errorChannel)

	// then
	assert.NotNil(t, err)
}

func TestRedisClient_DeleteEntrypoint(t *testing.T) {
	// given
	testData := []struct {
		key   string
		count int
	}{
		{key: "group-atom", count: 200},
		{key: "group-atic", count: 200},
		{key: "group-artic", count: 200},
		{key: "group-bolton", count: 200},
	}

	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	for _, d := range testData {
		// In the case, also create the parent to check that only this one is delete.
		client.client.Set(d.key, d.key, time.Minute)
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(s, s, time.Minute)
		}
	}
	// Check the number of keys before.
	keys, err := client.client.Keys("*").Result()
	assert.Nil(t, err)
	if len(keys) != 804 { // There are three root nodes plus all their leaves.
		t.Fatalf("Expected result size is %d, but actual is %d", 804, len(keys))
	}

	// when
	err = client.DeleteEntrypoint("group-artic")

	// then
	assert.Nil(t, err)

	keys, err = client.client.Keys("*").Result()
	assert.Nil(t, err)
	if len(keys) != 803 { // There are three root nodes plus all the leaves of all four nodes.
		t.Fatalf("Expected result size is %d, but actual is %d", 803, len(keys))
	}
}

func TestRedisClient_DeleteEntrypointInReadOnlyMode(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
			ReadOnly:  true,
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	// when
	err = client.DeleteEntrypoint("group-artic")

	// then
	assert.NotNil(t, err)
}

func TestRedisClient_CommandInReadOnlyMode(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Id:        "test",
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
			ReadOnly:  true,
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	// when + then
	_, err = client.ExecuteCommand([]interface{}{"KEYS", "*"}, "")
	assert.Nil(t, err)

	// when + then
	_, err = client.ExecuteCommand([]interface{}{"SET", "test", "1234"}, "")
	assert.Equal(t, err.Error(), "the data source test can only be read")

	// when + then
	_, err = client.ExecuteCommand([]interface{}{"CLUSTER", "NODES"}, "")
	assert.NotNil(t, err)
	assert.NotEqual(t, err.Error(), "the data source test can only be read")

	// when + then
	_, err = client.ExecuteCommand([]interface{}{"CLUSTER", "FORGET"}, "")
	assert.Equal(t, err.Error(), "the data source test can only be read")

	// when + then
	_, err = client.ExecuteCommand([]interface{}{"CLIENT", "KILL"}, "")
	assert.Equal(t, err.Error(), "the data source test can only be read")
}

func TestRedisClient_GetInfos(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	// when
	infos, err := client.GetInfos()

	// then
	assert.Nil(t, err)
	assert.Equal(t, datasource.Cluster{}, infos)
}

func TestRedisClient_GetStatus(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	// when
	status, err := client.GetStatus()

	// then
	assert.Nil(t, err)
	assert.NotNil(t, status.Timestamp)
	assert.Empty(t, status.StateSections)
	assert.NotEmpty(t, status.NodeStates)
}

func TestRedisClient_SetAndGetValueWithExecuteCommand(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Id:        "test",
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
			ReadOnly:  false,
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()
	_, _ = client.ExecuteCommand([]interface{}{"SET", "key", "value"}, "")

	// when
	rs, _ := client.ExecuteCommand([]interface{}{"GET", "key"}, "")

	// then
	assert.Equal(t, rs, "value")
}

func TestRedisClient_ExecuteUnknownCommand(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Id:        "test",
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
			ReadOnly:  false,
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll()
		client.Close()
	}()

	// when
	_, err = client.ExecuteCommand([]interface{}{"AN_UNKNOWN_COMMAND", "AN_UNKNOWN_ARG"}, "")

	// then
	assert.Contains(t, err.Error(), "ERR unknown command")
}

func sliceUnorderedEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v1 := range a {
		equal := false
		for _, v2 := range b {
			if v1 == v2 {
				equal = true
				break
			}
		}
		if !equal {
			return false
		}
	}
	return true
}

func TestRedisClient_GetContentForStream(t *testing.T) {
	// TODO
}

func TestRedisClient_Consume(t *testing.T) {
	// TODO
}
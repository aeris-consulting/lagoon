package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
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
	client.client.ConfigSet(context.Background(), "requirepass", "this-is-my-password")

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
		clientWithPassword.client.ConfigSet(context.Background(), "requirepass", "")
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for _, d := range testData {
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(context.Background(), s, s, time.Minute)
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

	// List all the entry points of level 0 only.
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

func TestRedisClient_ListAllEntryPointsWhenThereAreMoreThanChannelAndScanSize(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for i := 0; i < (int(scanSize) + 100); i++ {
		client.client.Set(context.Background(), strconv.Itoa(i), strconv.Itoa(i), time.Minute)
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
	assert.Equal(t, (int(scanSize) + 100), len(actualData))
}

func TestRedisClient_ListEntryPointsWithKeyHashTags(t *testing.T) {
	// given
	testData := []string{
		"{any:value}:split:here",
		"{{any:value}:here}:and:split-there",
		"just:{split:after}:there",
	}

	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for _, d := range testData {
		client.client.Set(context.Background(), d, d, time.Minute)
	}

	// List all the entry points in several parts.
	dataChannel := make(chan datasource.DataBatch, 10)

	// when
	actionStatus, err := client.ListEntryPoints("*", dataChannel, 0, 4)

	// then
	assert.Nil(t, err)
	assert.Equal(t, datasource.Moved, actionStatus)

	actualData := []interface{}{}
	for batch := range dataChannel {
		actualData = append(actualData, batch.Data...)
	}
	keys := []interface{}{}
	for _, entrypoint := range actualData {
		keys = append(keys, string(entrypoint.(*datasource.EntryPointNode).Path))
	}

	expectedResult := []interface{}{
		"just",
		"just:{split:after}",
		"just:{split:after}:there",
		"{any:value}",
		"{any:value}:split",
		"{any:value}:split:here",
		"{{any:value}:here}",
		"{{any:value}:here}:and",
		"{{any:value}:here}:and:split-there",
	}

	EqualUnorderedSlices(t, keys, expectedResult)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for _, d := range testData {
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(context.Background(), s, s, time.Minute)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for _, d := range testData {
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(context.Background(), s, s, time.Minute)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	client.client.Set(context.Background(), "my-string", "my-value", time.Minute)

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
	client.client.Persist(context.Background(), "my-string")

	// when
	infos, err = client.GetEntryPointInfos("my-string")

	// then
	assert.Nil(t, err)
	expected.TimeToLive = -1 * time.Second
	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Set(context.Background(), "my-integer", 12345, -1)

	// when
	infos, err = client.GetEntryPointInfos("my-string")

	// then
	assert.Nil(t, err)
	expected.Length = 8
	assertWithTimeToLive(t, expected, infos)

	// given
	client.client.Set(context.Background(), "my-integer", true, -1)

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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := map[string]interface{}{
		"my-field-3": true,
		"my-field-1": "my-value",
		"my-field-2": 1234,
	}
	client.client.HMSet(context.Background(), "my-hash", values)
	client.client.Expire(context.Background(), "my-hash", time.Minute)

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
	client.client.Persist(context.Background(), "my-hash")

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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.SAdd(context.Background(), "my-set", values...)
	client.client.Expire(context.Background(), "my-set", time.Minute)

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
	client.client.Persist(context.Background(), "my-set")

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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.LPush(context.Background(), "my-list", values...)
	client.client.Expire(context.Background(), "my-list", time.Minute)

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
	client.client.Persist(context.Background(), "my-list")

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
		client.client.FlushAll(context.Background())
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
	err = client.client.ZAdd(context.Background(), "my-zset", values...).Err()
	assert.Nil(t, err)
	client.client.Expire(context.Background(), "my-zset", time.Minute)

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
	client.client.Persist(context.Background(), "my-zset")

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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	client.client.Set(context.Background(), "my-string", "my-value", -1)
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
	client.client.Set(context.Background(), "my-integer", 12345, -1)
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
	client.client.Set(context.Background(), "my-boolean", true, -1)
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
	client.client.Set(context.Background(), "my-float", 1234.567, -1)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := map[string]interface{}{
		"my-field-3": true,
		"my-field-1": "my-value",
		"my-field-2": 1234,
	}
	client.client.HMSet(context.Background(), "my-hash", values)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-hash", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data

	// The keys should be ordered.
	assert.Equal(t, datasource.DataBatch{
		Size: 3, Data: []interface{}{
			HashValue{"my-field-1", "my-value"},
			HashValue{"my-field-2", "1234"},
			HashValue{"my-field-3", "1"},
		},
	}, result)
}

func TestRedisClient_GetContentForHashBiggerThanScanSize(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	var values = map[string]interface{}{}
	// Values are saved unsorted.
	for i := int(scanSize) + 100 - 1; i >= 0; i-- {
		values["my-field-"+fmt.Sprintf("%10d", i)] = "my-value-" + fmt.Sprintf("%10d", i)
	}
	client.client.HMSet(context.Background(), "my-hash", values)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-hash", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	var expectedValues []interface{}
	// Result is expected sorted.
	for i := 0; i < int(scanSize)+100; i++ {
		expectedValues = append(expectedValues, HashValue{"my-field-" + fmt.Sprintf("%10d", i), "my-value-" + fmt.Sprintf("%10d", i)})
	}
	assert.Equal(t, datasource.DataBatch{Size: uint64(scanSize + 100), Data: expectedValues}, result)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.SAdd(context.Background(), "my-set", values...)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-set", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data

	// The results are expected ordered.
	assert.Equal(t, datasource.DataBatch{Size: 3, Data: []interface{}{
		"1",
		"1234",
		"my-value",
	}}, result)
}

func TestRedisClient_GetContentForSetBiggerThanScanSize(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	var values []interface{}
	// Values are saved unordered.
	for i := int(scanSize) + 100 - 1; i >= 0; i-- {
		values = append(values, "my-value-"+fmt.Sprintf("%10d", i))
	}
	client.client.SAdd(context.Background(), "my-set", values...)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-set", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data

	var expected []interface{}
	// Values are expected ordered.
	for i := 0; i < int(scanSize)+100; i++ {
		expected = append(expected, "my-value-"+fmt.Sprintf("%10d", i))
	}
	assert.Equal(t, datasource.DataBatch{Size: uint64(scanSize + 100), Data: expected}, result)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := []interface{}{
		"my-value",
		1234,
		true,
	}
	client.client.RPush(context.Background(), "my-list", values...)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-list", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	assert.Equal(t, 3, len(result.Data))
	assert.Equal(t, uint64(3), result.Size)
	EqualUnorderedSlices(t, result.Data, []interface{}{
		"my-value",
		"1234",
		"1",
	})
}

func TestRedisClient_GetContentForListBiggerThanScanSize(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	var values []interface{}
	// Values are saved unsorted.
	for i := int(scanSize) + 100 - 1; i >= 0; i-- {
		values = append(values, "my-value-"+fmt.Sprintf("%10d", i))
	}
	client.client.RPush(context.Background(), "my-list", values...)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-list", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	// Values are expected in saved order.
	assert.Equal(t, datasource.DataBatch{uint64(scanSize + 100), values}, result)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	values := []redis.Z{
		{
			Score:  0.5,
			Member: "my-first-value",
		},
		{
			Score:  14.5,
			Member: 324,
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
	err = client.client.ZAdd(context.Background(), "my-zset", values...).Err()
	assert.Nil(t, err)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-zset", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	// Result is expected ordered, by score and in the values sets as well.
	assert.Equal(t, datasource.DataBatch{5, []interface{}{
		SortedSetValues{
			Score:  0.5,
			Values: []string{"my-first-value", "my-second-value"},
		},
		SortedSetValues{
			Score:  1.5,
			Values: []string{"1234"},
		},
		SortedSetValues{
			Score:  2.5,
			Values: []string{"my-third-value"},
		},
		SortedSetValues{
			Score:  14.5,
			Values: []string{"12654", "324"},
		},
		SortedSetValues{
			Score:  231.5,
			Values: []string{"562763.76"},
		},
	},
	}, result)
}

func TestRedisClient_GetContentForOrderedSetBiggerThanScanSize(t *testing.T) {
	// given
	client := RedisClient{
		datasource: &datasource.DataSourceDescriptor{
			Bootstrap: fmt.Sprintf("redis://%s:%d", redisIp, redisPort),
		},
	}
	err := client.Open()
	assert.Nil(t, err)
	defer func() {
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	var values = []redis.Z{}
	for i := int(scanSize+100) - 1; i >= 0; i-- {
		values = append(values, redis.Z{float64(i), "my-value-" + strconv.Itoa(i)})
	}
	err = client.client.ZAdd(context.Background(), "my-zset", values...).Err()
	assert.Nil(t, err)

	data := make(chan datasource.DataBatch, 100)

	// when
	actionStatus, err := client.GetContent("my-zset", "", data)

	// then
	assert.Equal(t, datasource.Completed, actionStatus)
	result := <-data
	var expectedValues []interface{}
	for i := 0; i < int(scanSize)+100; i++ {
		expectedValues = append(expectedValues, SortedSetValues{float64(i), []string{"my-value-" + strconv.Itoa(i)}})
	}
	assert.Equal(t, datasource.DataBatch{Size: uint64(scanSize + 100), Data: expectedValues}, result)
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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for _, d := range testData {
		// In the case, also create the parent to check it is not deleted.
		client.client.Set(context.Background(), d.key, d.key, time.Minute)
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(context.Background(), s, s, time.Minute)
		}
	}
	// Check the number of keys before.
	keys, err := client.client.Keys(context.Background(), "*").Result()
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

	keys, err = client.client.Keys(context.Background(), "*").Result()
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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	for _, d := range testData {
		// In the case, also create the parent to check that only this one is delete.
		client.client.Set(context.Background(), d.key, d.key, time.Minute)
		for i := 1; i <= d.count; i++ {
			s := fmt.Sprintf("%s:%d", d.key, i)
			client.client.Set(context.Background(), s, s, time.Minute)
		}
	}
	// Check the number of keys before.
	keys, err := client.client.Keys(context.Background(), "*").Result()
	assert.Nil(t, err)
	if len(keys) != 804 { // There are three root nodes plus all their leaves.
		t.Fatalf("Expected result size is %d, but actual is %d", 804, len(keys))
	}

	// when
	err = client.DeleteEntrypoint("group-artic")

	// then
	assert.Nil(t, err)

	keys, err = client.client.Keys(context.Background(), "*").Result()
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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()
	_, _ = client.ExecuteCommand([]interface{}{"SET", "key", "value"}, "")

	// when
	rs, _ := client.ExecuteCommand([]interface{}{"GET", "key"}, "")

	// then
	assert.Equal(t, "value", rs)
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
		client.client.FlushAll(context.Background())
		client.Close()
	}()

	// when
	_, err = client.ExecuteCommand([]interface{}{"AN_UNKNOWN_COMMAND", "AN_UNKNOWN_ARG"}, "")

	// then
	assert.Contains(t, err.Error(), "ERR unknown command")
}

func TestRedisClient_GetContentForStream(t *testing.T) {
	// TODO
}

func TestRedisClient_Consume(t *testing.T) {
	// TODO
}

func EqualUnorderedSlices(t *testing.T, actual, expected []interface{}) {
	if len(actual) != len(expected) {
		t.Error(fmt.Sprintf("Lengths are different: %d != %d", len(actual), len(expected)))
	}
	for _, v1 := range actual {
		equal := false
		for _, v2 := range expected {
			if reflect.DeepEqual(v1, v2) {
				equal = true
				break
			}
		}
		if !equal {
			t.Error(fmt.Sprintf("Expected: '%v' \n\t\tbut actual '%v'\n\t\tactual value '%v' was not found", expected, actual, v1))
		}
	}
}

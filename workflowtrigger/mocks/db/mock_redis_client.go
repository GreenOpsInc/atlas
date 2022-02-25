package db

import (
	"github.com/greenopsinc/util/auditlog"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/serializer"
	"github.com/greenopsinc/util/serializerutil"
	"github.com/greenopsinc/util/team"
)

type MockDbOperator interface {
	GetClient() db.DbClient
}

type MockRedisClientOperator struct {
}

func New() MockDbOperator {
	return &MockRedisClientOperator{}
}

func (r *MockRedisClientOperator) GetClient() db.DbClient {
	return &RedisClientImpl{}
}

type MockDbClient interface {
	Close()
	StoreValue(key string, schema interface{})
	FetchTeamSchema(key string) team.TeamSchema
}

var rdb map[string]string

func init() {
	rdb = make(map[string]string)
}

type RedisClientImpl struct {
}

func (r *RedisClientImpl) InsertValueInList(key string, schema interface{}) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) InsertValueInTransactionlessList(key string, schema interface{}) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) UpdateHeadInList(key string, schema interface{}) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) UpdateHeadInTransactionlessList(key string, schema interface{}) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchNotification(key string) clientrequest.Notification {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchPipelineInfoList(key string, increment int) []auditlog.PipelineInfo {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchLatestPipelineInfo(key string) auditlog.PipelineInfo {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchClusterSchema(key string) cluster.ClusterSchema {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchClusterSchemaTransactionless(key string) cluster.ClusterSchema {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchHeadInClientRequestList(key string) clientrequest.ClientRequestPacket {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchLogList(key string, increment int) []auditlog.Log {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchLatestLog(key string) auditlog.Log {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) FetchStringList(key string) []string {
	if rdb[key] == "" {
		return make([]string, 0)
	}
	return serializer.Deserialize(rdb[key], serializerutil.StringListType).([]string)
}

func (r *RedisClientImpl) DeleteByPrefix(prefix string) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClientImpl) Close() {

}

func (r *RedisClientImpl) StoreValue(key string, schema interface{}) {
	if schema == nil {
		delete(rdb, key)
	} else {
		rdb[key] = serializer.Serialize(schema)
	}
}

func (r *RedisClientImpl) FetchTeamSchema(key string) team.TeamSchema {
	if rdb[key] == "" {
		return team.TeamSchema{}
	}
	return serializer.Deserialize(rdb[key], serializerutil.TeamSchemaType).(team.TeamSchema)
}

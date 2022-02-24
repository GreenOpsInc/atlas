package db

import (
	"github.com/greenopsinc/util/team"
)

//type MockClientType func() DbClient
//
//type MockClient struct {
//	MockGetClient MockClientType
//}
//
//func (m *MockClient) GetClient() DbClient {
//	return m.MockGetClient()
//}

type MockDbOperator interface {
	GetClient() DbClient
}

type MockRedisClientOperator struct {
	MockGetClient func() DbClient
}

func (r *MockRedisClientOperator) GetClient() DbClient {
	return &RedisClientImpl{}
}

type DbClient interface {
	Close()
	StoreValue(key string, schema interface{})
	FetchTeamSchema(key string) team.TeamSchema
}

type RedisClientImpl struct {
	db map[string]interface{}
}

func (r *RedisClientImpl) Close() {

}

func (r *RedisClientImpl) StoreValue(key string, schema interface{}) {
	r.db[key] = schema
}

func (r *RedisClientImpl) FetchTeamSchema(key string) team.TeamSchema {
	if r.db[key] == nil {
		return team.TeamSchema{}
	}
	return r.db[key].(team.TeamSchema)
}

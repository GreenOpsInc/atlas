package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/greenopsinc/util/auditlog"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/serializer"
	"github.com/greenopsinc/util/serializerutil"
	"github.com/greenopsinc/util/team"
)

type ObjectType string

const (
	notification     ObjectType = "NOTIFICATION"
	teamSchema       ObjectType = "TEAM_SCHEMA"
	stringList       ObjectType = "STRING_LIST"
	logListObj       ObjectType = "LOG_LIST"
	singleLog        ObjectType = "SINGLE_LOG"
	pipelineInfoList ObjectType = "PIPELINE_INFO_LIST"
	pipelineInfo     ObjectType = "PIPELINE_INFO"
	clientRequest    ObjectType = "CLIENT_REQUEST"
	clusterSchema    ObjectType = "CLUSTER_SCHEMA"
	metadata         ObjectType = "METADATA"
)

type ListStoreOperation string

const (
	none   ListStoreOperation = "NONE"
	insert ListStoreOperation = "INSERT"
	update ListStoreOperation = "UPDATE"
)

type RedisCommand string

const (
	set     RedisCommand = "SET"
	get     RedisCommand = "GET"
	watch   RedisCommand = "WATCH"
	unwatch RedisCommand = "UNWATCH"
	exists  RedisCommand = "EXISTS"
	lrange  RedisCommand = "LRANGE"
	lindex  RedisCommand = "LINDEX"
	del     RedisCommand = "DEL"
	lpop    RedisCommand = "LPOP"
	rpop    RedisCommand = "RPOP"
	lpush   RedisCommand = "LPUSH"
	rpush   RedisCommand = "RPUSH"
	lset    RedisCommand = "LSET"
	multi   RedisCommand = "MULTI"
	exec    RedisCommand = "EXEC"
	keys    RedisCommand = "KEYS"
)

const (
	LogIncrement int = 15
)

type DbOperator interface {
	GetClient() DbClient
	Close()
}

type RedisClientOperator struct {
	pool *redis.Pool
}

type StepMetadata struct {
	ArgoRepoSchema *ArgoRepoSchema `json:"argoRepoSchema"`
}

type ArgoRepoSchema struct {
	RepoURL        string `json:"repoURL"`
	TargetRevision string `json:"targetRevision"`
	Path           string `json:"path"`
}

func NewArgoRepoSchema(repoURL, targetRevision, path string) *ArgoRepoSchema {
	if targetRevision == "" {
		targetRevision = "main"
	}
	if path == "" {
		path = "/"
	}
	return &ArgoRepoSchema{
		RepoURL:        repoURL,
		TargetRevision: targetRevision,
		Path:           path,
	}
}

func New(address string, password string) DbOperator {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address, redis.DialPassword(password))
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		MaxIdle:   3,
		MaxActive: 5,
		//Wait:        true,
		IdleTimeout: time.Duration(time.Minute),
	}
	return &RedisClientOperator{pool: pool}
}

func (r *RedisClientOperator) GetClient() DbClient {
	conn := r.pool.Get()
	_, err := conn.Do("PING")
	if err != nil {
		conn.Close()
		panic(fmt.Sprintf("Db client could not be fetched, please try again %s", err))
	}
	return &RedisClientImpl{client: conn}
}

func (r *RedisClientOperator) Close() {
	r.pool.Close()
}

type DbClient interface {
	Close()
	StoreValue(key string, schema interface{})
	InsertValueInList(key string, schema interface{})
	InsertValueInTransactionlessList(key string, schema interface{})
	UpdateHeadInList(key string, schema interface{})
	UpdateHeadInTransactionlessList(key string, schema interface{})
	FetchNotification(key string) clientrequest.Notification
	FetchPipelineInfoList(key string, increment int) []auditlog.PipelineInfo
	FetchLatestPipelineInfo(key string) auditlog.PipelineInfo
	FetchTeamSchema(key string) team.TeamSchema
	FetchClusterSchema(key string) cluster.ClusterSchema
	FetchClusterSchemaTransactionless(key string) cluster.ClusterSchema
	FetchHeadInClientRequestList(key string) clientrequest.ClientRequestPacket
	FetchLogList(key string, increment int) []auditlog.Log
	FetchLatestLog(key string) auditlog.Log
	FetchLatestDeploymentLog(key string) auditlog.Log
	FetchLatestRemediationLog(key string) auditlog.Log
	FetchMetadata(key string) *StepMetadata
	FetchTransactionless(key string, objectType ObjectType) interface{}
	FetchStringList(key string) []string
	DeleteByPrefix(prefix string)
}

type RedisClientImpl struct {
	client            redis.Conn
	currentWatchedKey string
}

func (r *RedisClientImpl) Close() {
	r.client.Close()
}

func (r *RedisClientImpl) StoreValue(key string, schema interface{}) {
	r.store(key, schema, none)
}

func (r *RedisClientImpl) InsertValueInList(key string, schema interface{}) {
	r.store(key, schema, insert)
}

func (r *RedisClientImpl) UpdateHeadInList(key string, schema interface{}) {
	r.store(key, schema, update)
}

func (r *RedisClientImpl) InsertValueInTransactionlessList(key string, object interface{}) {
	log.Printf("Storing schema for key without a transaction %s", key)
	serializedObject := serializer.Serialize(object)
	if object == nil {
		_ = redisWrapperFunc(r.client.Do(string(rpop), key))
	} else {
		_ = redisWrapperFunc(r.client.Do(string(rpush), key, serializedObject))
	}
}

func (r *RedisClientImpl) UpdateHeadInTransactionlessList(key string, object interface{}) {
	existsReply, _ := redis.Bool(r.client.Do(string(exists), key))
	if !existsReply {
		panic(errors.New("key does not exist in DB"))
	}
	log.Printf("Storing schema for key without a transaction %s", key)
	serializedObject := serializer.Serialize(object)
	if object == nil {
		_ = redisWrapperFunc(r.client.Do(string(lpop), key))
	} else {
		_ = redisWrapperFunc(r.client.Do(string(lpush), key, serializedObject))
	}
}

func (r *RedisClientImpl) store(key string, object interface{}, listStoreOperation ListStoreOperation) {
	log.Printf("Storing schema for key %s", key)
	serializedObject := serializer.Serialize(object)
	if key != r.currentWatchedKey {
		_ = redisWrapperFunc(r.client.Do(string(unwatch)))
		r.currentWatchedKey = ""
	}
	_ = redisWrapperFunc(r.client.Do(string(multi)))
	if listStoreOperation == none {
		if object == nil {
			_ = redisWrapperFunc(r.client.Do(string(del), key))
		} else {
			_ = redisWrapperFunc(r.client.Do(string(set), key, serializedObject))
		}
	} else if listStoreOperation == insert {
		if object == nil {
			_ = redisWrapperFunc(r.client.Do(string(lpop), key))
		} else {
			_ = redisWrapperFunc(r.client.Do(string(lpush), key, serializedObject))
		}
	} else if listStoreOperation == update {
		if object == nil {
			_ = redisWrapperFunc(r.client.Do(string(lpop), key))
		} else {
			_ = redisWrapperFunc(r.client.Do(string(lset), key, 0, serializedObject))
		}
	}
	var res interface{}
	res = redisWrapperFunc(r.client.Do(string(exec)))
	if res == nil {
		panic(errors.New("the transaction was interrupted"))
	}
}

func (r *RedisClientImpl) FetchNotification(key string) clientrequest.Notification {
	ret := r.fetch(key, notification, -1)
	if ret == nil {
		return clientrequest.Notification{}
	}
	return ret.(clientrequest.Notification)
}

func (r *RedisClientImpl) FetchPipelineInfoList(key string, increment int) []auditlog.PipelineInfo {
	infoList := r.fetch(key, pipelineInfoList, increment)
	if infoList == nil {
		return make([]auditlog.PipelineInfo, 0)
	}
	return infoList.([]auditlog.PipelineInfo)
}

func (r *RedisClientImpl) FetchLatestPipelineInfo(key string) auditlog.PipelineInfo {
	ret := r.fetch(key, pipelineInfo, -1)
	if ret == nil {
		return auditlog.PipelineInfo{}
	}
	return ret.(auditlog.PipelineInfo)
}

func (r *RedisClientImpl) FetchTeamSchema(key string) team.TeamSchema {
	ret := r.fetch(key, teamSchema, -1)
	if ret == nil {
		return team.TeamSchema{}
	}
	return ret.(team.TeamSchema)
}

func (r *RedisClientImpl) FetchClusterSchema(key string) cluster.ClusterSchema {
	ret := r.fetch(key, clusterSchema, -1)
	if ret == nil {
		return cluster.ClusterSchema{}
	}
	return ret.(cluster.ClusterSchema)
}

func (r *RedisClientImpl) FetchClusterSchemaTransactionless(key string) cluster.ClusterSchema {
	ret := r.FetchTransactionless(key, clusterSchema)
	if ret == nil {
		return cluster.ClusterSchema{}
	}
	return ret.(cluster.ClusterSchema)
}

func (r *RedisClientImpl) FetchStringList(key string) []string {
	ret := r.fetch(key, stringList, -1)
	if ret == nil {
		return make([]string, 0)
	}
	return ret.([]string)
}

func (r *RedisClientImpl) FetchLogList(key string, increment int) []auditlog.Log {
	logList := r.fetch(key, logListObj, increment)
	if logList == nil {
		return make([]auditlog.Log, 0)
	}
	return logList.([]auditlog.Log)
}

func (r *RedisClientImpl) FetchLatestLog(key string) auditlog.Log {
	fetchedLog := r.fetch(key, singleLog, -1)
	if fetchedLog == nil {
		return nil
	}
	return fetchedLog.(auditlog.Log)
}

func (r *RedisClientImpl) FetchLatestDeploymentLog(key string) auditlog.Log {
	var latestLog auditlog.Log
	latestLog = r.FetchLatestLog(key)
	if latestLog == nil {
		return nil
	}
	switch latestLog.(type) {
	case *auditlog.DeploymentLog:
		return latestLog
	default:
		idx := 0
		logIncrementVal := 0
		var logList []auditlog.Log
		logList = r.FetchLogList(key, 0)
		for idx < len(logList) {
			switch logList[idx].(type) {
			case *auditlog.DeploymentLog:
				return logList[idx]
			default:
				idx++
				if idx == len(logList) {
					logIncrementVal++
					logList = r.FetchLogList(key, logIncrementVal)
					idx = 0
				}
			}
		}
		return nil
	}
}

func (r *RedisClientImpl) FetchLatestRemediationLog(key string) auditlog.Log {
	var latestLog auditlog.Log
	latestLog = r.FetchLatestLog(key)
	if latestLog == nil {
		return nil
	}
	switch latestLog.(type) {
	case *auditlog.RemediationLog:
		return latestLog
	default:
		idx := 0
		logIncrementVal := 0
		var logList []auditlog.Log
		logList = r.FetchLogList(key, 0)
		for idx < len(logList) {
			switch logList[idx].(type) {
			case *auditlog.RemediationLog:
				return logList[idx]
			default:
				idx++
				if idx == len(logList) {
					logIncrementVal++
					logList = r.FetchLogList(key, logIncrementVal)
					idx = 0
				}
			}
		}
		return nil
	}
}

func (r *RedisClientImpl) FetchHeadInClientRequestList(key string) clientrequest.ClientRequestPacket {
	existsReply, _ := redis.Bool(r.client.Do(string(exists), key))
	if !existsReply {
		return clientrequest.ClientRequestPacket{}
	}
	ret := r.FetchTransactionless(key, clientRequest)
	if ret == nil {
		return clientrequest.ClientRequestPacket{}
	}
	return ret.(clientrequest.ClientRequestPacket)
}

func (r *RedisClientImpl) FetchTransactionless(key string, objectType ObjectType) interface{} {
	log.Printf("Fetching schema for key without a transaction %s", key)
	existsReply, _ := redis.Bool(r.client.Do(string(exists), key))
	if !existsReply {
		log.Printf("Key doesn't exist")
		return nil
	}
	var reply interface{}
	if objectType == clusterSchema {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(get), key)))
		return serializer.Deserialize(reply.(string), serializerutil.ClusterSchemaType)
	} else if objectType == clientRequest {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(lindex), key, 0)))
		return serializer.Deserialize(reply.(string), serializerutil.ClientPacketType)
	}
	panic(errors.New("objectType did not match type"))
}

func (r *RedisClientImpl) FetchMetadata(key string) *StepMetadata {
	return r.fetch(key, metadata, -1).(*StepMetadata)
}

func (r *RedisClientImpl) DeleteByPrefix(prefix string) {
	keys := redisWrapperFunc(r.client.Do(string(keys), fmt.Sprintf("%s*", prefix)))
	for _, k := range keys.([]interface{}) {
		key := string(k.([]byte))
		redisWrapperFunc(r.client.Do(string(del), key))
	}
}

func (r *RedisClientImpl) fetch(key string, objectType ObjectType, increment int) interface{} {
	log.Printf("Fetching schema for key %s", key)
	_ = redisWrapperFunc(r.client.Do(string(unwatch)))
	_ = redisWrapperFunc(r.client.Do(string(watch), key))
	r.currentWatchedKey = key
	//If the key doesn't exist, return null (1 is exists, 0 is does not exist)
	existsReply := redisWrapperFunc(r.client.Do(string(exists), key)).(int64)
	var reply interface{}
	if existsReply == 0 {
		return nil
	} else if objectType == teamSchema {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(get), key)))
		return serializer.Deserialize(reply.(string), serializerutil.TeamSchemaType)
	} else if objectType == notification {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(get), key)))
		return serializer.Deserialize(reply.(string), serializerutil.NotificationType)
	} else if objectType == stringList {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(get), key)))
		return serializer.Deserialize(reply.(string), serializerutil.StringListType)
	} else if objectType == logListObj {
		startIdx := increment * LogIncrement
		reply = redisWrapperFunc(redis.Strings(r.client.Do(string(lrange), key, startIdx, startIdx+LogIncrement-1)))
		logArray := make([]auditlog.Log, 0)
		for _, val := range reply.([]string) {
			logArray = append(logArray, serializer.Deserialize(val, serializerutil.LogType).(auditlog.Log))
		}
		return logArray
	} else if objectType == pipelineInfoList {
		startIdx := increment * LogIncrement
		reply = redisWrapperFunc(redis.Strings(r.client.Do(string(lrange), key, startIdx, startIdx+LogIncrement-1)))
		pipelineInfoArray := make([]auditlog.PipelineInfo, 0)
		for _, val := range reply.([]string) {
			pipelineInfoArray = append(pipelineInfoArray, serializer.Deserialize(val, serializerutil.PipelineInfoType).(auditlog.PipelineInfo))
		}
		return pipelineInfoArray
	} else if objectType == singleLog {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(lindex), key, 0)))
		return serializer.Deserialize(reply.(string), serializerutil.LogType)
	} else if objectType == pipelineInfo {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(lindex), key, 0)))
		return serializer.Deserialize(reply.(string), serializerutil.PipelineInfoType)
	} else if objectType == clusterSchema {
		reply = redisWrapperFunc(redis.String(r.client.Do(string(get), key)))
		return serializer.Deserialize(reply.(string), serializerutil.ClusterSchemaType)
	}
	panic(errors.New("could not find the correct match for fetching item from redis"))
}

func redisWrapperFunc(reply interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return reply
}

//WARNING: This method assumes that the payload is already a string
func getInterfaceArray(stringPayload interface{}) []interface{} {
	var interfaceArray []interface{}
	err := json.Unmarshal([]byte(stringPayload.(string)), &interfaceArray)
	if err != nil {
		panic(err)
	}
	return interfaceArray
}

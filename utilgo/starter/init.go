package starter

import "os"

const (
	//EnvVar Names
	dbAddress                     string = "ATLAS_DB_ADDRESS"
	dbPassword                    string = "ATLAS_DB_PASSWORD"
	kafkaAddress                  string = "KAFKA_BOOTSTRAP_SERVERS"
	repoServerAddress             string = "REPO_SERVER_ENDPOINT"
	commandDelegatorServerAddress string = "COMMAND_DELEGATOR_SERVER_ENDPOINT"
	noAuth                        string = "NO_AUTH"

	//Default Names
	dbDefaultAddress                     string = "localhost:6379"
	dbDefaultPassword                    string = ""
	kafkaDefaultAddress                  string = "localhost:29092"
	repoServerDefaultAddress             string = "http://localhost:8081"
	commandDelegatorServerDefaultAddress string = "http://localhost:8080"
	noAuthDefaultValue                   string = "false"
)

func GetDbClientConfig() (string, string) {
	address := dbDefaultAddress
	password := dbDefaultPassword
	if val := os.Getenv(dbAddress); val != "" {
		address = val
	}
	if val := os.Getenv(dbPassword); val != "" {
		password = val
	}
	return address, password
}

func GetKafkaClientConfig() string {
	address := kafkaDefaultAddress
	if val := os.Getenv(kafkaAddress); val != "" {
		address = val
	}
	return address
}

func GetRepoServerClientConfig() string {
	address := repoServerDefaultAddress
	if val := os.Getenv(repoServerAddress); val != "" {
		address = val
	}
	return address
}

func GetCommandDelegatorServerClientConfig() string {
	address := commandDelegatorServerDefaultAddress
	if val := os.Getenv(commandDelegatorServerAddress); val != "" {
		address = val
	}
	return address
}

func GetNoAuthClientConfig() bool {
	if val := os.Getenv(noAuth); val == "true" {
		return true
	}
	return false

}

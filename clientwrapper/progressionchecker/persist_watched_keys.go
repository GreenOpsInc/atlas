package progressionchecker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"greenops.io/client/k8sdriver"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
)

const (
	clientWrapperPersistedConfigMapName      string = "client-wrapper-configmap"
	clientWrapperPersistedConfigMapNamespace string = "atlas"
)

func PersistWatchedKeys(kubernetesClient k8sdriver.KubernetesClientGetRestricted, watchedKeys map[string]datamodel.WatchKey) error {
	modifiedWatchedKeys := make(map[string][]byte)
	for key := range watchedKeys {
		if !watchedKeys[key].GeneratedCompletionEvent {
			buf := new(bytes.Buffer)
			enc := gob.NewEncoder(buf)
			if err := enc.Encode(watchedKeys[key]); err != nil {
				fmt.Println(err)
			}
			modifiedWatchedKeys[key] = buf.Bytes()
		}
	}
	err := kubernetesClient.StoreConfigMap(clientWrapperPersistedConfigMapName, clientWrapperPersistedConfigMapNamespace, modifiedWatchedKeys)
	if err != nil {
		return err
	}
	return nil
}

func RetrieveWatchedKeys(kubernetesClient k8sdriver.KubernetesClientGetRestricted) map[string]datamodel.WatchKey {
	persistedKeys, err := kubernetesClient.GetConfigMap(clientWrapperPersistedConfigMapName, clientWrapperPersistedConfigMapNamespace)
	if err != nil {
		log.Fatalf(err.Error())
	}
	watchedKeys := make(map[string]datamodel.WatchKey)
	for key := range persistedKeys {
		buf := bytes.NewBuffer(persistedKeys[key])
		dec := gob.NewDecoder(buf)
		var watchKey *datamodel.WatchKey
		if err := dec.Decode(&watchKey); err != nil {
			log.Println(err)
		}
		watchedKeys[key] = *watchKey
	}
	return watchedKeys
}

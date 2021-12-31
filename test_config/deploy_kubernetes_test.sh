minikubestatus=$(minikube status | grep host)
if [[ $minikubestatus != 'host: Running' ]]
then
  minikube start
fi
docker compose down
kubectl delete -f atlasinfrastructure.yaml
localhostip="$(minikube ssh 'grep host.minikube.internal /etc/hosts | cut -f1')"
localhostip=$(echo "$localhostip" | tr -d '\r')
perl -p -e "s/localhost/$localhostip/g" docker-compose.yml | docker compose -f - up -d zookeeper kafka
cd ../workflowtrigger
env GOOS=linux go build -v .
docker build -f test/Dockerfile -t atlasworkflowtrigger .
minikube image load atlasworkflowtrigger
cd ../commanddelegatorgo
env GOOS=linux go build -v ./command_delegator_api.go
docker build -f test/Dockerfile -t atlascommanddelegator .
minikube image load atlascommanddelegator
cd ../clientwrapper
env GOOS=linux go build -v ./atlasoperator/atlas_operator.go
docker build -f test/Dockerfile -t atlasclientwrapper .
minikube image load atlasclientwrapper
eval $(minikube -p minikube docker-env)
cd ../PipelineRepoServer/
./gradlew jibDockerBuild --image=atlasreposerver
cd ../WorkflowOrchestrator/
./gradlew jibDockerBuild --image=atlasworkfloworchestrator
cd ../test_config
perl -p -e "s/LOCALHOSTDYNAMICADDRESS/$localhostip/g" atlasinfrastructure.yaml | kubectl apply -f -

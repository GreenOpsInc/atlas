kubectl delete -f atlasinfrastructure_gke.yaml
kubectl delete -f kafkaservice.yaml
kubectl delete -f zookeeperinfrastructure.yaml
docker system prune
docker rmi $(docker image ls -a -q) -f
cd ../../../
./gradlew jibDockerBuild --image=atlasworkflowtrigger
docker tag atlasworkflowtrigger gcr.io/greenops-dev/atlasworkflowtrigger
docker push gcr.io/greenops-dev/atlasworkflowtrigger
cd ../PipelineRepoServer/
./gradlew jibDockerBuild --image=atlasreposerver
docker tag atlasreposerver gcr.io/greenops-dev/atlasreposerver
docker push gcr.io/greenops-dev/atlasreposerver
cd ../WorkflowOrchestrator/
./gradlew jibDockerBuild --image=atlasworkfloworchestrator
docker tag atlasworkfloworchestrator gcr.io/greenops-dev/atlasworkfloworchestrator
docker push gcr.io/greenops-dev/atlasworkfloworchestrator
cd ../ClientWrapper
#go build -i -v ./atlasoperator/atlas_operator.go
docker build . -t atlasclientwrapper
docker tag atlasclientwrapper gcr.io/greenops-dev/atlasclientwrapper
docker push gcr.io/greenops-dev/atlasclientwrapper
cd ../WorkflowTrigger/src/test/test_config

kubectl apply -f zookeeperinfrastructure.yaml
kubectl apply -f kafkaservice.yaml

localhostip=""

while [[ $localhostip == "" ]]
do
localhostip=$(kubectl describe svc kafka-service | echo $(grep 'LoadBalancer Ingress') | awk -v N=3 '{print $N}')
done

echo $localhostip
perl -p -e "s/dynamickafkaaddress/$localhostip/g" atlasinfrastructure_gke.yaml | kubectl apply -f -


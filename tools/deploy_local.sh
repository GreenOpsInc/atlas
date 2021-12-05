set -e

ip=$(k3s kubectl get nodes -o yaml | yq '.items[0].metadata.annotations."k3s.io/internal-ip"' -r)
root=$(git rev-parse --show-toplevel)

echo "deploying atlas on cluster at $ip"

echo "removing any old altas deployment..."
k3s kubectl delete -f $root/tools/atlas_infrastructure.yaml

echo "starting zookeeper & kafka"
CLUSTER_IP="$ip" docker compose -f $root/tools/docker-compose.yml up -d zookeeper kafka

cd $root/commanddelegator/     && ./gradlew jibDockerBuild --image=atlascommanddelegator
cd $root/pipelinereposerver/   && ./gradlew jibDockerBuild --image=atlasreposerver
cd $root/workfloworchestrator/ && ./gradlew jibDockerBuild --image=atlasworkfloworchestrator
cd $root/workfloworchestrator/ && ./gradlew jibDockerBuild --image=atlasworkflowtrigger

sed "s/LOCALHOSTDYNAMICADDRESS/$ip/g" $root/tools/atlas_infrastructure.yaml | k3s kubectl apply -f -

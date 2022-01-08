minikubestatus=$(minikube status | grep host)
if [ "$minikubestatus" != 'host: Running' ]
then
  minikube start
fi
kubectl delete -f deployment.yaml
localhostip="$(minikube ssh 'grep host.minikube.internal /etc/hosts | cut -f1')"
localhostip=$(echo "$localhostip" | tr -d '\r')
eval $(minikube -p minikube docker-env)
./gradlew jibDockerBuild --image=atlasverificationtool
perl -p -e "s/LOCALHOSTDYNAMICADDRESS/$localhostip/g" deployment.yaml | kubectl apply -f -

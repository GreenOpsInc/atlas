docker compose down
cd ../
./gradlew jibDockerBuild --image=atlasworkflowtrigger
cd ../PipelineRepoServer/
./gradlew jibDockerBuild --image=atlasreposerver
cd ../WorkflowTrigger/test_config/
docker compose up #-d
docker compose down
cd ../../../
./gradlew jibDockerBuild --image=atlasworkflowtrigger
cd ../PipelineRepoServer/
./gradlew jibDockerBuild --image=atlasreposerver
cd ../WorkflowOrchestrator/
./gradlew jibDockerBuild --image=atlasworkfloworchestrator
cd ../WorkflowTrigger/src/test/test_config
docker compose up -d
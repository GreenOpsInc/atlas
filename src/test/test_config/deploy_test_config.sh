docker compose down
cd ../../../
./gradlew jibDockerBuild --image=atlasworkflowtrigger
cd ../PipelineRepoServer/
./gradlew jibDockerBuild --image=atlasreposerver
cd ../WorkflowTrigger/src/test/test_config
docker compose up -d
docker-compose down
cd ../../../
call ./gradlew jibDockerBuild --image=atlasworkflowtrigger
cd ../PipelineRepoServer/
call ./gradlew jibDockerBuild --image=atlasreposerver
cd ../WorkflowTrigger/src/test/test_config
docker-compose up -d
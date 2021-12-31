package serializer

import (
	"github.com/greenopsinc/util/serializer"
	"greenops.io/workflowtrigger/pipelinestatus"
)

func Serialize(object interface{}) string {
	var err error
	var bytes []byte
	switch object.(type) {
	case pipelinestatus.PipelineStatus:
		bytes, err = pipelinestatus.MarshallPipelineStatus(object.(pipelinestatus.PipelineStatus)).MarshalJSON()
	default:
		return serializer.Serialize(object)
	}
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

package auditlog

type PipelineInfo struct {
	PipelineUvn string   `json:"pipelineUniqueVersionNumber"`
	Errors      []string `json:"errors"`
	StepList    []string `json:"stepList"`
}

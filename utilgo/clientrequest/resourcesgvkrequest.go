package clientrequest

type ResourcesGVKRequest struct {
	ResourceGVKList []*ResourceGVK `json:"resourceGvkList"`
}

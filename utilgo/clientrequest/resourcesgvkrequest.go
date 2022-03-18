package clientrequest

type ResourcesGVKRequest struct {
	ResourceGVKList []*ResourceGvk `json:"resourceGvkList"`
}

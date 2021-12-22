package clientrequest

type Notification struct {
	Successful bool   `json:"Successful"`
	Body       string `json:"Body"`
}

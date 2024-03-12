package contract

type InstanceName struct {
	Name string `validate:"required,min=5,max=100" json:"name"`
}

type Instance struct {
	Name          string  `validate:"required,min=5,max=100" json:"name"`
	State         string  `validate:"oneof=active inactive" json:"state"`
	ApiKey        *string `json:"apikey,omitempty"`
	ExternalId    string  `json:"externalId"`
}

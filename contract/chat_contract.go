package contract

type IsOnWhatsAppResponse struct {
	Query string `json:"query,omitempty"` // The query string used
	JID   string `json:"jid,omitempty"`   // The canonical user ID
	IsIn  bool   `json:"isIn,omitempty"`  // Whether the phone is registered or not.
}

type ChatNumbers struct {
	Numbers []string `validate:"required,gt=0,dive" json:"numbers"`
}

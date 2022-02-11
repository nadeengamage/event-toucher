package util

type PendingList struct {
	Clientele Clientele `json:"clientele"`
	Status    string    `json:"status"`
}

type Clientele struct {
	IDX        string `json:"idx"`
	Identifier string `json:"identificationNumber"`
	CreatedBy  string `json:"createdBy"`
}

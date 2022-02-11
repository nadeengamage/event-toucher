package util

type UploadRequest struct {
	IDX            string `json:"idx"`
	MasterCategory string `json:"imgMasterCategory"`
	SubCategory    string `json:"imgSubCategory"`
	FileName       string `json:"imgOriginalName"`
	ContentType    string `json:"imgContentType"`
	Longitude      string `json:"longitude"`
	Latitude       string `json:"latitude"`
	File           string `json:"image"`
}

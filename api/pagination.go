package api

type Page struct {
	Number        int32 `json:"number"`
	Size          int32 `json:"size"`
	TotalPages    int32 `json:"totalPages"`
	TotalElements int32 `json:"totalElements"`
}

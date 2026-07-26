package entity

type BulkRequest struct {
	IDs []int64 `json:"ids"`
}

type BulkItemResult struct {
	ID      int64  `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type BulkResult struct {
	Results []BulkItemResult `json:"results"`
}

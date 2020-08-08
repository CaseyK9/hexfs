package main

type UploadResponseSuccess struct {
	Status int `json:"status"`
	FileId string `json:"file_id"`
	Size int64 `json:"size"`
}

type ResponseError struct {
	Status int `json:"status"`
	Message string `json:"message"`
}

type StatsResponseSuccess struct {
	Status int `json:"status"`
	WebhookEnabled bool `json:"webhook_enabled"`
	MemAllocated string `json:"mem_allocated"`
	SpaceMax int64 `json:"space_max"`
	SpaceUsed int64 `json:"space_used"`
	MaxFileSize string `json:"max_file_size"`
	MinFileSize string `json:"min_file_size"`
}

type Ping struct {
	Status int `json:"status"`
}
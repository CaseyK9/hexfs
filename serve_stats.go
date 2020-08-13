package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

func ServeStats(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	n, _ := strconv.ParseInt(os.Getenv(UploadDirMaxSize), 0, 64)
	max, _ := strconv.ParseInt(os.Getenv(MaxSizeBytes), 0, 64)
	min, _ := strconv.ParseInt(os.Getenv(MinSizeBytes), 0, 64)

	SendJSONResponse(&w, StatsResponseSuccess{
		Status:         0,
		WebhookEnabled: os.Getenv(DiscordWebhookURL) != "",
		MemAllocated:   ByteCountSI(mem.Alloc),
		MaxFileSize:    ByteCountSI(uint64(max)),
		MinFileSize:    ByteCountSI(uint64(min)),
		SpaceMax:       n,
		SpaceUsed:      SizeOfUploadDir,
		Version:        VERSION,
	})
	return
}

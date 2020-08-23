[![hex-Artboard-1.png](https://i.postimg.cc/hG3JbB0b/hex-Artboard-1.png)](https://postimg.cc/4KVNsjqm)

## hexFS

Private file host.

Example: https://hexfs.vysion.cc

Set `HFS_ENDPOINT` to your API endpoint and `HFS_UPLOAD_KEY` to the key to use when uploading. these ones are required

Full list of variables to use

	Port = "HFS_PORT"
	UploadKey = "HFS_UPLOAD_KEY"
	MinSizeBytes = "HFS_MIN_SIZE_BYTES"
	MaxSizeBytes = "HFS_MAX_SIZE_BYTES"
	DiscordWebhookURL = "HFS_DISCORD_WEBHOOK"
	UploadDirMaxSize = "HFS_UPLOAD_DIR_MAX_SIZE"
	UploadDirPath = "HFS_UPLOAD_DIR_PATH" (defaults to working directory/uploads)
	Endpoint = "HFS_ENDPOINT"
	Frontend = "HFS_FRONTEND"
	
## how to upload something

POST / with your key in the `Authorization` header and the file in the `file` field on a `multipart/form-data` encoded request. you'll get the response as plaintext along with 200 OK
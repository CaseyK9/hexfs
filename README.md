## PixelsFS Storage Engine

yo mama jokes belong in 2016 

[demo site with optional front-end](https://pixels.moe)

### Features

- Private uploading (only you know the key for uploading files; viewing them is public)
- No dependencies on databases or external cloud storage services, meaning you have full control over your files
- Basic protection against DoS attacks & XSS scripts (use with cloudflare for optimal results)
- File deletion support
- Update posting to Discord webhook URL (monitor file additions/deletions)
- Custom 404 images
- Upload directory size limit
- Made in Go 

### How to run

- Clone
- Build
- Make a service (using systemd or smth idfk)
- Start service
- now get money and bitches

concerning docker, idk maybe later

### Notes

- all environment variables will be prefixed with `PIXELSFS` so they don't collide with other variables
- environment variable names:
    - `PIXELSFS_PORT` (optional, default: 3030)
    - `PIXELSFS_UPLOAD_KEY` (required)
    - `PIXELSFS_MIN_SIZE_BYTES` (optional, default: 512B)
    - `PIXELSFS_MAX_SIZE_BYTES` (optional, default: 50MB)
    - `PIXELSFS_DISCORD_WEBHOOK` (optional, should be discord webhook url)
    - `PIXELSFS_UPLOAD_DIR_MAX_SIZE` (optional, default: 10GB)
    - `PIXELSFS_UPLOAD_DIR_PATH` (optional, default: working directory + /uploads)
    
### Endpoints

- POST / 
    - with Authorization header containing value you set for `PIXELSFS_UPLOAD_KEY`
    - and `multipart/form-data` containing field `file`
    
Returns
```
type UploadResponseSuccess struct {
	Status int `json:"status"`
	FileId string `json:"file_id"`
	Size int64 `json:"size"`
}
```
    
- GET /stats

Returns
```
type StatsResponseSuccess struct {
	Status int `json:"status"`
	WebhookEnabled bool `json:"webhook_enabled"`
	MemAllocated string `json:"mem_allocated"`
	SpaceMax int64 `json:"space_max"`
	SpaceUsed int64 `json:"space_used"`
	MaxFileSize string `json:"max_file_size"`
	MinFileSize string `json:"min_file_size"`
}
```

- GET /:name
    - returns file if it exists
    
- GET /ping
    - returns `{"status":0}`
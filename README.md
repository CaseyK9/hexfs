## Pixels Storage Engine

### Features

- Private uploading (only you know the key for uploading files; viewing them is public)
- No dependencies on databases or external cloud storage services, meaning you have full control over your files
- Update posting to Discord webhook URL (monitor file additions/deletions)
- Custom 404 images
- Upload directory size limit
- Extremely low memory usage

### How to run

- Clone
- Build
- Create .env file in project root and populate with variables
- Make a service (using systemd or smth idfk)
- Start service

### Notes

- all environment variables will be prefixed with `PSE` so they don't collide with other variables
- environment variable names:
    - `PSE_PORT` (optional, default: 7250)
    - `PSE_UPLOAD_KEY` (required)
    - `PSE_MIN_SIZE_BYTES` (optional, default: 512B)
    - `PSE_MAX_SIZE_BYTES` (optional, default: 50MB)
    - `PSE_DISCORD_WEBHOOK` (optional, should be discord webhook url)
    - `PSE_UPLOAD_DIR_MAX_SIZE` (optional, default: 10GB)
    - `PSE_UPLOAD_DIR_PATH` (optional, default: working directory + /uploads)
    - `PSE_ENDPOINT` (optional, set this to append URL before file name in webhook.)
    
### Endpoints

- POST / 
    - with Authorization header containing value you set for `PSE_UPLOAD_KEY`
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
    Version string `json:"version"`
}
```

- GET /:name
    - returns file if it exists
    
- GET /ping
    - returns `{"status":0}`
    
### ShareX

Replace every placeholder (marked with >> CONTENT <<) with your own domain.

```json
{
  "Version": "13.0.1",
  "Name": "Pixels Storage Engine ShareX",
  "DestinationType": "ImageUploader, TextUploader, FileUploader",
  "RequestMethod": "POST",
  "RequestURL": ">> YOUR ENDPOINT HERE, NO TRAILING SLASH <<",
  "Headers": {
    "Authorization": ">> YOUR KEY HERE <<"
  },
  "Body": "MultipartFormData",
  "FileFormName": "file",
  "URL": ">> YOUR ENDPOINT HERE, NO TRAILING SLASH <</$json:name$"
}
```

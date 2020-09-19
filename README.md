[![hex-Artboard-1.png](https://i.postimg.cc/hG3JbB0b/hex-Artboard-1.png)](https://postimg.cc/4KVNsjqm)

## hexFS

Private file host.

### setup notes

.env template

```
# What port to listen on (default: 3000)
HFS_PORT=

# Upload key to use. Required
HFS_UPLOAD_KEY=

# The max size, in bytes, any file can be. (default: 50 mib)
HFS_MAX_SIZE_BYTES=

# The default URL to use in the file response. Required (e.g. https://test.com/fileID.ext) 
HFS_ENDPOINT=

# The URL to redirect to if the file/page is not found. If not specified it will 404.
HFS_FRONTEND=

# Google Cloud Storage bucket name. Required
GCS_BUCKET_NAME=

# Base64 encoded AES256 32-bit key. Required
GCS_SECRET_KEY=

# The JSON format of your service account key file location. Required
GOOGLE_APPLICATION_CREDENTIALS=
```

## how to upload something

POST / with your key in the `Authorization` header and the file in the `file` field on a `multipart/form-data` encoded request. you'll get the response as plaintext along with 200 OK

You can also include an additional field, `proxy`, and specify a custom domain to use in place of HFS_ENDPOINT.
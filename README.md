[![hex-Artboard-1.png](https://i.postimg.cc/hG3JbB0b/hex-Artboard-1.png)](https://postimg.cc/4KVNsjqm)

## hexFS

Private file host.

### setup notes

.env template

```
# What port to listen on (default: 3000)
HFS_PORT=

# Upload key to use. Required. 
# You can share this with friends you trust, if you want to.
HFS_UPLOAD_KEY=

# Deletion key to use. Required.
# KEEP THIS TO YOURSELF. Anyone with the key you set can delete ANY file!
HFS_DELETION_KEY=

# If set to 1, ANY user can upload to this service without ANY authentication.
# Deletion still requires the correct deletion key.
# It is your responsibility to make sure you monitor what gets uploaded via journalctl or other methods.
# If you feel that your server is a target of spam or abuse you should set this to 0 or not set it at all.
# YOU HAVE BEEN WARNED!
HFS_PUBLIC_MODE=0

# If set to 1, this will disable the builtin file extension filter (which rejects any file with an extension inside the list).
# Opens your server to malicious uploads from any user if HFS_PUBLIC_MODE is 1.
HFS_DISABLE_FILE_BLACKLIST=0

# The max size, in bytes, any file can be. (default: 50 mib)
HFS_MAX_SIZE_BYTES=

# The default URL to use in the file response. Required (e.g. https://test.com/fileID.ext, you'd specify "https://test.com") 
HFS_ENDPOINT=

# The URL to redirect to if the file/page is not found. If not specified it will 404.
HFS_FRONTEND=

# Google Cloud Storage bucket name. Required
GCS_BUCKET_NAME=

# Base64 encoded AES256 32-bit key. Required
# You'll use this to encrypt files using your own key instead of google holding on to them.
# More information: https://cloud.google.com/storage/docs/encryption/customer-supplied-keys
GCS_SECRET_KEY=

# The JSON format of your service account key file location. Required
GOOGLE_APPLICATION_CREDENTIALS=
```

## how to upload something

POST / with your key in the `Authorization` header and the file in the `file` field on a `multipart/form-data` encoded request. you'll get the response as plaintext along with 200 OK

You can also include an additional field, `proxy`, and specify a custom domain to use in place of HFS_ENDPOINT.
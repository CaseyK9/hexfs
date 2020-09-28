[![logo-Artboard-1-0-4x.png](https://i.postimg.cc/YSXZmBDk/logo-Artboard-1-0-4x.png)](https://postimg.cc/cKnXV2N5)

# hexFS

Originally intended for personal use, this is a file host software. 

### Features

- Know what's uploaded to your server (SHA256, original IP, etc.) with MongoDB
- Delegation of files to external Google Cloud Storage with encryption at-rest
- Simple access system - either make it private, share with friends or open to the public (though the latter isn't recommended!).
- Bulk delete files by IP, SHA256 hash etc.

### setup notes

.env template

```
# What port to listen on (default: 3000)
HFS_PORT=

# The master key. Required. 
# Has full permissions.
# KEEP THIS PRIVATE!
HFS_MASTER_KEY=

# The standard key. Required. This can be used to upload even when HFS_PUBLIC_MODE=0.
# You can share this with friends because it can't be used to delete files.
HFS_STANDARD_KEY=

# The MongoDB connection URI to use. Required.
HFS_MONGO_CONNECTION_URI=

# The MongoDB database to connect to. Required.
HFS_MONGO_DATABASE=

# If set to 1, this will disable the builtin file extension filter (which rejects any file with an extension inside the list).
# Opens your server to malicious uploads from any user if HFS_PUBLIC_MODE is 1.
HFS_DISABLE_FILE_BLACKLIST=0

# If set to 1, ANYONE on the internet don't need permission to upload. Set this to 1 with caution.
HFS_PUBLIC_MODE=0

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

![logo-Artboard-1-0-4x.png](https://i.postimg.cc/YSXZmBDk/logo-Artboard-1-0-4x.png)
# hexFS

Do-it-yourself modern file host software, written in Golang. An excellent, efficient, and sleek alternative to JavaScript uploaders.


### What can you use it for?

- Your screenshots from ShareX
- Having *your* own platform to store *your* files on, not someone else's. Maybe you don't trust them. Or maybe you have a cool domain you want to run this under.


### About the project

- Store data about files (SHA256, original IP, etc.) with MongoDB
- Delegation of files to external Google Cloud Storage with encryption at-rest. 
- Simple access system - either make it private, share a standard key with friends which allows them to upload, not delete, or open to the public (though the latter isn't recommended!).
- Bulk delete files by IP, SHA256 hash and individual IDs. 
- Built-in global ratelimiter for 4 requests/second. Automatically handles forwarded IPs from Cloudflare.
- Built-in DoS protection by refusing to read files larger than the max file size you provide.
- Built-in file extension filter--protect your users from malicious extensions.
- Metrics server integration with Prometheus. 
- Final executable is small - only about 13 MB in size.
- It's not written in JavaScript. 
- The logo is very cool. :^)

### Some key notes

- You need to do a lot of things yourself, the program will only help you with having a connection between MongoDB, the uploader, and Google Cloud Storage. This means you must have your own nginx setup ready, MongoDB instance, firewall configuration, etc etc.
- hexFS cannot download files on your behalf from the Internet. This is deliberately a security decision.
- hexFS will run completely fine if not containerized, but it's still *recommended*! You could use tmux or systemd to run it. Just throwing out ideas.
- There are no fancy plugins or extensions, or frontend template for that matter. Make them yourself.
- hexFS will always listen on 3030 (main server) and 3031 (metrics). This should not be a problem if you ran it in a container. ;)

### .env template aka the massive fucking configuration

This project requires a `.env` to run. Just put it in the project's root and copy paste this shit in there. You can use the comments to help you figure out what to set. All variables are prefixed with `HFS` so they don't collide with other variables. 

The project *will not read the .env for you*. You must load them manually into the environment.

Also your google credential JSON file should be moved to the project root.

```
# ----------------------
# REQUIRED
# These MUST be set or the program will exit immediately!
# ----------------------

# The master key.  
# IT IS VERY IMPORTANT YOU KEEP THIS PRIVATE. DO NOT SHARE IT WITH
# ANYONE, EVEN YOUR CLOSEST FRIENDS. GIVE THEM THE STANDARD KEY 
# SHOWN BELOW INSTEAD. IF THIS KEY IS COMPROMISED YOU EFFECTIVELY EXPOSE
# ALL CONTROLS, SUCH AS UPLOADS AND BULK DELETION OF FILES. YOU HAVE BEEN WARNED.
HFS_MASTER_KEY=

# The standard key. This can be used to upload even when Public mode is not enabled.
# You can share this with friends because it can't be used to delete files.
HFS_STANDARD_KEY=

# The MongoDB connection URI to use.
HFS_MONGO_CONNECTION_URI=

# What MongoDB database to use.
HFS_MONGO_DATABASE=

# Google Cloud Storage bucket name.
GCS_BUCKET_NAME=

# Base64 encoded AES256 32-bit key. 
# You'll use this to encrypt files using your own key instead of google holding on to them.
# More information: https://cloud.google.com/storage/docs/encryption/customer-supplied-keys
# Generator: https://www.digitalsanctuary.com/aes-key-generator-free 
GCS_SECRET_KEY=

# The JSON format of your service account key file location.
# This should be an absolute path. If you run it in docker
# replace "key.json" with the name of your json key.
GOOGLE_APPLICATION_CREDENTIALS=/etc/opt/hexfs/key.json

# The default URL to use in the response for a successful upload.
# For example, if your host domain is https://files.host.com,
# you set this value to https://files.host.com so that the response 
# shall be something like https://files.host.com/a.png
HFS_ENDPOINT=

# ----------------------
# OPTIONAL
# These can optionally be set.
# ----------------------

# If set to 1, this will disable the built-in file extension filter 
# (which rejects any file with an extension inside the list). If
# you think it needs something added or removed, by all means, go into
# the code, edit it, build the image again and restart the container(s).
# WARNING: This also exposes your server to malicious uploads from any 
# user if HFS_PUBLIC_MODE is 1. Thus it is recommended you leave it set to 0.
# HFS_DISABLE_FILE_BLACKLIST=0

# If set to 1, ANYONE ON THE INTERNET can upload without the standard key. 
# Set this to 1 with caution and if you understand the risks of doing so.
# HFS_PUBLIC_MODE=

# The max size, in bytes, any file can be. Default: 50 MiB.
# HFS_MAX_SIZE_BYTES=

# The container's human-readable nickname. If you use Docker this can
# be used to identify the container in Prometheus.
# If none is given hexFS will generate one for you.
# HFS_CONTAINER_NICKNAME=

# Users will be redirected here if they visit hexFS's server directly (GET /). All other
# routes will still 404 and not redirect anywhere. For example, if you
# have a web panel, you would set that URL here.
# HFS_FRONTEND=
```


### How to run

To get the GCS key in you'll need to use a bind mount. Also make sure to bind hexFS ports (3030 and 3031) to other ports on your system. Here's an example of how you would run it, after building the image.

`sudo docker container run -d -p 127.0.0.1:3030:3030 -p 127.0.0.1:3031:3031 -v /where/you/store/config/locally:/mnt/hexfs --name hexfs --env-file /path/to/.env hexfs`

### Support

[discord server](https://discord.gg/F7RBKh2).

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
- Built-in file extension filter--protect your users from malicious extensions.
- Final executable is small - only about 13 MB in size.
- It's not written in JavaScript. 
- The logo is very cool. :^)

### Some key notes

- You need to do a lot of things yourself, the program will only help you with having a connection between MongoDB, the uploader, and Google Cloud Storage. This means you must have your own nginx setup ready, MongoDB instance, firewall configuration, etc etc.
- hexFS cannot download files on your behalf from the Internet. This is deliberately a security decision.
- hexFS will run completely fine if not containerized, but it's still *recommended*! You could use tmux or systemd to run it. Just throwing out ideas.
- There are no fancy plugins or extensions, or frontend template for that matter. Make them yourself.
- hexFS will always listen on 3030 (main server).

### How to run

To get the GCS key in as well as the favicon, you'll need to use a bind mount. Also make sure to bind hexFS port (3030) to other ports on your system. Here's an example of how you would run it, after building the image.

`/where/you/store/everything/locally` should contain your key and favicon image if you have one. You can also move your .env here if you want.

`sudo docker container run -d -p 127.0.0.1:3030:3030 -v /where/you/store/everything/locally:/mnt/hexfs --name hexfs --env-file .env hexfs`

### Support

[discord server](https://discord.gg/F7RBKh2).
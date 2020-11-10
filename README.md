
  

[![hlogo.png](https://i.postimg.cc/qRz0Bh9M/hlogo.png)](https://postimg.cc/CBT9m1hW)

<a href="https://codeclimate.com/github/vysiondev/hexfs/maintainability"><img src="https://api.codeclimate.com/v1/badges/4dd903ec7420f1d080b6/maintainability" /></a>

# hexFS

A lightweight, speedy file host intended for privte use (but you can make it open if you want). Good with ShareX/MagicCap.
  
### Features
  
- Delegation of files to external Google Cloud Storage with encryption at-rest.   
- Simple access system - either:  
  - make it private,   
  - share a standard key with friends which allows them to upload, not delete,   
  - or open to the public.  
- Final executable is small - only about 16 MB in size.  
- Custom extension whitelist/blacklists.  
- Ratelimiting with Redis.  
- No reliance on a database.  
  
### Run  
  
- Put your key in conf/ as "key.json"  
- Put your config in conf/ as "config.yml" using "conf/example.yml" as a reference  
  
Make sure to bind hexFS port (3030) to other ports on your system. Here's an example of how you would run it, after building the image.  
  
`sudo docker container run -d -p 127.0.0.1:3030:3030 --name hexfs hexfs`  
  
### Support  
  
[Discord Server](https://discord.gg/F7RBKh2)


  

[![hlogo.png](https://i.postimg.cc/qRz0Bh9M/hlogo.png)](https://postimg.cc/CBT9m1hW)

# hexFS

<a href="https://codeclimate.com/github/vysiondev/hexfs/maintainability"><img src="https://api.codeclimate.com/v1/badges/4dd903ec7420f1d080b6/maintainability" /></a>

A lightweight but fast file host intended for private use (but you can make it open if you want). Good with ShareX/MagicCap.
  
### Features
  
- Delegation of files to external Google Cloud Storage with encryption at-rest.   
- Simple access system - either:  
  - make it private (default),
  - or open to the public.  
- Final executable is small - only about 20 MB in size.  
- Custom extension whitelist/blacklists.  
- Ratelimiting with Redis.  
- No reliance on a database.  
- Very fast: 9.2ms to upload a 5MB file and return a response. (on a bad network!)

### Run  
  
- Put your GCS service account JSON key in conf/ as "key.json"  
- Put your config in conf/ as "config.yml" using "conf/example.yml" as a reference  
- Optionally, you can replace "favicon.ico" with your own icon! But, it must have the same name.
  
#### With Docker

Make sure to bind hexFS port (3030) to other ports on your system. Here's an example of how you would run it, after building the image.  
  
`sudo docker container run -d -p 127.0.0.1:3030:3030 --name hexfs hexfs`  

#### Other

Just build and run the executable. Make sure you've set the correct configurations.
# Silo
[![Build Status (Travis)](https://travis-ci.org/unixvoid/silo.svg?branch=master)](https://travis-ci.org/unixvoid/silo)  
Silo is a tool used to aid in hosting your own ACI repository. It scans
a directory and generates metadata to be used in the [app container image discovery process](https://github.com/appc/spec/blob/master/spec/discovery.md)
You can use silo stand alone with a webserver like Nginx or in conjunction
with other tools like [binder](https://github.com/unixvoid/binder) to create a web
frontend and API for your images. You can see silo running in conjunction with
binder over at [cryo.unixvoid.com](https://cryo.unixvoid.com/bin)


## Running silo
There are two main ways to run silo:

1. **ACI/rkt**: ironically we can run silo as an aci image. We have public rkt images
hosted on the site! check them out [here](https://cryo.unixvoid.com/bin/rkt/silo/) or
give us a fetch for 64bit machines!  
`rkt fetch unixvoid.com/silo`.   
This image can be run with rkt manually or you can grab
our handy [service file](https://github.com/unixvoid/silo/blob/master/deps/silo.service)

2. **From Source**: Are we not compiled for your architecture? Wanna hack on the source?  
Lets build and deploy:  
`make dependencies`  
`make run`  
If you want to build an ACI use: `make build_aci`

## Configuration
The configuration is very straightforward, we can take a look at the default
config file and break it down.
```
[silo]
  loglevel        = "debug"            # loglevel, this can be [debug, cluster, info, error]
  port            = 8080               # port to listen on
  content         = "test.com"         # domain that will serve metadata
  domain          = "http://test.com"  # url of the domain that will serve the actual images
  basedir         = "rkt"              # directory where images are stored
  bootstrapdelay  = 1                  # delay in seconds before redis binds (useful in containers/autostart applications)
  polldelay       = 2                  # delay in seconds between filesystem polling times (how fast silo will update its images)

[ssl]
  usetls          = false              # whether or not to run with ssl
  servercert      = deps/test.crt      # path to server cert (if usetls is enabled)
  serverkey       = deps/test.key      # path to server key (if usetls is enabled)

[redis]
  host            = "localhost:6379"   # host and port where redis is listening
  password        = ""                 # password to redis database
  cleanonboot     = true               # whether or not to clean out redis db on boot (useful if silo is the only thing running in the db)
```

The most crucial part of this configuration is the `content` and `domain` sections.
I am going to explain this with how I am using it:
```
  content        = "unixvoid.com"
  domain         = "https://cryo.unixvoid.com/bin"
```
Here you can see I am running all my images as `unixvoid.com` images (for example
`unixvoid.com/silo`, `unixvoid.com/nsproxy`, `unixvoid.com/nginx`) but all my images
are hosted on [https://cryo.unixvoid.com/bin](https://cryo.unixvoid.com/bin) (you can
see this in action by visiting the `rkt` link).  
Here is a snippet of what `https://unixvoid.com?ac-discovery=1` produces:
```
<meta name="ac-discovery" content="unixvoid.com/nginx https://cryo.unixvoid.com/bin/rkt/nginx/nginx-{version}-{os}-{arch}.{ext}">
<meta name="ac-discovery-pubkeys" content="unixvoid.com/nginx https://cryo.unixvoid.com/bin/rkt/pubkey/pubkeys.gpg">
```
Here we can see that silo is serving up the images from `cryo.unixvoid.com/bin` 
even though the images are named as `unixvoid.com`  
If you have any problems setting this up, it may be wise to see how ac-discovery
actually works over on the [appc discovery readme](https://github.com/appc/spec/blob/master/spec/discovery.md)  
If you are running this in conjunction with nginx go read the nginx section for
upsteam settings in the config


## Nginx config
The easiest way to configure silo with your existing infrastructure is to test
for clients looking for the `ac-discovery` url.  I use the following bit in my
nginx configs to achieve this:  

```
if ($request_uri ~ .*ac-discovery.*) {
	return 301 https://drone.unixvoid.com$request_uri;
}
```
This will redirect all traffic with `ac-discovery=1` to `drone.unixvoid.com`
which is where silo is running. Now silo can serve metadata properly

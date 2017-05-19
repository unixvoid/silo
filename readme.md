TODO
----
- Check for filesystem changes to pick up new packages  
  this means updating master:metadata to for insertions/deletes  

- Set upstream and inject headers?  
  many people are going to want the metadata injected on their websites home  
  page.. this means we will have to run silo in front of their existing site  
  and just inject the header metadata..

```
rkt
├── binder
│   ├── binder.aci
│   └── binder.aci.asc
├── cryodns
│   ├── cryodns.aci
│   └── cryodns.aci.asc
└── nsproxy
    ├── nsproxy.aci
    └── nsproxy.aci.asc
```

This would contain the following:

```
master:packages  :: [binder, cryodns, nsproxy]
package:binder   ::
  <meta name="ac-discovery" content="unixvoid.com/binder https://unixvoid.com/rkt/binder/binder-{version}-{os}-{arch}.{ext}"> 
  <meta name="ac-discovery-pubkeys" content="unixvoid.com/binder https://cryo.unixvoid.com/rkt/pubkey/pubkeys.gpg"> 
package:cryodns  ::
  <meta name="ac-discovery" content="unixvoid.com/cryodns https://unixvoid.com/rkt/cryodns/cryodns-{version}-{os}-{arch}.{ext}"> 
  <meta name="ac-discovery-pubkeys" content="unixvoid.com/cryodns https://unixvoid.com/rkt/pubkey/pubkeys.gpg"> 
package:nsproxy  ::
  <meta name="ac-discovery" content="unixvoid.com/nsproxy https://unixvoid.com/rkt/nsproxy/nsproxy-{version}-{os}-{arch}.{ext}"> 
  <meta name="ac-discovery-pubkeys" content="unixvoid.com/nsproxy https://unixvoid.com/rkt/pubkey/pubkeys.gpg"> 
```

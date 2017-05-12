TODO
----
On startup check the 'basedir'. We should populate the following structures with information
master:packages (redis set) a set of package names in the directory
package:<name>  (redis key) a key with the value being the meta tag to be used

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

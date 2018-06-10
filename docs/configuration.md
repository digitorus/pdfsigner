# Configuration file 

Configuration file allows to define the following settings:

- path to license file
- signers to be used in commands such as:
  - `pdfsigner sign signer`
  - `pdfsigner watch signer`
  - `pdfsigner serve signer`
  - `pdfsigner serve multiple-signers`
  - `pdfsigner services`
- services to be used with `pdfsigner services command`


The configuration file could be in json, yaml, toml format config files.

## Basic settings

`licensePath` allows to set the path to license file

## Signers settings

The config file should contain multiple signers as an array.

### Signer settings

`name` - name of the signer, to allow identify the signer for the commands and by consumers of the Web API.
`type` - type of the signer, allowed settings "pem" or "pksc11"

PEM specific settings:
`crtPath` - path to certificate file
`keyPath` - path to private key file

PKSC11 specifc settings:
`libPath` - path to library
`pass` - password

signature settings are provided inside `signature` section
`approval` - defines if the signature is approval or not, allowed values `true` and `false`
`certType` - defines cert type???

signature information settings are provided inside `signature.info` section
`name` - name of the person creating signature
`location` - location of the person creating signature
`reason` - reason why the signature is created
`contactInfo` - contact finformation

## Services settings

Services setting is used only for `pdfsigner services` command.

The config file should contain multiple services as an array.

### Service settings

`name` - name of the service
`type` - type of the service, allowed values are: `watch` or `serve`

Watch specific setting:
`signer` - signer name
`in` - folder to watch
`out` - folder where signed files going to be stored

Serve specific setting: 
`signers` - array of signer names
`addr` - address to serve on
`port` - port to serve on


## Example using Toml


```toml
licensePath = "./pdfsigner.lic"

# signers
[[signer]]
name = "simple"
type = "pem"
crtPath = "testfiles/test.crt"
keyPath = "testfiles/test.pem"
[signer.signature]
approval = false
certType = 2
[signer.signature.info]
name = "Tim"
location = "Spain"
reason = "Test"
contactInfo = "None"

[[signer]]
name = "simple2"
type = "pksc11"
libPath = "path/to/lib"
pass = "path/to/lib"
crtChainPath = "path/to/certificate/chain"
[signer.signature]
approval = true
certType = 1
[signer.signature.info]
name = "name"
location = "location"
reason = "reason"
contactInfo = "contact"

[[service]]
name = "watch1"
type = "watch"
signer = "simple"
in = "/Users/tim/go/src/bitbucket.org/digitorus/pdfsigner/temp/in/"
out = "/Users/tim/go/src/bitbucket.org/digitorus/pdfsigner/temp/out/"

[[service]]
name = "web1"
type = "serve"
signers = ["simple"]
addr = "127.0.0.1"
port = 3000

```

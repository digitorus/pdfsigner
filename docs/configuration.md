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

signature settings are provided inside `signData.signature` section
`certType` - defines certificate type. Allowed values are: 
  - `1` - Approval signature
  - `2` - Certification signature. Certification signature required to provide the `docmdp` setting as well.  
  - `3` - UsageRightsSignature.
  - `4` - TimeStampSignature.
`docmdp` - defines the certification signature type. Allowd values are:
  - `1` - Do not allow any changes
  - `2` - Allow filling in existing form fields and signatures
  - `3` - Allow filling in existing form fields and signatures and annotation creation, deletion, and modification. 

signature information settings are provided inside `signData.signature.info` section
`name` - name of the person creating signature
`location` - location of the person creating signature
`reason` - reason why the signature is created
`contactInfo` - contact finformation

## Services settings

Services setting is used only for `pdfsigner services` command.

The config file should contain multiple services as an array.

### Service settings

`name` - name of the service
`validateSignature` - defines weather to validate or not signature after sign, allowed values are: `true` and `false`
`type` - type of the service, allowed values are: `watch` and `serve`


Watch specific setting:
`signer` - signer name
`in` - folder to watch
`out` - folder where signed files going to be stored

Serve specific setting: 
`signers` - array of signer names
`addr` - address to serve on
`port` - port to serve on


## Example using YAML

```yaml
# Main configuration
licensePath: ./pdfsigner.lic

# Common signature settings (anchor)
.signature_defaults: &signature_defaults
  docMDP: 1
  certType: 1

# Common signature info (anchor)
.signature_info_defaults: &signature_info_defaults
  name: Company Name
  location: New York
  reason: Document approval
  contactInfo: support@example.com

# Services Configuration
services:
  watch_incoming:
    type: watch
    signer: company_cert # Reference to the signer configuration below
    in: ./incoming # Where to look for new PDFs
    out: ./signed # Where to put signed PDFs
    validateSignature: true # Verify signature after signing

  api_endpoint:
    type: serve
    signers:
      - company_cert # List of allowed signers
    addr: 127.0.0.1 # Listen address
    port: 3000 # Listen port
    validateSignature: true

# Signers Configuration
signers:
  company_cert:
    type: pem
    crtPath: /path/to/certificate.crt
    keyPath: /path/to/private.key
    crtChainPath: /path/to/chain.pem
    signData:
      signature:
        <<: *signature_defaults # Reuse common signature settings
        info:
          <<: *signature_info_defaults # Reuse common info settings

  hardware_token:
    type: pkcs11
    libPath: /usr/lib/softokn3.so
    pass: token_password
    crtChainPath: /path/to/chain.pem
    signData:
      signature:
        <<: *signature_defaults # Reuse common signature settings
        info:
          name: Hardware Token
          location: Secure Element
          reason: Secure signing
          contactInfo: security@company.com

```

Usage:

```sh
pdfsigner services --config config.example.yaml
```
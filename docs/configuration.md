# Configuration file 

Configuration file allows to define the following settings:

- path to license file
- signers to be used in commands such as:
  - `pdfsigner sign signer`
  - `pdfsigner watch signer`
  - `pdfsigner serve signer`
  - `pdfsigner serve signers`
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

Both signer types accept an optional `crtChainPath` - a PEM bundle of
intermediate/root CA certificates used to resolve the certificate chain,
needed when the leaf certificate isn't self-signed.

Signature settings are provided inside the `options` section:

`certType` - defines certificate type. Allowed values:
  - `0` - Approval signature
  - `1` - Certification signature (requires docMDP)
  - `2` - DocumentTimestamp signature

`docMDP` - defines certification signature restrictions:
  - `1` - No changes allowed
  - `2` - Allow form filling and signatures
  - `3` - Allow form filling, signatures and annotations

TSA settings (optional):
`tsaUrl` - URL of Time Stamping Authority
`tsaUsername` - TSA authentication username
`tsaPassword` - TSA authentication password

Signature information settings, also inside `options`:
`name` - name of the person creating signature
`location` - location of the person creating signature
`reason` - reason why the signature is created
`contactInfo` - contact finformation

### Visual signature appearance (optional)

`options.appearance` draws a visible signature widget on the page. Omit it
to sign without a visible widget (the default).

`page` - 1-indexed page number the widget is placed on
`x`, `y` - lower-left corner of the widget, in PDF points
`width`, `height` - size of the widget, in PDF points
`image` - path to a PNG/JPEG stamp image to draw instead of the standard
  name/reason/location/date text layout; omit to use the text layout

Note: pdfsign only allows visible appearances on approval signatures
(`certType: 0`).

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
    # crtChainPath: /path/to/chain.pem # only needed if the leaf isn't self-signed
    options:
      <<: *signature_defaults # Reuse common signature settings

  hardware_token:
    type: pksc11
    libPath: /usr/lib/softokn3.so
    pass: token_password
    crtChainPath: /path/to/chain.pem
    options:
      <<: *signature_defaults # Reuse common signature settings
      name: Hardware Token
      location: Secure Element
      reason: Secure signing
      contactInfo: security@company.com

```

Usage:

```sh
pdfsigner services --config config.example.yaml
```

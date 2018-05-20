# Description

PDF signer is a multi purpose PDF signer and verifier application.


## Before starting


## Available commands

PDF signer allows to use it in many different ways, as a command line tool, watch folders for files to be signed, to use it as a Web API, and to use multiple services in combination.


## Complete list of commands

- sign
- verify            
- watch             
- serve             
- multiple-services
- license - allows to view the information about the license and 
- version           
- help 


## Command line signer

`pdfsigner sign` command allows to sign document using PEM or PKSC11 as well as using preconfigured signer from the config file.


### Using PEM

`pdfsigner sign pem` 

specific flags: 


```
--key string             Private key path
--crt string             Certificate path

```

#### Example

```
go run main.go sign pem \
  --crt path/to/certificate \
  --key path/to/private/key \
  --chain path/to/certificate/chain \
  --approval true \
  --info-contact "Contact information" \
  --info-location "Location" \
  --info-name "Name" \
  --info-reason "Reason" \
  --tsa-url "http://timestamp-authority.org" \
  --tsa-username "timestamp-authority-username" \
  --tsa-password "timestamp-authority-password" \
  --type 1 \
  path/to/file.pdf 
```


### Using PKSC11

`pdfsigner sign pksc11` 

specific flags:

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```

#### Example

```
pdfsigner sign pksc11 \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
  --crt path/to/certificate \
  --key path/to/private/key \
  --chain path/to/certificate/chain \
  --approval true \
  --info-contact "Contact information" \
  --info-location "Location" \
  --info-name "Name" \
  --info-reason "Reason" \
  --tsa-url "http://timestamp-authority.org" \
  --tsa-username "timestamp-authority-username" \
  --tsa-password "timestamp-authority-password" \
  --type 1 \
  path/to/file.pdf 
```

### With preconfigured signer

> More information on the config file later


`pdfsigner sign signer`


### Example

`pdfsigner sign signer --signer-name signerNameFromTheConfig path/to/file.pdf`

specific flags:

```
--signer-name "name-of-the-signer-from-the-config-file"

```

Preconfigured signer settings could be overwritten with flags:

```
pdfsigner sign signer --signer-name "name-of-the-signer" \
  --crt path/to/certificate \
  --key path/to/private/key \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
  --crt path/to/certificate \
  --key path/to/private/key \
  --chain path/to/certificate/chain \
  --approval true \
  --info-contact "Contact information" \
  --info-location "Location" \
  --info-name "Name" \
  --info-reason "Reason" \
  --tsa-url "http://timestamp-authority.org" \
  --tsa-username "timestamp-authority-username" \
  --tsa-password "timestamp-authority-password" \
  --type 1 \
  path/to/file.pdf 
```

Depending on the type of the signer appropriate flags should be used:

PEM:

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```

PKSC11

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```



## Web API

`pdfsigner serve` allows to run Web API using flags as well as signers from the config file







## Configuration file


Configuration file allows 


```
name = "simple"
type = "pem"
crtPath = "path/to/crt"
keyPath = "path/to/private/key"
crtChainPath = "path/to/certificate/chain"
[signer.signature]
approval = true
certType = 1
[signer.signature.info]
name = "name"
location = "location"
reason = "reason"
contactInfo = "contact"
```


### Shared flags

`pdfsigner sign` flags that are available for all subcommands:



```
--approval               Approval
--chain string           Certificate chain
--crt string             Certificate path
--info-contact string    Signature info contact
--info-location string   Signature info location
--info-name string       Signature info name
--info-reason string     Signature info reason
--tsa-password string    TSA password
--tsa-url string         TSA url
--tsa-username string    TSA username
--type uint              Certificate type (default 1)
```

`pdfsigner sign pem` specific flags

```
--key string             Private key path
--crt string             Certificate path

```

`pdfsigner sign pksc11` specific flags: 














# Description

PDF signer is a multi purpose PDF signer and verifier application.


## Before starting


## Available commands

PDF signer allows to use it in many different ways, as a command line tool, watch folders for new files and save the signed file, to use it as a Web API, and to use multiple services in combination.


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

`pdfsigner sign` allows to sign document using PEM or PKSC11 as well as using preconfigured signer from the config file.

### Example:

`pdfsigner sign signer --out /path/to/new`

### Shared flags

`pdfsigner sign` flags that are available for all subcommands:

Required flags

`--out string             Output path`



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

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```













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
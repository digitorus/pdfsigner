# Command line signer

Command line signer allows to sign document using PEM or PKSC11 provided directly as well as using preconfigured signer from the config file.

Command - `pdfsigner sign`  


## Run with PEM

`pdfsigner sign pem` 

specific flags: 


```
--key string             Private key path
--crt string             Certificate path

```

### Example

```
pdfsigner sign pem \
  --crt path/to/certificate \
  --key path/to/private/key \
  --chain path/to/certificate/chain \
  --contact "Contact information" \
  --location "Location" \
  --name "Name" \
  --reason "Reason" \
  --tsa-url "http://timestamp-authority.org" \
  --tsa-username "timestamp-authority-username" \
  --tsa-password "timestamp-authority-password" \
  --type 1 \
  --docmdp 1 \
  --validate-signature true \
  path/to/file.pdf 
```


## Run with PKSC11

`pdfsigner sign pksc11` 

specific flags:

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```

### Example

```
pdfsigner sign pksc11 \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
  --crt path/to/certificate \
  --key path/to/private/key \
  --chain path/to/certificate/chain \
  --contact "Contact information" \
  --location "Location" \
  --name "Name" \
  --reason "Reason" \
  --tsa-url "http://timestamp-authority.org" \
  --tsa-username "timestamp-authority-username" \
  --tsa-password "timestamp-authority-password" \
  --type 1 \
  --docmdp 1 \
  --validate-signature true \
  path/to/file.pdf 
```

## Run with preconfigured signer

[More information about config file](configuration.md)

`pdfsigner sign signer`

```
--config string          Path to config file
--signer-name string     Signer name
```

### Example

`pdfsigner sign signer --config path/to/config/file --signer-name signerNameFromTheConfig path/to/file.pdf`

specific flags:



Preconfigured signer settings could be overwritten with flags:

```
pdfsigner sign signer --config path/to/config/file --signer-name "name-of-the-signer" \
  --crt path/to/certificate \
  --key path/to/private/key \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
  --crt path/to/certificate \
  --key path/to/private/key \
  --chain path/to/certificate/chain \
  --contact "Contact information" \
  --location "Location" \
  --name "Name" \
  --reason "Reason" \
  --tsa-url "http://timestamp-authority.org" \
  --tsa-username "timestamp-authority-username" \
  --tsa-password "timestamp-authority-password" \
  --type 1 \
  --docmdp 1 \
  --validate-signature true \
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

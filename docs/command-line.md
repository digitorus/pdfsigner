# Command line signer

`pdfsigner sign` command allows to sign document using PEM or PKSC11 as well as using preconfigured signer from the config file.


## Run with PEM

`pdfsigner sign pem` 

specific flags: 


```
--key string             Private key path
--crt string             Certificate path

```

### Example

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

## Run with preconfigured signer

[More inrofmation about config file](./configuration-file.md)

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

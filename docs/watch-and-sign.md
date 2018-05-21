# Watch and sign

`pdfsigner watch` command allows to watch folder for new documents, sign it using PEM or PKSC11 or preconfigured signer from the config file and put signed files into another folder.

specific flags: 

```
--in string              Input path
--out string             Output path
```


## Run with PEM

`pdfsigner watch pem` 

PEM specific flags: 


```
--key string             Private key path
--crt string             Certificate path

```

### Example

```
pdfsigner watch --in path/to/folder/to/watch --out path/to/folder/with/signed/files pem \
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
  --type 1 
```


## Run with PKSC11

`pdfsigner watch pksc11` 

PKSC11 specific flags:

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```

### Example

```
pdfsigner watch pksc11 \
  --in path/to/folder/to/watch \
  --out path/to/folder/with/signed/files \
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
  --type 1
```

## Run with preconfigured signer

[More inrofmation about config file](./configuration-file.md)

`pdfsigner watch signer`

specific flags:

```
--config string          Path to config file
--signer-name string     Signer name
```


### Example

```
pdfsigner watch signer \
--config path/to/config/file \
--signer-name signerNameFromTheConfig \
```

Preconfigured signer settings could be overwritten with flags:

```
pdfsigner watch signer \
  --config path/to/config/file \
  --signer-name "name-of-the-signer" \
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
  --type 1
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



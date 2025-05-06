# Watch and sign

PDFSigner allows to watch folder for new PDF documents, sign it using PEM or PKSC11 or preconfigured signer from the config file and put signed files into specified folder.

Command: `pdfsigner watch`

specific flags: 

```sh
--in string              # Input path
--out string             # Output path
```


## Run with PEM

`pdfsigner watch pem` 

PEM specific flags: 


```sh
--key string             # Private key path
--cert string             # Certificate path

```

### Example

```sh
pdfsigner watch --in path/to/folder/to/watch --out path/to/folder/with/signed/files pem \
  --cert path/to/certificate \
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
  --validate-signature true
```


## Run with PKSC11

`pdfsigner watch pksc11` 

PKSC11 specific flags:

```sh
--lib string             # Path to PKCS11 library
--pass string            # PKCS11 password

```

### Example

```sh
pdfsigner watch pksc11 \
  --in path/to/folder/to/watch \
  --out path/to/folder/with/signed/files \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
  --cert path/to/certificate \
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
  --validate-signature true
```

## Run with preconfigured signer

[More information about config file](configuration.md)

`pdfsigner watch signer`

specific flags:

```sh
--config string          # Path to config file
--signer-name string     # Signer name
```


### Example

```sh
pdfsigner watch signer \
--config path/to/config/file \
--signer-name signerNameFromTheConfig \
```

Preconfigured signer settings could be overwritten with flags:

```sh
pdfsigner watch signer \
  --config path/to/config/file \
  --signer-name "name-of-the-signer" \
  --cert path/to/certificate \
  --key path/to/private/key \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
  --cert path/to/certificate \
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
  --validate-signature true
```

Depending on the type of the signer appropriate flags should be used:

**PEM:**

```sh
--lib string             # Path to PKCS11 library
--pass string            # PKCS11 password

```

**PKSC11**

```sh
--lib string             # Path to PKCS11 library
--pass string            # PKCS11 password
```



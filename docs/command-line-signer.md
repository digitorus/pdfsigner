# Command line signer

Command line signer allows to sign document using PEM or PKSC11 provided directly as well as using preconfigured signer from the config file.

Command - `pdfsigner sign`  


## Run with PEM

`pdfsigner sign pem` 

specific flags: 

```sh
--key string             # Private key path
--cert string             # Certificate path

```

### Example

```sh
pdfsigner sign pem \
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
  --validate-signature true \
  path/to/file.pdf 
```


## Run with PKSC11

`pdfsigner sign pksc11` 

specific flags:

```sh
--lib string             # Path to PKCS11 library
--pass string            # PKCS11 password

```

### Example

```sh
pdfsigner sign pksc11 \
  --lib path/to/pksc11/lib \
  --pass "pksc11-password" \
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

```sh
--config string          # Path to config file
--signer string          # Signer name
```

### Example

```sh
pdfsigner sign signer --config path/to/config/file --signer signerNameFromTheConfig path/to/file.pdf
```

specific flags:

Preconfigured signer settings could be overwritten with flags:

```sh
pdfsigner sign signer \
  --config path/to/config/file \
  --signer-name "name-of-the-signer" \
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
  --validate-signature true \
  path/to/file.pdf 
```

Depending on the type of the signer appropriate flags should be used:

**PEM:**

```sh
--key string             # Private key path
--cert string            # Certificate path
--chain string           # Certificate chain path
```

**PKSC11**

```sh
--lib string             # Path to PKCS11 library
--pass string            # PKCS11 password
--chain string           # Certificate chain path
```

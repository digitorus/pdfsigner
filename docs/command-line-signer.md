# Command line signer

Command line signer allows to sign document using PEM or PKSC11 provided directly as well as using preconfigured signer from the config file.

Command - `pdfsigner sign`  


## Run with PEM

`pdfsigner sign pem` 

specific flags: 

```sh
--key string             # Private key path
--crt string             # Certificate path

```

### Example

```sh
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
--signer-name string     # Signer name
```

### Example

```sh
pdfsigner sign signer --config path/to/config/file --signer-name signerNameFromTheConfig path/to/file.pdf
```

specific flags:

Preconfigured signer settings could be overwritten with flags:

```sh
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

**PEM:**

```sh
--key string             # Private key path
--crt string             # Certificate path

```

**PKSC11**

```sh
--lib string             # Path to PKCS11 library
--pass string            # PKCS11 password
```

## Visual signature appearance

Any `pdfsigner sign` command accepts flags to draw a visible signature
widget on the page instead of an invisible signature. Only approval
signatures (`--type 0`) support a visible appearance.

```sh
--visible                     # Draw a visual signature widget on the page
--appearance-page int         # Page for the widget, 1-indexed (default 1)
--appearance-x float          # X position, in points (default 0)
--appearance-y float          # Y position, in points (default 0)
--appearance-width float      # Width, in points (default 200)
--appearance-height float     # Height, in points (default 80)
--appearance-image string     # PNG/JPEG to draw instead of the standard
                               # name/reason/location/date text layout
```

### Example

```sh
pdfsigner sign pem \
  --crt path/to/certificate \
  --key path/to/private/key \
  --name "Name" \
  --reason "Reason" \
  --type 0 \
  --visible \
  --appearance-x 20 \
  --appearance-y 20 \
  --appearance-width 200 \
  --appearance-height 80 \
  path/to/file.pdf
```

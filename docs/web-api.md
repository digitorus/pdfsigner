# Web API

## Usage

### Scheduling signing job

Scheduling job is done with `POST /sign` [multipart/form-data](https://en.wikipedia.org/wiki/MIME) request with fields and files provided as parts.

The request should contain a `signer` field which defines which signer to use and one or more files.

The Web API is setup with default signature information which could be overwritten using folowing fields: 

- `approval` - defines if the signature is approval or not, allowed values `true` and `false`
- `certType` - defines cert type???
- `name` - name of the person creating signature
- `location` - location of the person creating signature
- `reason` - reason why the signature is created
- `contactInfo` - contact finformation

The request returns JSON `{"job_id":"jobidstr"}` that contains job id which could be used to get information about the job and download signed files


### Getting the status of the job

GET /sign/jobid

GET /sign/jobid/taskid/download



more information about 

## Commands

`pdfsigner serve` allows to run Web API to sign documents with PEM or PKSC11 flags as well as preconfigured signers from the config file.


serve specific flags:

```
--serve-address string   serve address
--serve-port string      serve port
```


### Run with PEM

`pdfsigner serve pem` 

PEM specific flags: 

```
--key string             Private key path
--crt string             Certificate path
```

#### Example

```
pdfsigner serve pem \
  --serve-address "127.0.0.1"\
  --serve-port "8080"\
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


### Run with PKSC11

`pdfsigner serve pksc11` 

PKSC11 specific flags:

```
--lib string             Path to PKCS11 library
--pass string            PKCS11 password

```

#### Example

```
pdfsigner serve pksc11 \
  --serve-address "127.0.0.1"\
  --serve-port "8080"\
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

### Using preconfigured signer

[More inrofmation about config file](./configuration-file.md)

`pdfsigner serve signer`

specific flags:

```
--config path/to/config/file 
--signer-name "name-of-the-signer-from-the-config-file"
```

#### Example

```
pdfsigner serve signer \
  --signer-name signerNameFromTheConfig   
  --serve-address "127.0.0.1"\
  --serve-port "8080"
```

Preconfigured signer settings could be overwritten with flags:

```
pdfsigner serve signer --signer-name "name-of-the-signer" \
  --signer-name signerNameFromTheConfig   
  --serve-address "127.0.0.1"\
  --serve-port "8080"
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


### Using multiple preconfigured signers

Multiple preconfigured signers 


`pdfsigner serve multiple-signers` command allows the consumer of the Web API to choose which signer to use.


#### Example

```
pdfsigner serve multiple-signers signer1 signer2 signer3 \
  --serve-address "127.0.0.1"\
  --serve-port "8080"
```


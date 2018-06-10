# Web API

Web API allows to sign and verify files by communicating with the application using HTTP protocol

## Usage

### How it works

To process files using Web API first a job that contains one or more files needs to be put into the queue. After that the status of the job that contains pending, successful or failed tasks(files) should be requested. In case of signing job after files were processed successfully they could be downloaded.

Available end points:

`POST /sign` - put one or more files with specified signer into the signing queue 
`GET /sign/jobid` - get status of the job with tasks
`GET /sign/jobid/taskid/download` - download completed file

`POST /verify` - put one or more files into the verification queue  
`GET /verify/jobid` - get status of the job with tasks


### Signing

#### Schedule signing job

Scheduling job is done with `POST /sign` [multipart/form-data](https://developer.mozilla.org/en-US/docs/Web/API/FormData/Using_FormData_Objects) request with fields and files provided as parts.

The request should contain a `signer` field which defines which signer to use and one or more files.

The Web API is setup with default signature information which could be overwritten using folowing fields: 

- `approval` - defines if the signature is approval or not, allowed values `true` and `false`
- `certType` - defines cert type???
- `name` - name of the person creating signature
- `location` - location of the person creating signature
- `reason` - reason why the signature is created for
- `contactInfo` - contact finformation

The successful request returns JSON `{"job_id":"jobidstr"}` that contains job id which could be then used to get information about the tasks and to download signed files.

It also may return JSON formatted error Ex.`{"error":"no files provided","code":400}`. 

That error may only contain the error of putting a job to the queue, not the signing results. Signing results could be obtained with `GET /sign/jobid` request.


#### Get status of the signing job

Getting the status of the job is done using `GET /sign/jobid` request which returns the tasks associated with the job, every task contains it's id, original file name, and status that could be "Pending" - the task is not processed yet and Failed. When the task is going to fail it's going to contain the error.

The request may fail with JSON response. Ex: `{"error":"job doesn't exists","code":400}`

Job with successfully completed task:

```json
{
	"job":{"id":"bc5g4tl2m9sn837gm00g"},
	"tasks":[
		{
			"id":"bc5g4tl2m9sn837gm010",
			"file_name":"testfile12.pdf",
			"status":"Completed"
		}
	]
}
```

Job with failed to complete task:

```json
{
	"job":{"id":"bc5g4tl2m9sn837gm00g"},
	"tasks":[
		{
			"id":"bc5g4tl2m9sn837gm010",
			"file_name":"testfile12.pdf",
			"status":"Failed",
			"error": "malformed pdf file"
		}
	]
}
```

#### Download signed files

Signed files could be downloaded with `GET /sign/jobid/taskid/download` request. 

The response would be the file.

Note: postman offers to download file as download.pdf instead of the file name provided with content disposition, see issue: https://github.com/postmanlabs/postman-app-support/issues/2082

The request may fail with JSON response. Ex: `{"error":"task is not found","code":400}`

__
### Verifying

#### Schedule verifying job

Scheduling job is done with `POST /verify` [multipart/form-data](https://developer.mozilla.org/en-US/docs/Web/API/FormData/Using_FormData_Objects) request with files provided as parts.

The successful request returns JSON `{"job_id":"jobidstr"}` that contains job id which could be then used to get information about the tasks and to download signed files.

It also may return JSON formatted error Ex.`{"error":"no files provided","code":400}`. 

That error may only contain the error of putting a job to the queue, not the signing results. Signing results could be obtained with `GET /verify/jobid` request.


#### Get status of the verifying job

Getting the status of the job is done using `GET /sign/jobid` request which returns the tasks associated with the job, every task contains it's id, original file name, and status that could be "Pending" - the task is not processed yet and Failed. When the task is going to fail it's going to contain the error.

The request may fail with JSON response. Ex: `{"error":"job doesn't exists","code":400}`

Job with successfully completed task:

```json
{
	"job":{"id":"bc5g4tl2m9sn837gm00g"},
	"tasks":[
		{
			"id":"bc5g4tl2m9sn837gm010",
			"file_name":"testfile12.pdf",
			"status":"Completed"
		}
	]
}
```

Job with failed to complete task:

```json
{
	"job":{"id":"bc5g4tl2m9sn837gm00g"},
	"tasks":[
		{
			"id":"bc5g4tl2m9sn837gm010",
			"file_name":"testfile12.pdf",
			"status":"Failed",
			"error": "malformed pdf file"
		}
	]
}
```



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

[More inrofmation about config file](configuration.md)

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


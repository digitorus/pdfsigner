# Command line verifier

`pdfsigner verify` command allows to verify PDF document.


## Run with PEM

`pdfsigner sign pem` 

specific flags: 


```
--key string             Private key path
--crt string             Certificate path

```

### Example

```
pdfsigner verify path/to/file.pdf path/to/file2.pdf
```

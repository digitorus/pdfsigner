# Services

Services command allows to run multiple watch and Web API services

Command - `pdfsigner services` 

services specific flags:

```
--config path/to/config/file 
```

### Run all the services from the config file

```
pdfsigner services
```

### Run specific services from the config

[More information about config file](configuration.md)


```
pdfsigner services service1 service2
```

#### Override config settings with CLI

It's possible to override signers and services settings provided with the config using CLI. 



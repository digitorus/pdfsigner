module github.com/digitorus/pdfsigner

go 1.18

require (
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/digitorus/pdf v0.1.2
	github.com/digitorus/pdfsign v0.0.0-20220715153233-d3f4c7735954
	github.com/digitorus/pkcs11 v0.0.0-20220708123826-82d5d203495a
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-test/deep v1.0.8
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gtank/cryptopasta v0.0.0-20170601214702-1f550f6f2f69
	github.com/hyperboloide/lk v0.0.0-20221004131154-cb9733bc66d0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.1
	go.etcd.io/bbolt v1.3.6
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/digitorus/pkcs7 v0.0.0-20220704143225-a9c8106cbfc6 // indirect
	github.com/digitorus/timestamp v0.0.0-20221019073249-0b6a45065722 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattetti/filebuffer v1.0.1 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/digitorus/pkcs11 v0.0.0-20220705083045-3847d33b47af => ../pkcs11

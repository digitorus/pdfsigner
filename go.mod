module github.com/digitorus/pdfsigner

go 1.18

require (
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/digitorus/pdf v0.1.2
	github.com/digitorus/pdfsign v0.0.0-20220704151736-5a4ccb37cce6
	github.com/digitorus/pkcs11 v0.0.0-20220708123826-82d5d203495a
	github.com/fsnotify/fsnotify v1.5.4
	github.com/go-test/deep v1.0.8
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gtank/cryptopasta v0.0.0-20170601214702-1f550f6f2f69
	github.com/hyperboloide/lk v0.0.0-20200504060759-b535f1973118
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.2
	go.etcd.io/bbolt v1.3.6
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/digitorus/pkcs7 v0.0.0-20220704143225-a9c8106cbfc6 // indirect
	github.com/digitorus/timestamp v0.0.0-20220704143351-8225fba02d52 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattetti/filebuffer v1.0.1 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sys v0.0.0-20220708085239-5a0f0661e09d // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/digitorus/pkcs11 v0.0.0-20220705083045-3847d33b47af => ../pkcs11

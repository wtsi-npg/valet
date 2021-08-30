module github.com/wtsi-npg/valet

go 1.17

require (
	github.com/klauspost/compress v1.9.1 // indirect
	github.com/klauspost/pgzip v1.2.5
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.24.0
	github.com/spf13/cobra v0.0.7
	github.com/stretchr/testify v1.7.0
	github.com/wtsi-npg/extendo/v2 v2.4.0
	github.com/wtsi-npg/fsnotify v1.4.8-0.20190705153444-45ca73e9793a
	github.com/wtsi-npg/logshim v1.3.0
	github.com/wtsi-npg/logshim-zerolog v1.3.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

// replace github.com/wtsi-npg/extendo/v2 => ../extendo

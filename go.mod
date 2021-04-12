module github.com/wtsi-npg/valet

go 1.14

require (
	github.com/klauspost/compress v1.9.1 // indirect
	github.com/klauspost/pgzip v1.2.5
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.21.0
	github.com/spf13/cobra v0.0.7
	github.com/stretchr/testify v1.7.0
	github.com/wtsi-npg/extendo/v2 v2.3.0
	github.com/wtsi-npg/fsnotify v1.4.8-0.20190705153444-45ca73e9793a
	github.com/wtsi-npg/logshim v1.2.0
	github.com/wtsi-npg/logshim-zerolog v1.2.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
)

// replace github.com/wtsi-npg/extendo/v2 => ../extendo

module github.com/kjsanger/valet

go 1.13

require (
	github.com/kjsanger/extendo/v2 v2.2.0
	github.com/kjsanger/fsnotify v1.4.8-0.20190705153444-45ca73e9793a
	github.com/kjsanger/logshim v1.1.0
	github.com/kjsanger/logshim-zerolog v1.1.0
	github.com/klauspost/compress v1.9.1 // indirect
	github.com/klauspost/pgzip v1.2.1
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
)

// replace github.com/kjsanger/extendo/v2 => ../extendo

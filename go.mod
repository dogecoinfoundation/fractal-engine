module dogecoin.org/fractal-engine

replace dogecoin.org/chainfollower => ../chainfollower

replace dogecoin.org/dogeclient => ../dogeclient

replace code.dogecoin.org/gossip => github.com/dogeorg/gossip v0.0.18

replace code.dogecoin.org/governor => github.com/dogeorg/governor v1.0.2

go 1.24.0

require (
	code.dogecoin.org/gossip v0.0.0
	dogecoin.org/chainfollower v0.0.0
	dogecoin.org/dogeclient v0.0.0-00010101000000-000000000000
	dogecoin.org/dogetest v0.0.0
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.27
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.5.2
)

require (
	code.dogecoin.org/governor v0.0.0-00010101000000-000000000000 // indirect
	github.com/btcsuite/golangcrypto v0.0.0-20150304025918-53f62d9b43e8 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/crypto/blake256 v1.1.0 // indirect
	github.com/dogeorg/doge v0.0.12 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.36.3 // indirect
	modernc.org/ccgo/v3 v3.16.9 // indirect
	modernc.org/libc v1.17.1 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.2.1 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.18.1 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.0 // indirect
)

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/gin-gonic/gin v1.10.0
	github.com/lib/pq v1.10.9
	github.com/shopspring/decimal v1.4.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
)

replace dogecoin.org/dogetest => ../dogetest

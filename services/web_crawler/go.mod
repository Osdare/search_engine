module web_crawler

go 1.25.1

require (
	github.com/redis/go-redis/v9 v9.14.0
	github.com/reiver/go-porterstemmer v1.0.1
	golang.org/x/net v0.44.0
	utils v0.0.0-20250101000000-deadbeef
)

replace utils => ../../libs/utils


require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

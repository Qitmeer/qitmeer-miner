module hlc-miner

go 1.12

require (
    github.com/HalalChain/qitmeer-lib latest
    github.com/HalalChain/go-opencl latest
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/blake256 v1.0.0
	github.com/deckarep/golang-set v1.7.1
	github.com/go-stack/stack v1.8.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/logrotate v1.0.0
	github.com/mattn/go-colorable v0.1.1
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20190412213103-97732733099d
	golang.org/x/text v0.3.0

)

replace (
	golang.org/x/sync => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190222072716-a9d3bda3a223
	golang.org/x/text => github.com/golang/text v0.3.0
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3

)

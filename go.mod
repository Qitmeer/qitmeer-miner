module qitmeer-miner

go 1.12

require (
	github.com/HalalChain/go-opencl v0.0.0-20190704222003-c93200893312
	github.com/HalalChain/qitmeer-lib v0.0.0-20190708060928-91b89dc718d2

)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/sync => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190222072716-a9d3bda3a223
	golang.org/x/text => github.com/golang/text v0.3.0

)

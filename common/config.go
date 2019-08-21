// Copyright (c) 2019 The qitmeer developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/core/address"
	"github.com/Qitmeer/qitmeer-lib/params"
	"qitmeer-miner/common/go-flags"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultConfigFilename = "qitmeer.conf"
)

var (
	minerHomeDir          = GetCurrentDir()
	defaultConfigFile     = filepath.Join(minerHomeDir, defaultConfigFilename)
	defaultRPCServer      = "127.0.0.1"
	defaultRPCPort      = "1234"
	defaultIntensity = 24
	defaultTrimmerCount = 40
	defaultWorkSize = 256
	minIntensity  = 1
	defaultRpcMinerLog  = GetCurrentDir() + "/miner.log"
	maxIntensity  = 31
	maxWorkSize   = uint32(0xFFFFFFFF - 255)
	defaultPow  ="blake2bd"
	defaultSymbol  ="PMEER"
)

type DeviceConfig struct {
	ListDevices bool `short:"l" long:"listdevices" description:"List number of devices."`
}

type FileConfig struct {
	ConfigFile   string `short:"C" long:"configfile" description:"Path to configuration file"`
	// Debugging options
	MinerLogFile string `long:"minerlog" description:"Write miner log file"`
}

type OptionalConfig struct {
	// Config / log options
	CPUMiner       bool   `long:"cpuminer" description:"CPUMiner" default-mask:"false"`
	Proxy       string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser   string `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass   string `long:"proxypass" default-mask:"-" description:"Password for proxy server"`
	TrimmerCount     int `long:"trimmerTimes" description:"the cuckaroo trimmer times"`
	Intensity         int `long:"intensity" description:"Intensities (the work size is 2^intensity) per device. Single global value or a comma separated list."`
	WorkSize          int `long:"worksize" description:"The explicitly declared sizes of the work to do per device (overrides intensity). Single global value or a comma separated list."`
}

type PoolConfig struct {
	// Pool related options
	Pool         string `short:"o" long:"pool" description:"Pool to connect to (e.g.stratum+tcp://pool:port)"`
	PoolUser     string `short:"m" long:"pooluser" description:"Pool username"`
	PoolPassword string `short:"n" long:"poolpass" default-mask:"-" description:"Pool password"`
}

type SoloConfig struct {
	// RPC connection options
	MinerAddr   string `short:"M" long:"mineraddress" description:"Miner Address" default-mask:""`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username"`
	RPCPassword string `short:"p" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RandStr     string `long:"randstr" description:"Rand String,Your Unique Marking." default-mask:"Come from Qitmeer!"`
	NoTLS       bool   `long:"notls" description:"Do not verify tls certificates" default-mask:"true"`
	RPCCert     string `long:"rpccert" description:"RPC server certificate chain for validation"`
}

type NecessaryConfig struct {
	// Config / log options
	Pow     string `short:"P" long:"pow" description:"blake2bd|cuckaroo|cuckatoo"`
	Symbol      string   `short:"S" long:"symbol" description:"Symbol" default-mask:"PMEER"`
	NetWork      string   `short:"N" long:"network" description:"network privnet|testnet|mainnet" default-mask:"mainnet"`
	Param      *params.Params
}

type GlobalConfig struct {
	OptionConfig OptionalConfig
	LogConfig FileConfig
	DeviceConfig DeviceConfig
	SoloConfig SoloConfig
	PoolConfig PoolConfig
	NecessaryConfig NecessaryConfig
}

// removeDuplicateAddresses returns a new slice with all duplicate entries in
// addrs removed.
func removeDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

// normalizeAddress returns addr with the passed default port appended if
// there is not already a port specified.
func normalizeAddress(addr string, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// normalizeAddresses returns a new slice with all the passed peer addresses
// normalized with the given default port, and all duplicates removed.
func normalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = normalizeAddress(addr, defaultPort)
	}

	return removeDuplicateAddresses(addrs)
}

// cleanAndExpandPath expands environement variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(minerHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in btcd functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options.  Command line options always take precedence.
func LoadConfig() (*GlobalConfig, []string, error) {
	// Default config.
	soloCfg := SoloConfig{
		RPCServer:  defaultRPCServer,
		NoTLS:  true,
	}
	poolCfg := PoolConfig{}
	// Default config.
	deviceCfg := DeviceConfig{}
	// Default config.
	fileCfg := FileConfig{
		//ConfigFile:defaultConfigFile,
		MinerLogFile:  defaultRpcMinerLog,
	}
	necessaryCfg := NecessaryConfig{
		Pow:defaultPow,
		Symbol:defaultSymbol,
	}
	optionalCfg := OptionalConfig{
		Intensity:  defaultIntensity,
		WorkSize:  defaultWorkSize,
		TrimmerCount:  defaultTrimmerCount,
		CPUMiner:  false,
	}

	// Create the home directory if it doesn't already exist.
	err := os.MkdirAll(minerHomeDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.
	preParser := flags.NewNamedParser(appName, flags.HelpFlag)

	_,err = preParser.AddGroup("Debug Command", "The Miner Debug tools", &deviceCfg)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(0)
	}
	_,err = preParser.AddGroup("The Config File Options", "The Config File Options", &fileCfg)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(0)
	}
	_,err = preParser.AddGroup("The Necessary Config Options", "The Necessary Config Options", &necessaryCfg)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(0)
	}
	_,err = preParser.AddGroup("The Solo Config Option", "The Solo Config Option", &soloCfg)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(0)
	}
	_,err = preParser.AddGroup("The pool Config Option", "The pool Config Option", &poolCfg)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(0)
	}
	_,err = preParser.AddGroup("The Optional Config Option", "The Optional Config Option", &optionalCfg)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(0)
	}

	_,err = preParser.Parse()
	if err != nil{
		log.Printf("%v", err)
		log.Println(fmt.Sprintf("Usage to see  ./%s -h",appName))
		os.Exit(0)
	}
	if fileCfg.ConfigFile == ""{
		log.Printf("[warn] Don't have config file.")
	} else {
		err = flags.NewIniParser(preParser).ParseFile(fileCfg.ConfigFile)
		if err != nil {
			if _, ok := err.(*os.PathError); !ok {
				fmt.Fprintln(os.Stderr, err)
				return nil, nil, err
			}
		}
	}


	remainingArgs,err := preParser.Parse()
	if err != nil {
		if _, ok := err.(*flags.Error); !ok {
			fmt.Fprintln(os.Stderr, err)
			return nil, nil, err
		}
		preParser.WriteHelp(os.Stderr)
		os.Exit(0)
	}



	if deviceCfg.ListDevices{
		log.Println("【CPU Devices List】:")
		GetDevices(DevicesTypesForCPUMining)
		log.Println("【GPU Devices List】:")
		GetDevices(DevicesTypesForGPUMining)
		os.Exit(0)
	}
	if poolCfg.Pool == "" && soloCfg.MinerAddr == ""{
		log.Println("[error] Solo need address -M , pool need -o pool address")
		preParser.WriteHelp(os.Stderr)
		os.Exit(0)
	}
	necessaryCfg.Param = InitNet(necessaryCfg.NetWork,necessaryCfg.Param)
	if necessaryCfg.Param == nil{
		os.Exit(0)
	}
	if poolCfg.Pool == "" && !CheckBase58Addr(soloCfg.MinerAddr,necessaryCfg.NetWork,necessaryCfg.Param){
		os.Exit(0)
	}
	// Show the version and exit if the version flag was specified.

	if optionalCfg.Intensity < minIntensity || optionalCfg.Intensity > maxIntensity{
		optionalCfg.Intensity = defaultIntensity
	}

	// Check the work size if the user is setting that.
	if optionalCfg.WorkSize > int(maxWorkSize){
		optionalCfg.WorkSize = defaultWorkSize
	}

	// Handle environment variable expansion in the RPC certificate path.
	soloCfg.RPCCert = cleanAndExpandPath(soloCfg.RPCCert)

	// Add default port to RPC server based on --testnet flag
	// if needed.
	soloCfg.RPCServer = normalizeAddress(soloCfg.RPCServer, defaultRPCPort)

	return &GlobalConfig{
		optionalCfg,
		fileCfg,
		deviceCfg,
		soloCfg,
		poolCfg,
		necessaryCfg,
	}, remainingArgs, nil
}

func CheckBase58Addr(addr ,network string,p *params.Params) bool {
	_,err := address.DecodeAddress(addr)
	if err != nil{
		log.Fatalln(network,"qitmeer address error!",err,addr)
		return false
	}
	networkChar := addr[0:1]
	if p.NetworkAddressPrefix != networkChar{
		log.Fatalln(network,"qitmeer address not match the network,please check your config network param!",p.NetworkAddressPrefix,networkChar)
		return false
	}
	return true
}

func InitNet(network string,p *params.Params) *params.Params {
	switch network {
	case params.MainNetParams.Name:
		p = &params.MainNetParams
	case params.TestNetParams.Name:
		p = &params.TestNetParams
	case params.PrivNetParams.Name:
		p = &params.PrivNetParams
	default:
		log.Fatalln(network,"qitmeer error!")
		return nil
	}
	return p
}

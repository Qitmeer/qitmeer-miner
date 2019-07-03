// Copyright (c) 2019 The halalchain developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/params"
	"hlc-miner/common/go-flags"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultConfigFilename = "halalchainminer.conf"
)

var (
	minerHomeDir          = GetCurrentDir()
	defaultConfigFile     = filepath.Join(minerHomeDir, defaultConfigFilename)
	defaultRPCServer      = "127.0.0.1"
	defaultTrimmerCount = 40
	minIntensity  = 8
	maxIntensity  = 31
	maxWorkSize   = uint32(0xFFFFFFFF - 255)
	ChainParams  *params.Params
)

type Config struct {
	ListDevices bool `short:"l" long:"listdevices" description:"List number of devices."`

	// Config / log options
	Experimental bool   `long:"experimental" description:"enable EXPERIMENTAL features such as setting a temperature target with (-t/--temptarget) which may DAMAGE YOUR DEVICE(S)."`
	ConfigFile   string `short:"C" long:"configfile" description:"Path to configuration file"`
	Pow     string `long:"pow" description:"blake2bd|cuckroo|cucktoo"`
	TrimmerCount     int `long:"trimmerCount" description:"the cuckaroo trimer times"`

	// Debugging options
	Profile    string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	CPUProfile string `long:"cpuprofile" description:"Write CPU profile to the specified file"`
	MemProfile string `long:"memprofile" description:"Write mem profile to the specified file"`
	MinerLogFile string `long:"minerlog" description:"Write miner log file"`

	// RPC connection options
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username"`
	//Dag     	bool `short:"dag" long:"dag" description:"dag mining"`
	RPCPassword string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert     string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	NoTLS       bool   `long:"notls" description:"Do not verify tls certificates"`
	Symbol      string   `long:"symbol" description:"Symbol" default-mask:"NOX"`
	RandStr     string   `long:"randstr" description:"Rand String" default-mask:"Come from halalchain!"`
	MinerAddr   string   `long:"mineraddress" description:"Miner Address" default-mask:""`
	Proxy       string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser   string `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass   string `long:"proxypass" default-mask:"-" description:"Password for proxy server"`

	Intensity         int `short:"i" long:"intensity" description:"Intensities (the work size is 2^intensity) per device. Single global value or a comma separated list."`
	WorkSize          int `short:"W" long:"worksize" description:"The explicitly declared sizes of the work to do per device (overrides intensity). Single global value or a comma separated list."`

	// Pool related options
	Pool         string `short:"o" long:"pool" description:"Pool to connect to (e.g.stratum+tcp://pool:port)"`
	PoolUser     string `short:"m" long:"pooluser" description:"Pool username"`
	PoolPassword string `short:"n" long:"poolpass" default-mask:"-" description:"Pool password"`
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
func LoadConfig() (*Config, []string, error) {
	// Default config.
	cfg := Config{
		ConfigFile: defaultConfigFile,
		RPCServer:  defaultRPCServer,
	}

	// Create the home directory if it doesn't already exist.
	err := os.MkdirAll(minerHomeDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.
	preCfg := cfg
	preParser := flags.NewParser(&preCfg, flags.Default)
	_, err = preParser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			preParser.WriteHelp(os.Stderr)
		}
		return nil, nil, err
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))

	// Load additional config from file.
	var configFileError error
	parser := flags.NewParser(&cfg, flags.Default)
	err = flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Fprintln(os.Stderr, err)
			parser.WriteHelp(os.Stderr)
			return nil, nil, err
		}
		configFileError = err
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		}
		return nil, nil, err
	}

	if cfg.MinerLogFile == ""{
		cfg.MinerLogFile = GetCurrentDir() + "/miner.log"
	}

	if cfg.TrimmerCount <= 0{
		cfg.TrimmerCount = defaultTrimmerCount
	}

	if cfg.Intensity <= 0 || cfg.Intensity < minIntensity || cfg.Intensity > maxIntensity{
		cfg.Intensity = 28
	}

	if cfg.Experimental {
		fmt.Fprintln(os.Stderr, "enabling EXPERIMENTAL features "+
			"that may possibly DAMAGE YOUR DEVICE(S)")
		time.Sleep(time.Second * 3)
	}

	// Check the work size if the user is setting that.
	if cfg.WorkSize <= 0 || cfg.WorkSize > int(maxWorkSize){
		cfg.WorkSize = 256
	}


	// Handle environment variable expansion in the RPC certificate path.
	cfg.RPCCert = cleanAndExpandPath(cfg.RPCCert)

	var defaultRPCPort string

	// Add default port to RPC server based on --testnet flag
	// if needed.
	cfg.RPCServer = normalizeAddress(cfg.RPCServer, defaultRPCPort)

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configFileError != nil {
		log.Printf("%v", configFileError)
	}

	return &cfg, remainingArgs, nil
}

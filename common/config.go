/**
	HLC FOUNDATION
	james
 */

package common

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/noxproject/nox/params"
	"log"
	"hlc-miner/common/go-flags"
)

const (
	defaultConfigFilename = "halalchainminer.conf"
	defaultLogLevel       = "info"
	defaultLogDirname     = "logs"
	defaultLogFilename    = "halalchainminer.log"
	defaultClKernel       = "crypto/blake256/kernel.cl"
)

var (
	minerHomeDir          = AppDataDir("halalchainminer", false)
	noxHomeDir           = AppDataDir("nox", false)
	defaultConfigFile     = filepath.Join(minerHomeDir, defaultConfigFilename)
	defaultRPCServer      = "127.0.0.1"
	defaultRPCCertFile    = filepath.Join(noxHomeDir, "rpc.cert")
	defaultRPCPortMainNet = "1234"
	defaultRPCPortTestNet = "1234"
	defaultRPCPortSimNet  = "1234"
	defaultAPIPort        = "3333"
	defaultLogDir         = filepath.Join(minerHomeDir, defaultLogDirname)
	defaultAutocalibrate  = 500

	minIntensity  = 8
	maxIntensity  = 31
	maxWorkSize   = uint32(0xFFFFFFFF - 255)
	ChainParams  *params.Params
)

type Config struct {
	ListDevices bool `short:"l" long:"listdevices" description:"List number of devices."`
	Version string `short:"V" long:"version" description:"Display version information and exit"`

	// Config / log options
	Experimental bool   `long:"experimental" description:"enable EXPERIMENTAL features such as setting a temperature target with (-t/--temptarget) which may DAMAGE YOUR DEVICE(S)."`
	ConfigFile   string `short:"C" long:"configfile" description:"Path to configuration file"`
	Temp   string `short:"t" long:"temp" description:"temp"`
	LogDir       string `long:"logdir" description:"Directory to log output."`
	DebugLevel   string `short:"d" long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	ClKernel     string `short:"k" long:"kernel" description:"File with cl kernel to use"`

	// Debugging options
	Profile    string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	CPUProfile string `long:"cpuprofile" description:"Write CPU profile to the specified file"`
	MemProfile string `long:"memprofile" description:"Write mem profile to the specified file"`

	// Status API options
	APIListeners []string `long:"apilisten" description:"Add an interface/port to expose miner status API"`

	// RPC connection options
	RPCUser     string `short:"u" long:"rpcuser" description:"RPC username"`
	//Dag     	bool `short:"dag" long:"dag" description:"dag mining"`
	RPCPassword string `short:"P" long:"rpcpass" default-mask:"-" description:"RPC password"`
	RPCServer   string `short:"s" long:"rpcserver" description:"RPC server to connect to"`
	RPCCert     string `short:"c" long:"rpccert" description:"RPC server certificate chain for validation"`
	NoTLS       bool   `long:"notls" description:"Disable TLS"`
	Symbol      string   `long:"symbol" description:"Symbol" default-mask:"NOX"`
	RandStr     string   `long:"randstr" description:"Rand String" default-mask:"Come from halalchain!"`
	MinerAddr   string   `long:"mineraddress" description:"Miner Address" default-mask:""`
	Proxy       string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	ProxyUser   string `long:"proxyuser" description:"Username for proxy server"`
	ProxyPass   string `long:"proxypass" default-mask:"-" description:"Password for proxy server"`

	Benchmark bool `short:"B" long:"benchmark" description:"Run in benchmark mode."`

	TestNet       bool `long:"testnet" description:"Connect to testnet"`
	CPUMiner       bool `long:"cpuminer" default-mask:"true" description:"CPU MINER"`
	DAG       		bool `long:"dag" default-mask:"false" description:"DAG MINER"`
	SimNet        bool `long:"simnet" description:"Connect to the simulation test network"`
	TLSSkipVerify bool `long:"skipverify" description:"Do not verify tls certificates (not recommended!)"`

	Autocalibrate     string `short:"A" long:"autocalibrate" description:"Time target in milliseconds to spend executing hashes on the device during each iteration. Single global value or a comma separated list."`
	AutocalibrateInts []int
	Devices           string `short:"D" long:"devices" description:"Single device ID or a comma separated list of device IDs to use."`
	DeviceIDs         []int
	Intensity         int `short:"i" long:"intensity" description:"Intensities (the work size is 2^intensity) per device. Single global value or a comma separated list."`
	WorkSize          int `short:"W" long:"worksize" description:"The explicitly declared sizes of the work to do per device (overrides intensity). Single global value or a comma separated list."`
	WorkSizeInts      []uint32

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
		DebugLevel: defaultLogLevel,
		LogDir:     defaultLogDir,
		RPCServer:  defaultRPCServer,
		RPCCert:    defaultRPCCertFile,
		ClKernel:   defaultClKernel,
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

	// Multiple networks can't be selected simultaneously.
	numNets := 0
	if cfg.TestNet {
		numNets++
	}
	if cfg.SimNet {
		numNets++
	}
	if numNets > 1 {
		str := "%s: The testnet and simnet params can't be used " +
			"together -- choose one of the two"
		err := fmt.Errorf(str, "LoadConfig")
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	// Check the autocalibrations if the user is setting that.
	if len(cfg.Autocalibrate) > 0 {
		// Parse a list like -A 450,600
		if strings.Contains(cfg.Autocalibrate, ",") {
			specifiedAutocalibrates := strings.Split(cfg.Autocalibrate, ",")
			cfg.AutocalibrateInts = make([]int, len(specifiedAutocalibrates))
			for i := range specifiedAutocalibrates {
				j, err := strconv.Atoi(specifiedAutocalibrates[i])
				if err != nil {
					err := fmt.Errorf("Could not convert autocalibration "+
						"(%v) to int: %s", specifiedAutocalibrates[i],
						err.Error())
					fmt.Fprintln(os.Stderr, err)
					return nil, nil, err
				}

				cfg.AutocalibrateInts[i] = j
			}
			// Use specified device like -A 600
		} else {
			cfg.AutocalibrateInts = make([]int, 1)
			i, err := strconv.Atoi(cfg.Autocalibrate)
			if err != nil {
				err := fmt.Errorf("Could not convert autocalibration %v "+
					"to int: %s", cfg.Autocalibrate, err.Error())
				fmt.Fprintln(os.Stderr, err)
				return nil, nil, err
			}

			cfg.AutocalibrateInts[0] = i
		}
		// Apply default
	} else {
		cfg.AutocalibrateInts = []int{defaultAutocalibrate}
	}

	// Check the devices if the user is setting that.
	if len(cfg.Devices) > 0 {
		// Parse a list like -D 1,2
		if strings.Contains(cfg.Devices, ",") {
			specifiedDevices := strings.Split(cfg.Devices, ",")
			cfg.DeviceIDs = make([]int, len(specifiedDevices))
			for i := range specifiedDevices {
				j, err := strconv.Atoi(specifiedDevices[i])
				if err != nil {
					err := fmt.Errorf("Could not convert device number %v "+
						"(%v) to int: %s", i+1, specifiedDevices[i],
						err.Error())
					fmt.Fprintln(os.Stderr, err)
					return nil, nil, err
				}

				cfg.DeviceIDs[i] = j
			}
			// Use specified device like -D 1
		} else {
			cfg.DeviceIDs = make([]int, 1)
			i, err := strconv.Atoi(cfg.Devices)
			if err != nil {
				err := fmt.Errorf("Could not convert specified device %v "+
					"to int: %s", cfg.Devices, err.Error())
				fmt.Fprintln(os.Stderr, err)
				return nil, nil, err
			}

			cfg.DeviceIDs[0] = i
		}
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

	// Initialize log rotation.  After log rotation has been initialized,

	if len(cfg.APIListeners) != 0 {
		cfg.APIListeners = normalizeAddresses(cfg.APIListeners, defaultAPIPort)
	}

	// Handle environment variable expansion in the RPC certificate path.
	cfg.RPCCert = cleanAndExpandPath(cfg.RPCCert)

	var defaultRPCPort string
//	cfg.SimNet = true
	switch {
	case cfg.TestNet:
		defaultRPCPort = defaultRPCPortTestNet
		//chainParams = &chaincfg.TestNet3Params
		ChainParams = &params.TestNetParams
	case cfg.SimNet:
		defaultRPCPort = defaultRPCPortSimNet
		ChainParams = &params.PrivNetParams
	default:
		defaultRPCPort = defaultRPCPortMainNet
		ChainParams = &params.MainNetParams
	}

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

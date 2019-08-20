/**
Qitmeer
james
*/

package common

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	qitmeer "github.com/HalalChain/qitmeer-lib/common/hash"
	"math"
	"math/big"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"
)

func SliceContains(s []uint64, e uint64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func SliceRemove(s []uint64, e uint64) []uint64 {
	for i, a := range s {
		if a == e {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}

func Int2varinthex(x int64) string  {
	if x < 0xfd {
		return fmt.Sprintf("%02x",x)
	} else if x < 0xffff {
		return "fd" + Int2lehex(x, 2)
	} else if x < 0xffffffff {
		return "fe" + Int2lehex(x, 4)
	} else {
		return "ff" + Int2lehex(x, 8)
	}
}
func Int2lehex(x int64,width int ) string {
	if x <= 0 {
		return fmt.Sprintf("%016x",x)
	}
	bs := make([]byte, width)
	switch width {
	case 2:
		binary.LittleEndian.PutUint16(bs, uint16(x))
	case 4:
		binary.LittleEndian.PutUint32(bs, uint32(x))
	case 8:
		binary.LittleEndian.PutUint64(bs, uint64(x))
	}
	return hex.EncodeToString(bs)
}

// Reverse reverses a byte array.
func Reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	for i := len(src); i > 0; i-- {
		dst[len(src)-i] = src[i-1]
	}
	return dst
}

func BlockBitsToTarget(bits string,width int) []byte {
	nbits ,err:=hex.DecodeString(bits[0:2])
	if err != nil{
		fmt.Println("error",err.Error())
	}
	shift := nbits[0] - 3
	value,_ := hex.DecodeString(bits[2:])
	target0 :=make([]byte,int(shift))
	tmp := string(value) + string(target0)
	target1 := []byte(tmp)
	if len(target1)<width {
		head:=make([]byte,width-len(target1))
		target := string(head)+string(target1)
		return []byte(target)
	}
	return target1
}

// FormatHashRate sets the units properly when displaying a hashrate.
func FormatHashRate(h float64) string {
	if h > 1000000000000 {
		return fmt.Sprintf("%.3fTH/s", h/1000000000000)
	} else if h > 1000000000 {
		return fmt.Sprintf("%.3fGH/s", h/1000000000)
	} else if h > 1000000 {
		return fmt.Sprintf("%.0fMH/s", h/1000000)
	} else if h > 1000 {
		return fmt.Sprintf("%.1fkH/s", h/1000)
	} else if h == 0 {
		return "0H/s"
	}

	return fmt.Sprintf("%.1f TH/s", h)
}

func ReverseByWidth(s []byte,width int ) []byte {
	newS := make([]byte,len(s))
	for i := 0;i< (len(s) / width) ; i += 1 {
		j := i * width
		copy(newS[len(s)-j-width:len(s)-j],s[j:j+width])
	}
	return newS
}

func DiffToTarget(diff float64, powLimit *big.Int) (*big.Int, error) {
	if diff <= 0 {
		return nil, fmt.Errorf("invalid pool difficulty %v (0 or less than "+
			"zero passed)", diff)
	}

	// Round down in the case of a non-integer diff since we only support
	// ints (unless diff < 1 since we don't allow 0)..
	if diff < 1 {
		diff = 1
	} else {
		diff = math.Floor(diff)
	}
	divisor := new(big.Int).SetInt64(int64(diff))
	max := powLimit
	target := new(big.Int)
	target.Div(max, divisor)

	return target, nil
}

// Uint32EndiannessSwap swaps the endianness of a uint32.
func Uint32EndiannessSwap(v uint32) uint32 {
	return (v&0x000000FF)<<24 | (v&0x0000FF00)<<8 |
		(v&0x00FF0000)>>8 | (v&0xFF000000)>>24
}

// RolloverExtraNonce rolls over the extraNonce if it goes over 0x00FFFFFF many
// hashes, since the first byte is reserved for the ID.
func RolloverExtraNonce(v *uint32) {
	if *v&0x00FFFFFF == 0x00FFFFFF {
		*v = *v & 0xFF000000
	} else {
		*v++
	}
}


func ConvertHashToString( hash qitmeer.Hash )string{
	newB := make([]byte,32)
	copy(newB[:],hash[:])
	return hex.EncodeToString(newB)
}

// appDataDir returns an operating system specific directory to be used for
// storing application data for an application.  See AppDataDir for more
// details.  This unexported version takes an operating system argument
// primarily to enable the testing package to properly test the function by
// forcing an operating system that is not the currently one.
func appDataDir(goos, appName string, roaming bool) string {
	if appName == "" || appName == "." {
		return "."
	}

	// The caller really shouldn't prepend the appName with a period, but
	// if they do, handle it gracefully by stripping it.
	appName = strings.TrimPrefix(appName, ".")
	appNameUpper := string(unicode.ToUpper(rune(appName[0]))) + appName[1:]
	appNameLower := string(unicode.ToLower(rune(appName[0]))) + appName[1:]

	// Get the OS specific home directory via the Go standard lib.
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	switch goos {
	// Attempt to use the LOCALAPPDATA or APPDATA environment variable on
	// Windows.
	case "windows":
		// Windows XP and before didn't have a LOCALAPPDATA, so fallback
		// to regular APPDATA when LOCALAPPDATA is not set.
		appData := os.Getenv("LOCALAPPDATA")
		if roaming || appData == "" {
			appData = os.Getenv("APPDATA")
		}

		if appData != "" {
			return filepath.Join(appData, appNameUpper)
		}

	case "darwin":
		if homeDir != "" {
			return filepath.Join(homeDir, "Library",
				"Application Support", appNameUpper)
		}

	case "plan9":
		if homeDir != "" {
			return filepath.Join(homeDir, appNameLower)
		}

	default:
		if homeDir != "" {
			return filepath.Join(homeDir, "."+appNameLower)
		}
	}

	// Fall back to the current directory if all else fails.
	return "."
}

// AppDataDir returns an operating system specific directory to be used for
// storing application data for an application.
//
// The appName parameter is the name of the application the data directory is
// being requested for.  This function will prepend a period to the appName for
// POSIX style operating systems since that is standard practice.  An empty
// appName or one with a single dot is treated as requesting the current
// directory so only "." will be returned.  Further, the first character
// of appName will be made lowercase for POSIX style operating systems and
// uppercase for Mac and Windows since that is standard practice.
//
// The roaming parameter only applies to Windows where it specifies the roaming
// application data profile (%APPDATA%) should be used instead of the local one
// (%LOCALAPPDATA%) that is used by default.
//
// Example results:
//  dir := AppDataDir("myapp", false)
//   POSIX (Linux/BSD): ~/.myapp
//   Mac OS: $HOME/Library/Application Support/Myapp
//   Windows: %LOCALAPPDATA%\Myapp
//   Plan 9: $home/myapp
func AppDataDir(appName string, roaming bool) string {
	return appDataDir(runtime.GOOS, appName, roaming)
}

func Target2BlockBits(target string) []byte {
	// 8
	d,_ := hex.DecodeString(target[0:16])
	return Reverse(d)
}

func HexMustDecode(hexStr string) []byte {
	b, err := hex.DecodeString(hexStr)
	if err != nil {

		panic(err)
	}
	return b
}

func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))  //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

// fileName:文件名字(带全路径)
// content: 写入的内容
func AppendToFile(fileName string, content string) error {
	file,er:=os.Open(fileName)
	defer func(){file.Close()}()
	if er!=nil && os.IsNotExist(er){
		os.Create(fileName)
	}
	// 以只写的模式，打开文件
	f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fileName+" file create failed. err: " + err.Error())
	} else {
		// 查找文件末尾的偏移量
		n, _ := f.Seek(0, os.SEEK_END)
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt([]byte(content), n)
	}
	defer f.Close()
	return err
}

func GenerateRand(length int) uint32 {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	//buf,_ := seed.GenerateSeed(32)
	//log.Println(buf)
	//os.Exit(1)
	//buf := make([]byte, length)
	//rand.Read(buf)
	s2 := rand.NewSource(time.Now().UnixNano())

	r1 := rand.New(s2)
	r := uint32(r1.Intn(2<<32))
	return r
}

func RandUint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return uint64(binary.LittleEndian.Uint64(b[:])), nil
}

func RandGenerator(n int) chan uint32 {
	rand.Seed(time.Now().UnixNano())
	out := make(chan uint32)
	go func(x int) {
		for {
			out <- uint32(rand.Intn(x))
		}
	}(n)
	return out
}

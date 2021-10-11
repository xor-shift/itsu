package util

import (
	"github.com/intel-go/cpuid"
	"os"
	"os/user"
	"runtime"
)

type SystemInformation struct {
	GONumCPU int
	GOOS     string
	GOARCH   string

	ProcVendor           string
	ProcBranding         string
	ProcMaxID            uint32
	ProcFeatures         uint64
	ProcExtendedFeatures uint64
	ProcExtraFeatures    uint64
	//Caches               []cpuid.CacheDescriptor

	Hostname   string
	Username   string
	CacheDir   string
	ConfigDir  string
	HomeDir    string
	WorkingDir string
	ExecPath   string

	UID    int
	UIDStr string
	EUID   int
	GID    int
	GidStr string
	EGID   int

	Env []string
}

func GetSystemInformation() SystemInformation {
	features := uint64(0)
	extendedFeatures := uint64(0)
	extraFeatures := uint64(0)

	for i := uint64(0); i < 64; i++ {
		feature := uint64(1) << i

		if cpuid.HasFeature(feature) {
			features |= feature
		}

		if cpuid.HasExtendedFeature(feature) {
			extendedFeatures |= feature
		}

		if cpuid.HasExtraFeature(feature) {
			extraFeatures |= feature
		}
	}

	username := "nil"
	hostname := "nil"
	cacheDir := "nil"
	configDir := "nil"
	homeDir := "nil"
	workingDir := "nil"
	execPath := "nil"

	hostname, _ = os.Hostname()
	cacheDir, _ = os.UserCacheDir()
	configDir, _ = os.UserConfigDir()
	homeDir, _ = os.UserHomeDir()
	workingDir, _ = os.Getwd()
	execPath, _ = os.Executable()

	uidStr := "nil"
	gidStr := "nil"

	if u, err := user.Current(); err == nil {
		username = u.Username
		uidStr = u.Uid
		gidStr = u.Gid
	}

	return SystemInformation{
		GONumCPU: runtime.NumCPU(),
		GOOS:     runtime.GOOS,
		GOARCH:   runtime.GOARCH,

		ProcVendor:           cpuid.VendorIdentificatorString,
		ProcMaxID:            cpuid.MaxLogicalCPUId,
		ProcBranding:         cpuid.ProcessorBrandString,
		ProcFeatures:         features,
		ProcExtendedFeatures: extendedFeatures,
		ProcExtraFeatures:    extraFeatures,
		//Caches:               cpuid.CacheDescriptors,

		Hostname:   hostname,
		Username:   username,
		CacheDir:   cacheDir,
		ConfigDir:  configDir,
		HomeDir:    homeDir,
		WorkingDir: workingDir,
		ExecPath:   execPath,

		UID:    os.Getuid(),
		UIDStr: uidStr,
		EUID:   os.Geteuid(),
		GID:    os.Getgid(),
		GidStr: gidStr,
		EGID:   os.Getegid(),

		Env: os.Environ(),
	}
}

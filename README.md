# LameGotato
Spawn processes as NT AUTHORITY\SYSTEM


Written in Go. 

LameGotato spawns processes as SYSTEM using the LookupPrivilegeValueW, AdjustTokenPrivileges, CreateProcessWithTokenW, OpenProcessToken, DuplicateTokenEx Windows API calls. This must be invoked with administrator privileges, and you must supply the Process ID of a process currently running as SYSTEM (LSASS and Winlogon work well). The default program spawned is C:\Windows\System32\cmd.exe, but you can specify a different program with "-Spawn".

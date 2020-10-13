package reslove

const (
	ipv4AddressReg = `(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`
	//HostBruteLoginRecordReg = `^.*pam_unix\(sshd:auth\): authentication failure.*rhost=(\S*)\s+.*$`
	//HostBruteLoginRecordReg = `^.*pam_unix\(sshd:auth\): authentication failure.*rhost=(\S*)`
	HostBruteLoginRecordReg = `^.*pam_unix\(sshd:.+\):.*rhost=(\S*)`
	//HostBruteLoginRecordWithUserReg = `^.*pam_unix\(sshd:auth\): authentication failure.*rhost=(.*)\s+user=(\S*)\s*.*$`
)

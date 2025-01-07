//go:build !windows

package placeholders

func GetPlaceholderConfigMumble() string {
	return "Ex.  /home/user/mumble/mumble"
}

func GetPlaceholderConfigTor() string {
	return "Ex. /user/sbin/tor"
}

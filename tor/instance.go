package tor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"autonomia.digital/tonio/app/config"
)

func createDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

func createFile(configFile string) error {
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}

	_, _ = f.WriteString("#Tor Proxy Port \n")
	_, _ = f.WriteString("SocksPort 9950 \n")

	_, _ = f.WriteString("#Tor Control Port \n")
	_, _ = f.WriteString("ControlPort 9951 \n")

	_, _ = f.WriteString("#Data directory where authentication cookie would be saved \n")
	_, _ = f.WriteString("DataDirectory ~/.config/tonio/tor/data \n")

	_, _ = f.WriteString("CookieAuthentication 1\n")

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func checkConfigFile() {
	//homeDir, err := os.UserHomeDir()
	homeDir := config.XdgConfigHome()

	configDir := fmt.Sprintf("%s/", homeDir)

	dataApp := fmt.Sprintf("%s/%s", configDir, "tonio/tor/data")
	_, err := os.Stat(dataApp)
	if err != nil {
		errDir := createDirectory(dataApp)
		if errDir != nil {
			log.Fatalf("Tor data directory can not be created: %v", errDir)
		}

		torConfigFile := fmt.Sprintf("%s/%s/%s/%s", configDir, "tonio", "tor", "torrc")

		errFile := createFile(torConfigFile)
		if errFile != nil {
			log.Fatalf("Tor configuration file can not be created: %v", errDir)
		}
	}
}

func LaunchTorInstance() {
	checkConfigFile()

	ctx := context.Background()

	ch := fmt.Sprintf("%s/tonio/tor/torrc", config.XdgConfigHome())
	cmd := exec.CommandContext(ctx, "tor", "-f", ch)
	if err := cmd.Start(); err != nil {
		log.Fatalf("Tor instance can not be launched: %s", err)
	}
}

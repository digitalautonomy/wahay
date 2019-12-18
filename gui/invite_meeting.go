package gui

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) getInviteCodeEntities() (gtki.Entry, gtki.ApplicationWindow, *uiBuilder) {
	builder := u.g.uiBuilderFor("InviteCodeWindow")
	url := builder.get("entMeetingID").(gtki.Entry)
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	return url, win, builder
}

func (u *gtkUI) openDialog() {
	url, win, builder := u.getInviteCodeEntities()

	builder.ConnectSignals(map[string]interface{}{
		"on_join": func() {
			idEntered, err := url.GetText()
			if err != nil {
				u.openErrorDialog()
				return
			}
			err = openMumble(idEntered)
			if err != nil {
				u.openErrorDialog()
			}
		},
		"on_cancel": func() {
			win.Hide()
		},
	})
	win.ShowAll()
}

func openMumble(inviteID string) error {
	fmt.Println("Opening Mumble....")
	log.Println(inviteID)
	//qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion

	if !isMeetingIDValid(inviteID) {
		return fmt.Errorf("invalid Onion Address %s", inviteID)
	}

	go launchMumbleClient(inviteID)

	return nil
}

const onionServiceLength = 60

//This function needs to be improved in order to make a real validation of the Meeting ID or Onion Address.
//At the moment, this function helps to test the error code window render.
func isMeetingIDValid(meetingID string) bool {
	return len(meetingID) > onionServiceLength && strings.HasSuffix(meetingID, ".onion")
}

func launchMumbleClient(inviteID string) {
	mumbleURL := fmt.Sprintf("mumble://%s", inviteID)
	cmd := exec.Command("torify", "mumble", mumbleURL)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

package gui

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/coyim/gotk3adapter/gtki"
)

type inviteMeetingData struct {
	builder *uiBuilder
	url     gtki.Entry
}

type inviteMeetingDetails struct {
	url string
}

func (u *gtkUI) getInviteMeetingBuilderAndData() *inviteMeetingData {
	data := &inviteMeetingData{}
	data.builder = u.g.uiBuilderFor("InviteCodeWindow")

	data.builder.getItems(
		"entMeetingID", &data.url,
	)

	return data
}

func getInviteMeetingDetails(data *inviteMeetingData) *inviteMeetingDetails {

	urlTxt, _ := data.url.GetText()

	details := &inviteMeetingDetails{
		urlTxt,
	}

	return details
}

func (u *gtkUI) openDialog() {

	data := u.getInviteMeetingBuilderAndData()

	builder := data.builder
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	builder.ConnectSignals(map[string]interface{}{
		"on_join": func() {
			inviteDetails := getInviteMeetingDetails(data)
			err := openMumble(inviteDetails.url)
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

func openMumble(onionURL string) error {
	fmt.Println("Opening Mumble....")
	log.Println(onionURL)
	//qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion

	if !validateMeetingID(onionURL) {
		msg := fmt.Sprintf("Invalid Onion Address %s", onionURL)
		return errors.New(msg)
	}
	cmd := exec.Command("torify", "mumble", fmt.Sprintf("mumble://%s", onionURL))

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	return nil
}

//This function needs to be improved in order to make a real validation of the Meeting ID or Onion Address.
//At the moment, this function helps to test the error code window render.
func validateMeetingID(meetingID string) bool {
	if meetingID == "" || !strings.Contains(meetingID, ".onion") {
		return false
	}
	return true
}

package gui

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) openDialog() {

	builder := u.g.uiBuilderFor("InviteCodeWindow")
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	builder.ConnectSignals(map[string]interface{}{
		"on_join": func() {
			openMumble()
		},
		"on_cancel": func() {
			win.Hide()
		},
	})
	win.ShowAll()
}

func openMumble() {
	fmt.Println("Opening Mumble....!")
	// cmd := exec.Command("torify mumble mumble://qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion")
	//cmd := exec.Command("mumble mumble://127.0.0.1")
	cmd := exec.Command("torify","mumble","mumble://127.0.0.1")
	// cmd.Env = append(os.Environ(),
	// 	"mumble ",                              // ignored
	// 	"mumble://code@s5a3jjqeqsch46vc.onion", // this value is used
	// )
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

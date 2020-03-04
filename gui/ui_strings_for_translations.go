package gui

func init() {
	noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt1()
	noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt2()
	noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt3()
	noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt4()
}

func noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt1() {
	// Please add all strings from the definitions directory that
	// should be translated to this function. If not, the generation of
	// translation files will not work correctly.
	// As always, test that what you did works - that includes checking
	// that strings are translated properly in your code

	// The reason why we need to do this, is that "gotext" will only allow
	// us to use translation strings for strings that it can detect.
	// Since it doesn't see the strings in the definitions directory, that
	// means it would be impossible to provide translations for them if
	// we didn't use this workaround. Don't judge us!

	// TODO: fix a better solution! Maybe patch gotext to have a flag to change
	// this behavior or something else.

	_ = i18n.Sprintf("Accept")
	_ = i18n.Sprintf("Allow the host to automatically join a newly created meeting")
	_ = i18n.Sprintf("Are you sure you want to do this action?")
	_ = i18n.Sprintf("Are you sure you want to end this meeting?")
	_ = i18n.Sprintf("Are you sure you want to leave this meeting?")
	_ = i18n.Sprintf("Automatically join this meeting")
	_ = i18n.Sprintf("Automatically join a meeting")
	_ = i18n.Sprintf("Automatically join this meeting when starting it")
	_ = i18n.Sprintf("Be very careful. This information is sensitive and could " +
		"potentially contain very private information. Only turn on these settings if you absolutely need it for debugging.")
	_ = i18n.Sprintf("Browse")
	_ = i18n.Sprintf("By clicking Yes, this meeting will end.")
	_ = i18n.Sprintf("By clicking Yes, you will leave this meeting.")
	_ = i18n.Sprintf("Cancel")
	_ = i18n.Sprintf("Client binary location")
	_ = i18n.Sprintf("Configuration settings will be lost in the next session")
	_ = i18n.Sprintf("Configure master password")
}

func noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt2() {
	_ = i18n.Sprintf("Confirmation")
	_ = i18n.Sprintf("Connecting, please wait...")
	_ = i18n.Sprintf("Continue")
	_ = i18n.Sprintf("Copy Invitation")
	_ = i18n.Sprintf("Copy Meeting ID")
	_ = i18n.Sprintf("Copy URL")
	_ = i18n.Sprintf("Check this option to automatically join every meeting you host")
	_ = i18n.Sprintf("Choose your email service to send invitation")
	_ = i18n.Sprintf("Debugging")
	_ = i18n.Sprintf("Default Email")
	_ = i18n.Sprintf("Encrypt the configuration file")
	_ = i18n.Sprintf("Error")
	_ = i18n.Sprintf("Finish")
	_ = i18n.Sprintf("End this meeting")
	_ = i18n.Sprintf("End this meeting for all")
	_ = i18n.Sprintf("General")
	_ = i18n.Sprintf("Gmail")
	_ = i18n.Sprintf("Host a new meeting")
	_ = i18n.Sprintf("Hosting")
	_ = i18n.Sprintf("Host meeting")
	_ = i18n.Sprintf("If you backup the configuration file, we will reset " +
		"the settings and continue normally. If the configuration file is encrypted, then we will ask you " +
		"for a password to encrypt the new settings file.")
	_ = i18n.Sprintf("If you set this option to a file name, low level information will be logged there.")
	_ = i18n.Sprintf("Invalid configuration file")
	_ = i18n.Sprintf("Invalid password. Please, try again.")
	_ = i18n.Sprintf("Invite others")
	_ = i18n.Sprintf("Join")
	_ = i18n.Sprintf("Join a meeting")
	_ = i18n.Sprintf("Join meeting")
	_ = i18n.Sprintf("Join the meeting")
	_ = i18n.Sprintf("Join this meeting")
	_ = i18n.Sprintf("Keep configuration file when Wahay closes")
	_ = i18n.Sprintf("Leave")
	_ = i18n.Sprintf("Leave this meeting")
	_ = i18n.Sprintf("Log debug info")
	_ = i18n.Sprintf("Log debug output to the selected log file. If no file is " +
		"selected then the log output will be written to the default log file.")
	_ = i18n.Sprintf("Master password")
	_ = i18n.Sprintf("Meeting ID")
}

func noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt3() {
	_ = i18n.Sprintf("Meeting ID:")
	_ = i18n.Sprintf("Meeting password")
	_ = i18n.Sprintf("Mumble")
	_ = i18n.Sprintf("No, cancel")
	_ = i18n.Sprintf("Now you are hosting a meeting.")
	_ = i18n.Sprintf("Outlook")
	_ = i18n.Sprintf("Password")
	_ = i18n.Sprintf("Please enter the master password for the configuration file.")
	_ = i18n.Sprintf("Port")
	_ = i18n.Sprintf("Port out of range")
	_ = i18n.Sprintf("Raw log file")
	_ = i18n.Sprintf("Repeat the password")
	_ = i18n.Sprintf("Save changes")
	_ = i18n.Sprintf("Security")
	_ = i18n.Sprintf("Settings")
	_ = i18n.Sprintf("Show")
	_ = i18n.Sprintf("Specify a password for the meeting")
	_ = i18n.Sprintf("Start meeting")
	_ = i18n.Sprintf("The error message")
	_ = i18n.Sprintf("The meeting ID has been copied to the clipboard")
	_ = i18n.Sprintf("A valid port is between 1 and 65535")
	_ = i18n.Sprintf("This action cannot be undone")
	_ = i18n.Sprintf("Toggle password visibility")
	_ = i18n.Sprintf("Type the Meeting ID (normally a .onion address)")
	_ = i18n.Sprintf("Type the password")
	_ = i18n.Sprintf("Type the password to join the meeting")
	_ = i18n.Sprintf("Type your preferred screen name")
	_ = i18n.Sprintf("Type your screen name")
	_ = i18n.Sprintf("Username")
	_ = i18n.Sprintf("Wahay is ready to use")
	_ = i18n.Sprintf("We have detected that the configuration file is invalid or corrupted. " +
		"Do you want to make a copy (backup) of it and continue?")
	_ = i18n.Sprintf("Welcome")
	_ = i18n.Sprintf("When this option is checked, the configuration settings will be stored in the device.")
	_ = i18n.Sprintf("Yahoo Mail")
	_ = i18n.Sprintf("Yes, back it up &amp; continue")
	_ = i18n.Sprintf("Yes, confirm")
	_ = i18n.Sprintf("You will not be asked for this password again until you restart Wahay.")
}

func noPointInEverCallingThisButYouCanIfYouReallyFeelLikeIt4() {
	_ = i18n.Sprintf("Executable Mumble location")
	_ = i18n.Sprintf("Ex. /home/user/mumble/mumble")
	_ = i18n.Sprintf("If you want to use your own Mumble instance, please enter the location " +
		"where Mumble is available in the system.")
	_ = i18n.Sprintf("Mumble service port")
	_ = i18n.Sprintf("Ex. 9800")
	_ = i18n.Sprintf("If you want to set up a custom port to run the Mumble service, " +
		"please a port number between 1 and 65535")
	_ = i18n.Sprintf("What is Wahay?")
	_ = i18n.Sprintf("Communication is a basic need of the human being, in its beginnings it is carried out verbally " +
		"from person to person through the use of technology, various tools have been developed for this purpose stories " +
		"such as: Skype, Zoom, Google Hangouts, etc. However, there are several aspects that have not been considered " +
		"in the development of these solutions: centralized servers, proprietary technology, security, are some aspects " +
		"that have not been contemplated or have been partially implemented.")
	_ = i18n.Sprintf("Wahay (https://wahay.org) has been developed as a tool for conducting voice " +
		"conferences in an easy, extremely secure and decentralized manner (without the need for " +
		"any centralized server or service). Internally it uses Tor (https://www.torproject.org/) " +
		"as a tool to establish secure communications and Mumble (https://www.mumble.com/) as a " +
		"client to establish voice over IP.")
	_ = i18n.Sprintf("What is Tor?")
	_ = i18n.Sprintf("Tor is a free and open source tool that allows you to establish anonymous " +
		"and distributed communications. Tor directs its internet traffic through a series of " +
		"routers called 'onion routers' allowing anonymous communication between its nodes, " +
		"this network works from a set of organizations and individuals that donate their " +
		"bandwidth and processing power.")
	_ = i18n.Sprintf("What is Mumble?")
	_ = i18n.Sprintf("Mumble is a free and open source application that allows voice over " +
		"IP conferences between users with high sound quality and low latency.")
	_ = i18n.Sprintf("Functionalities")
	_ = i18n.Sprintf("Wahay allows you to host a meeting or join an existing meeting, " +
		"for this you establish an ID that will serve as the identifier of the meeting to use.")
	_ = i18n.Sprintf("Host a meeting")
	_ = i18n.Sprintf("This option allows starting the server that will support the connection " +
		"of users to a meeting which is defined by its ID (meeting identifier), this ID must be " +
		"used by the rest of users who wish to access it. Additionally it is possible to define " +
		"the user name (not mandatory) that will be used to identify the user in the meeting, it " +
		"is also possible to configure the password to access the meeting, which will be required " +
		"by users who wish to access Wahay.")
	_ = i18n.Sprintf("The option to automatically join this meeting allows you to start the server and " +
		"enter it, if you do not select it, you can access it later by selecting the join button. It is " +
		"also possible to copy the meeting ID and send the invitation by the most used email clients.")
	_ = i18n.Sprintf("This option allows the user to access a meeting if already exist, for this you must " +
		"enter the meeting id (required), username (not required) and password (if was set).")
	_ = i18n.Sprintf("Help")
}

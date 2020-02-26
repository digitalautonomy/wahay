package gui

import . "gopkg.in/check.v1"

type WahayInviteMeetingSuite struct{}

var _ = Suite(&WahayInviteMeetingSuite{})

func (s *WahayInviteMeetingSuite) Test_InviteMeeting_ValidateMeetingID(c *C) {
	v1 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion")
	v2 := isAValidMeetingID("aaabbbcccddd.onion")
	v3 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid:8080")
	v4 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion:8080")
	v5 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid")
	v6 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion:4100")

	c.Assert(v1, Equals, true)
	c.Assert(v2, Equals, false)
	c.Assert(v3, Equals, false)
	c.Assert(v4, Equals, true)
	c.Assert(v5, Equals, false)
	c.Assert(v6, Equals, true)
}

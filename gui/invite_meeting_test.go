package gui

import . "gopkg.in/check.v1"

type WahayInviteMeetingSuite struct{}

var _ = Suite(&WahayInviteMeetingSuite{})

func (s *WahayInviteMeetingSuite) Test_InviteMeeting_extractMeetingIDandPort_SucceedIfValidUrl(c *C) {
	h1, p1, e1 := extractMeetingIDandPort("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion")
	h2, p2, e2 := extractMeetingIDandPort("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion:8080")

	c.Assert(h1, Equals, "qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion")
	c.Assert(p1, Equals, 0)
	c.Assert(e1, Equals, nil)

	c.Assert(h2, Equals, "qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion")
	c.Assert(p2, Equals, 8080)
	c.Assert(e2, Equals, nil)
}

func (s *WahayGUIDefinitionsSuite) Test_InviteMeeting_extractMeetingIDandPort_FailsIfNoValidUrl(c *C) {
	_, _, e1 := extractMeetingIDandPort("aaabbbcccddd.onion")
	_, _, e2 := extractMeetingIDandPort("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid:8080")
	_, _, e3 := extractMeetingIDandPort("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid")
	_, _, e4 := extractMeetingIDandPort("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion:aaaa")

	c.Assert(e1, ErrorMatches, "invalid meeting address")
	c.Assert(e2, ErrorMatches, "invalid meeting address")
	c.Assert(e3, ErrorMatches, "invalid meeting address")
	c.Assert(e4, ErrorMatches, "invalid meeting address")
}

func (s *WahayInviteMeetingSuite) Test_InviteMeeting_isAValidMeetingID_SucceedIfValidUrl(c *C) {
	v1 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion")
	v2 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion:8080")
	v3 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion:4100")

	c.Assert(v1, Equals, true)
	c.Assert(v2, Equals, true)
	c.Assert(v3, Equals, true)
}

func (s *WahayInviteMeetingSuite) Test_InviteMeeting_isAValidMeetingID_FailsIfNoValidUrl(c *C) {
	v1 := isAValidMeetingID("aaabbbcccddd.onion")
	v2 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid:8080")
	v3 := isAValidMeetingID("qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid")

	c.Assert(v1, Equals, false)
	c.Assert(v2, Equals, false)
	c.Assert(v3, Equals, false)
}

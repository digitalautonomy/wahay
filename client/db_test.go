package client

import (
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_exists_returnsTrueIfStringExistsInContent(c *C) {
	db := &dbData{
		content: []byte("this is true"),
	}

	exampleString := "true"

	res := db.exists(exampleString)

	c.Assert(res, Equals, true)
}

func (s *clientSuite) Test_exists_returnsFalseIfStringDoesntExistInContent(c *C) {
	db := &dbData{
		content: []byte("this is false"),
	}

	exampleString := "true"

	res := db.exists(exampleString)

	c.Assert(res, Equals, false)
}

func (s *clientSuite) Test_replaceString_findsAndReplacesContentInDB(c *C) {
	db := &dbData{
		content: []byte("this is true"),
	}

	find := "true"
	replace := "false"
	expectedContent := []byte("this is false")

	db.replaceString(find, replace)

	c.Assert(db.content, DeepEquals, expectedContent)
}

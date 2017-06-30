package main

import (
	"log"
	"testing"

	check "gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type createDbSuite struct{}

var _ = check.Suite(&createDbSuite{})

func (s *createDbSuite) TestDoTable(c *check.C) {
	m1 := func() error {
		log.Println("Successful execution 1")
		return nil
	}

	m2 := func() error {
		log.Println("Successful execution 2")
		return nil
	}

	exec([]operation{
		{m1, create, "the table 1"},
		{m2, update, "the table 2"},
	})
}

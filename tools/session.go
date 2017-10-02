package tools

import (
	"gopkg.in/mgo.v2"
	"fmt"
	"errors"
)

var (
	session *mgo.Session
)

func InitSession() *mgo.Session {
	var err error
	dialInfo := mgo.DialInfo{Addrs: []string{"localhost"}, Database: "jc_test"}
	session, err = mgo.DialWithInfo(&dialInfo)

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB server. %s", err))
	}
	return session
}

func GetSession() (*mgo.Session, error) {
	var err error
	if session == nil {
		err = errors.New("Session is not initialized")
	}
	return session, err
}
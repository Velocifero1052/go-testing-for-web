package main

import (
	"os"
	"testing"
	"testingCourserWeb/pkg/repository/dbrepo"
)

var app application

var expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiYXVkIjoiZXhhbXBsZS5jb20iLCJleHAiOjE2NzM4MDE5OTUsImlzcyI6ImV4YW1w\nbGUuY29tIiwibmFtZSI6IkpvaG4gRG9lIiwic3ViIjoiMSJ9.xyA4lFlOnfOhvMTRSGLCwafC3f1-hGkLFxGR2FWMzSw"

func TestMain(m *testing.M) {
	app.DB = &dbrepo.TestDBRepo{}
	app.Domain = "example.com"
	app.JWTSecret = "asdf123sadafasdf123123sadfasdf12312asdfasdf123123asdfasdf"
	os.Exit(m.Run())
}

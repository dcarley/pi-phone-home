package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPiPhoneHome(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PiPhoneHome Suite")
}

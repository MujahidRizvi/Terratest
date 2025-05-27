package test

import "encoding/xml"

type TestCase struct {
	Classname string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
	Status    string   `xml:"status"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuite struct {
	XMLName   xml.Name   `xml:"testsuite"`
	Tests     int        `xml:"tests,attr"`
	Failures  int        `xml:"failures,attr"`
	Errors    int        `xml:"errors,attr"`
	Time      float64    `xml:"time,attr"`
	TestCases []TestCase `xml:"testcase"`
}

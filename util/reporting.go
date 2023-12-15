package util

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wasmmock/wasm_mock_server/util/postman"
)

type Report struct {
	Tests              []UnitTest `json:"tests"`
	StartTime          time.Time
	StartTimeFormatted string `json:"start_time_formatted"`
	EndTimeFormatted   string `json:"end_time_formatted"`
	Date               string `json:"date"`
	Duration           string `json:"duration"`
	TestEndPoint       string `json:"test_end_point"`
	Passes             int32  `json:"passes"`
	Skips              int32  `json:"skips"`
	TotalTests         int32  `json:"total_tests"`
	BackgroundMock     []Mock `json:"background_mock"`
	Hits               []Hit  `json:"hits"`
	Postman            string `json:"postman"`
	PostmanItem        []postman.Item
}
type UnitTest struct {
	Index              int64         `json:"index"`
	Expectations       []Expectation `json:"expectations"`
	MockData           []Mock        `json:"MockData"`
	Request            string        `json:"request"`
	Response           string        `json:"response"`
	Steps              []Step        `json:"Step"`
	StartTime          time.Time
	StartTimeFormatted string `json:"start_time_formatted"`
	EndTimeFormatted   string `json:"end_time_formatted"`
	Duration           string `json:"duration"`
}
type Mock struct {
	Command  string `json:"command"`
	Request  string `json:"request"`
	Response string `json:"response"`
	Pass     bool   `json:"pass"`
	Index    int64  `json:"index"`
	Duration string `json:"duration"`
	EndTime  string `json:"end_time"`
	Source   string `json:"source"`
	TraceId  string `json:"trace_id"`
	Time     time.Time
}
type Expectation struct {
	Pass        bool   `json:"pass"`
	Description string `json:"description"`
}
type Step struct {
	Pass        bool   `json:"pass"`
	Description string `json:"description"`
}
type Hit struct {
	Command   string `json:"command"`
	StartTime string `json:"start_time"`
	Time      time.Time
}

func UnitTestGen(index int64) UnitTest {
	startTime := time.Now()
	return UnitTest{Index: index, Request: "", Response: "", Expectations: []Expectation{},
		StartTime: startTime, StartTimeFormatted: startTime.Format("15:04:05")}
}
func ReportGen() Report {
	return Report{
		Date:               time.Now().Format("01-02-2006"),
		StartTime:          time.Now(),
		StartTimeFormatted: time.Now().Format("15:04:05"),
		Tests:              []UnitTest{},
		TestEndPoint:       "",
		Passes:             0,
		Skips:              0,
		TotalTests:         0,
		BackgroundMock:     []Mock{},
		Hits:               []Hit{},
		Postman:            "",
		PostmanItem:        []postman.Item{},
	}
}
func (m *Report) appendUnitTest(index int64) {
	unitTest := UnitTestGen(index)
	m.Tests = append(m.Tests, unitTest)
}
func (m *Report) appendRequest(req string) int64 {
	n := int64(len(m.Tests))
	m.Tests[n-1].Request = req
	return n - 1 //index of appended
}
func (m *Report) setRequest(req string, index int) bool {
	if len(m.Tests) < index+1 {
		return false
	}
	m.Tests[index].Request = req
	return true
}
func (m *Report) appendResponse(res string) {
	m.Tests[len(m.Tests)-1].Response = res
}
func (m *Report) setResponse(res string, index int) bool {
	if len(m.Tests) < index+1 {
		return false
	}
	m.Tests[index].Response = res
	return true
}
func (m *Report) appendExpectation(ex Expectation) bool {
	if len(m.Tests) == 0 {
		return false
	}
	m.Tests[len(m.Tests)-1].Expectations = append(m.Tests[len(m.Tests)-1].Expectations, ex)
	return true
}
func (m *Report) setExpectation(ex Expectation, index int) bool {
	if len(m.Tests) < index+1 {
		return false
	}
	m.Tests[index].Expectations = append(m.Tests[index].Expectations, ex)
	return true
}
func (m *Report) appendEnd(backgroundMock []Mock, hits []Hit) {
	if len(m.Tests) > 0 {
		endTime := time.Now()
		m.Tests[len(m.Tests)-1].EndTimeFormatted = endTime.Format("15:04:05")
		m.Tests[len(m.Tests)-1].Duration = DurationToString(endTime.Sub(m.Tests[len(m.Tests)-1].StartTime))
	}
	m.BackgroundMock = backgroundMock
	m.Hits = hits
}
func (m *Report) appendMock(mock Mock) bool {
	if len(m.Tests) == 0 {
		return false
	}
	m.Tests[len(m.Tests)-1].MockData = append(m.Tests[len(m.Tests)-1].MockData, mock)
	return true
}
func (m *Report) setMock(mock Mock, index int) bool {
	if len(m.Tests) < index+1 {
		return false
	}
	m.Tests[index].MockData = append(m.Tests[index].MockData, mock)
	return true
}
func (m *Report) appendStep(s Step) bool {
	if len(m.Tests) == 0 {
		return false
	}
	m.Tests[len(m.Tests)-1].Steps = append(m.Tests[len(m.Tests)-1].Steps, s)
	return true
}
func (m *Report) appendToLastStep(s Step) bool {
	if len(m.Tests) == 0 {
		return false
	}
	if len(m.Tests[len(m.Tests)-1].Steps) > 0 {
		len_of_steps := len(m.Tests[len(m.Tests)-1].Steps)
		m.Tests[len(m.Tests)-1].Steps[len_of_steps-1].Description += "<br/>" + s.Description
		return true
	}
	return false

}
func (m *Report) appendPostmanItem(s postman.Item) bool {
	m.PostmanItem = append(m.PostmanItem, s)
	return true
}
func (m *Report) setStep(s Step, index int) bool {
	if len(m.Tests) < index+1 {
		return false
	}
	m.Tests[index].Steps = append(m.Tests[index].Steps, s)
	return true
}
func (m *Report) len() int {
	return len(m.Tests)
}

type SafeReports struct {
	mu      sync.Mutex
	Reports map[string]Report
}

func (c *SafeReports) AppendUnitTest(uid string, index int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	l.appendUnitTest(index)
	c.Reports[uid] = l
}
func (c *SafeReports) AppendRequest(uid string, req string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	len := l.appendRequest(req)
	c.Reports[uid] = l
	return len
}
func (c *SafeReports) SetRequest(uid string, req string, index int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	n := l.setRequest(req, int(index))
	c.Reports[uid] = l
	return n
}
func (c *SafeReports) AppendResponse(uid string, res string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	l.appendResponse(res)
	c.Reports[uid] = l
}
func (c *SafeReports) SetResponse(uid string, res string, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	if l.setResponse(res, index) {
		c.Reports[uid] = l
	} else {
		fmt.Println("setRespobse err", time.Now(), uid, index)
	}
}
func (c *SafeReports) AppendExpectation(uid string, ex Expectation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	if l.appendExpectation(ex) {
		c.Reports[uid] = l
	} else {
		fmt.Println("appendExpectation err", time.Now(), uid)
	}
}
func (c *SafeReports) SetExpectation(uid string, ex Expectation, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	if l.setExpectation(ex, index) {
		c.Reports[uid] = l
	} else {
		fmt.Println("setExpectation err", time.Now(), uid, index)
	}
}
func (c *SafeReports) AppendEnd(uid string, backgroundMock []Mock, hit []Hit) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	l.appendEnd(backgroundMock, hit)
	c.Reports[uid] = l
}
func (c *SafeReports) AppendStep(uid string, s Step) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	if l.appendStep(s) {
		c.Reports[uid] = l
	} else {
		fmt.Println("appendStep err", time.Now(), uid, s.Description)
	}
}
func (c *SafeReports) AppendToLastStep(uid string, s Step) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	if l.appendToLastStep(s) {
		c.Reports[uid] = l
	} else {
		fmt.Println("appendToLastStep err", time.Now(), uid, s.Description)
	}
}
func (c *SafeReports) SetStep(uid string, s Step, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if l, ok := c.Reports[uid]; ok {
		l.setStep(s, index)
		c.Reports[uid] = l
	} else {
		fmt.Println("setExpectation err", time.Now(), uid, index)
	}
}
func (c *SafeReports) AppendMock(uid string, mock Mock) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if l, ok := c.Reports[uid]; ok {
		l.appendMock(mock)
		c.Reports[uid] = l
	} else {
		fmt.Println("appendMock err", time.Now(), uid)
	}
}
func (c *SafeReports) SetMock(uid string, mock Mock, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if l, ok := c.Reports[uid]; ok {
		l.setMock(mock, index)
		c.Reports[uid] = l
	} else {
		fmt.Println("setExpectation err", time.Now(), uid, index)
	}
}
func (c *SafeReports) AppendPostmanItem(uid string, s postman.Item) {
	c.mu.Lock()
	defer c.mu.Unlock()
	l := c.Reports[uid]
	l.appendPostmanItem(s)
	c.Reports[uid] = l
}
func (c *SafeReports) Save(uid string) {
	c.mu.Lock()
	report := c.Reports[uid]
	endTime := time.Now()
	report.Duration = DurationToString(endTime.Sub(report.StartTime))
	report.EndTimeFormatted = endTime.Format("15:04:05")

	var timeOuts = []Mock{}
	for ui, unit := range report.Tests {
		for i := len(unit.MockData) - 1; i >= 0; i-- {
			if int(unit.MockData[i].Index) != ui {
				timeOuts = append(timeOuts, unit.MockData[i])
				unit.MockData = append(unit.MockData[:i], unit.MockData[i+1:]...)
			}
		}
		for _, e := range unit.Expectations {
			if e.Pass {
				report.Passes = report.Passes + 1
			}
			report.TotalTests = report.TotalTests + 1
		}
		for _, e := range unit.Steps {
			if !e.Pass {
				report.Skips = report.Skips + 1
			}
		}
	}
	for _, m := range timeOuts {
		mn := m
		mn.Pass = false
		report.Tests[int(m.Index)].MockData = append(report.Tests[int(m.Index)].MockData, mn)
	}
	if len(report.PostmanItem) > 0 {
		collection := postman.Collection{
			Info: postman.Info{
				Postman_id: "",
				Schema:     "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
			},
			Item: report.PostmanItem,
		}
		if collection_bytes, err := json.Marshal(collection); err == nil {
			//report.Postman = b64.StdEncoding.EncodeToString(collection_bytes)
			report.Postman = string(collection_bytes)
		}
		report.PostmanItem = []postman.Item{}
	}
	c.mu.Unlock()

	SaveReport(uid, report)
}
func (c *SafeReports) ClearUnitTest(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Reports, uid)
}
func (c *SafeReports) ReportGen(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	report := ReportGen()
	c.Reports[uid] = report
}
func (c *SafeReports) Len(uid string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if l, ok := c.Reports[uid]; ok {
		return l.len()
	}
	return 0
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func DurationToString(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder

	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	fmt.Fprintf(&b, "%dd%s", days, d)

	return b.String()
}

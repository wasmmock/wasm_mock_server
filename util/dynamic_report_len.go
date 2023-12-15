package util

import "sync"

type ReportLen map[string]int64
type SafeReportLen struct {
	mu        sync.Mutex
	reportlen ReportLen
}

func SafeReportLenNew() SafeReportLen {
	return SafeReportLen{
		reportlen: make(ReportLen),
	}
}
func (c *SafeReportLen) CreateReport(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reportlen[uid] = 0
}
func (c *SafeReportLen) Increment(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.reportlen[uid]
	if ok {
		c.reportlen[uid] = c.reportlen[uid] + 1
	}
}
func (c *SafeReportLen) Get(uid string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.reportlen[uid]
	if ok {
		return c.reportlen[uid]
	} else {
		c.reportlen[uid] = 0
	}
	return 0
}
func (c *SafeReportLen) DeleteReport(uid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.reportlen[uid]
	if ok {
		delete(c.reportlen, uid)
	}
}

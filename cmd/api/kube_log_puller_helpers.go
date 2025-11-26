package main

import (
	"fmt"
	"strings"
	"time"
)

func parseK8sTimestampLine(line string) (time.Time, string, error) {
	// แยก timestamp กับส่วนที่เหลือด้วย space ตัวแรก
	idx := strings.IndexByte(line, ' ')
	if idx == -1 {
		// รูปแบบไม่ตรงที่คาด → คืน error พร้อมทั้งบรรทัด
		return time.Time{}, line, fmt.Errorf("invalid log line: no space")
	}

	tsStr := line[:idx]
	msg := strings.TrimSpace(line[idx+1:])

	// parse timestamp ตาม format ของ k8s
	ts, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		// parse ไม่ได้ → คืน msg ให้ caller ไปใช้ทั้งบรรทัดแทน
		return time.Time{}, msg, err
	}

	return ts, msg, nil
}

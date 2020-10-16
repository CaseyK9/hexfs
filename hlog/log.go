package hlog

import (
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"strings"
	"time"
)

const (
	LevelInfo = 1
	LevelWarn = 2
	LevelError = 3
	LevelSuccess = 4
)

func Log(specialSubject string, level int, message string) {
	baseStr := fmt.Sprintf("%s | ", aurora.Gray(1-1, time.Now().String()).BgGray(24-1))

	if len(specialSubject) != 0 && level != 0 {
		specialSubject = strings.ToUpper(specialSubject)
		var col aurora.Value
		switch level {
		case LevelInfo:
			col = aurora.BgBlue(specialSubject).Black()
			break
		case LevelWarn:
			col = aurora.BgYellow(specialSubject).Black()
			break
		case LevelError:
			col = aurora.BgRed(specialSubject).Black().SlowBlink()
			break
		case LevelSuccess:
			col = aurora.BgGreen(specialSubject).Black()
			break
		}
		baseStr += fmt.Sprintf("[%s] ", col)
	}
	fmt.Println(baseStr + message)
}

package tish

import (
	"bytes"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/color"
)

var (
	homeDirPrefix = "~" + string(filepath.Separator)

	timeColor = color.S256(255, 39)
	userColor = color.S256(238, 81)
	userSepColor = color.S256(100, 81)
	pathColor = color.S256(238, 159)
	pathSepColor = color.S256(166, 159)
	statusOkColor = color.S256(28, 195)
	statusNgColor = color.S256(196, 195)
	cursorColor = color.C256(61)
)

// https://github.com/chris-marsh/pureline
func Prompt(user, host, workdir, homedir string, now time.Time, status int, test bool) string {
	var buffer bytes.Buffer
	buffer.WriteString(timeColor.Sprint(now.Format("  🕓15:04:05  ")))
	if test {
		buffer.WriteString(userColor.Sprint("  👤" + user + "@" + host + "  "))
	} else {
		buffer.WriteString(userColor.Sprint("  👤" + user))
		buffer.WriteString(userSepColor.Sprint("@"))
		buffer.WriteString(userColor.Sprint(host + "  "))
	}

	var pathStr string
	if rel, err := filepath.Rel(homedir, workdir); err == nil {
		if rel == "." {
			pathStr = "~"
		} else if strings.HasPrefix(rel, "..") {
			pathStr = workdir
		} else {
			pathStr = homeDirPrefix + strings.TrimPrefix(rel, homeDirPrefix)
		}
	} else {
		pathStr = workdir
	}
	if test {
		buffer.WriteString(pathColor.Sprint("  📁" + pathStr + "  "))
	} else {
		var fragments []string
		if pathStr == "~" {
			fragments = append(fragments, "~")
		} else {
			for {
				dir, file := filepath.Split(pathStr)
				if file == "" {
					if dir == "/" {
						fragments = append(fragments, "")
					}
					break
				}
				fragments = append(fragments, file)
				pathStr = strings.TrimSuffix(dir, string(filepath.Separator))
			}
		}
		buffer.WriteString(pathColor.Sprint("  📁"))
		for i := len(fragments)-1; i >= 0; i-- {
			f := fragments[i]
			if f != "" {
				buffer.WriteString(pathColor.Sprint(f))
			}
			if i != 0 {
				buffer.WriteString(pathSepColor.Sprint("/"))
			}
		}
		buffer.WriteString(pathColor.Sprint("  "))
	}
	if status != 0 {
		buffer.WriteString(statusNgColor.Sprint("  ✘  "))
	} else {
		buffer.WriteString(statusOkColor.Sprint("  ✔︎  "))
	}
	buffer.WriteString(cursorColor.Sprint("\n≫ "))
	return buffer.String()
}

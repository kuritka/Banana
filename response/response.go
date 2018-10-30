package response

import (
	"github.com/Deutsche-Boerse/edt-sftp/constants"
	"path/filepath"
	"time"
)

type Acknowledge struct {
	Name    string
	Content []byte
}

//GetAcknowledge gets acknowledge for file specified by sourcepath
func GetAcknowledge(zipFilePath string) Acknowledge {
	fileName := filepath.Base(zipFilePath)
	ackName := fileName + constants.RESPONSE
	timestamp := time.Now().Format("20060102T150405")
	content := fileName + ";" + timestamp
	return Acknowledge{ackName, []byte(content)}
}

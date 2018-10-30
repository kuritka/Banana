package response

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"bytes"
	"github.com/Deutsche-Boerse/edt-sftp/utils"
)

var testData = struct {
	InPath  string
	OutPath string
}{
	filepath.Join("testdata", "in"),
	filepath.Join("testdata", "out"),
}

func TestGetAcknowledgementSuccessfully(t *testing.T) {

	// arrange
	file := filepath.Join(testData.InPath, "KV0011_T_EDT_Warrant01.zip")

	// act
	acknowledge := GetAcknowledge(file)

	//assert
	assert.Equal(t, "KV0011_T_EDT_Warrant01.zip.response", acknowledge.Name)
	timestamp := string(bytes.Split(acknowledge.Content[:], []byte(";"))[1])
	assert.Equal(t, filepath.Base(file)+";"+timestamp, string(acknowledge.Content[:]))
}

func TestGetAcknowledgementForNonExistingFile(t *testing.T) {
	//arrange
	file := filepath.Join(testData.InPath, "KV0011_T_EDT_Warrant05.zip")

	// act
	acknowledge := GetAcknowledge(file)

	//assert
	assert.Equal(t, "KV0011_T_EDT_Warrant05.zip.response", acknowledge.Name)
	timestamp := string(bytes.Split(acknowledge.Content[:], []byte(";"))[1])
	assert.Equal(t, filepath.Base(file)+";"+timestamp, string(acknowledge.Content[:]))
}

func TestMain(m *testing.M) {
	//if anything failed before and files are still present
	utils.RemoveAllExcept(filepath.Join(testData.OutPath, ".gitkeep"))
	m.Run()
	//cleaning
	utils.RemoveAllExcept(filepath.Join(testData.OutPath, ".gitkeep"))
}

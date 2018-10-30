package unzip

import (
	"path/filepath"
	"testing"

	"github.com/ahmetb/go-linq"
	"github.com/stretchr/testify/assert"

	"github.com/Deutsche-Boerse/edt-sftp/utils"
)

var testData = struct {
	InPath  string
	OutPath string
}{
	filepath.Join("testdata", "in"),
	filepath.Join("testdata", "out"),
}

func TestUnzipCorrectZipFile(t *testing.T) {
	//arrange

	//act
	unzippedFiles, err := Unzip(filepath.Join(testData.InPath, "p001-1234-XY0011_CB8899-EdtCertUpload.zip"), testData.OutPath)

	//assert
	assert.NoError(t, err)
	assert.True(t, linq.From(unzippedFiles).Contains(filepath.Join(testData.OutPath, "pp_20180808_145332.xml")), "expected "+filepath.Join(testData.OutPath, "pp_20180808_145332.xml"))
	assert.True(t, linq.From(unzippedFiles).Contains(filepath.Join(testData.OutPath, "pp_a121212_aa.XmL")), "expected "+filepath.Join(testData.OutPath, "PPCZ02_180808-145332.xml"))
}

func TestUnzipEmptyZipFile(t *testing.T) {
	//arrange

	//act
	unzippedFiles, err := Unzip(filepath.Join(testData.InPath, "empty.zip"), testData.OutPath)

	//assert
	assert.NoError(t, err)
	assert.Equal(t, 0, len(unzippedFiles))
}

func TestUnzipBrokenZipFile(t *testing.T) {
	//arrange

	//act
	unzippedFiles, err := Unzip(filepath.Join(testData.InPath, "broken.zip"), testData.OutPath)

	//assert
	assert.Nil(t, unzippedFiles)
	assert.NotNil(t, err, "err cannot be nil")
}

func TestUnzipNonExistingZipFile(t *testing.T) {
	//arrange

	//act
	unzippedFiles, err := Unzip(filepath.Join(testData.InPath, "non_existing.zip"), testData.OutPath)

	//assert
	assert.Nil(t, unzippedFiles)
	assert.Error(t, err)
}

func TestUnzipTwoZipFilesInRow(t *testing.T) {
	//arrange

	//act
	unzippedFiles, err := Unzip(filepath.Join(testData.InPath, "KV0011_T_EDT_Warrant02.zip"), testData.OutPath)
	assert.NoError(t, err)
	unzippedFiles2, err := Unzip(filepath.Join(testData.InPath, "KV0011_T_EDT_Warrant01.zip"), testData.OutPath)

	//assert
	assert.NoError(t, err, "err for second unzipped file cannot obtain")
	assert.True(t, linq.From(unzippedFiles2).Contains(filepath.Join(testData.OutPath, ".gitignore")), "expected "+filepath.Join(testData.OutPath, ".gitignore"))
	assert.True(t, linq.From(unzippedFiles).Contains(filepath.Join(testData.OutPath, "README.md")), "expected "+filepath.Join(testData.OutPath, "README.md"))
}

func TestMain(m *testing.M) {
	//if anything fails before and files are still present
	utils.RemoveAllExcept(filepath.Join(testData.OutPath, ".gitkeep"))
	m.Run()
	//cleaning
	utils.RemoveAllExcept(filepath.Join(testData.OutPath, ".gitkeep"))
}

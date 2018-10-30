package host2host

import (
	"fmt"
	"github.com/Deutsche-Boerse/edt-sftp/conf"
	"github.com/Deutsche-Boerse/edt-sftp/constants"
	"github.com/Deutsche-Boerse/edt-sftp/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"testing"
)

var (
	ErrCorruptedFileMustNotExists = "corrupted file must not be kept in target directory"
	ErrResponseCannotBeSend       = "response cannot be sent"
	ErrResponseMustExist          = "expected response"
	ErrEdtFileMustExist           = ".edt file must exist"
	ErrEdtFileMustNotExist        = ".edt file must NOT exist"
	ErrExpectedConfiguration      = "missing configuration"
	ErrExpectedFileExists         = "expected file exist"
	ErrZipMustBeDeleted           = "expected .zip is removed"
	ErrDiffResponses              = "response is not rewritten"
	Skip                          = "skipping test in short mode"
)

var testData = struct {
	InPath           string
	OutPathLocal     string
	OutPathSftpCoba  string
	OutPathSftpBrcls string
	OutPathSftpEmpty string
}{
	filepath.Join("testdata", "in"),
	filepath.Join("testdata", "out", "local"),
	filepath.Join("/home/ec2-user/COBA/"),
	filepath.Join("/home/ec2-user/BRCLS/"),
	filepath.Join("/home/ec2-user/Empty/"),
}

func TestSuccessfullyDownloadedOneFile(t *testing.T) {
	const tested = "KV1212_T_EDT_Bonds180808.zip"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, err := testInit(testData.OutPathSftpBrcls, tested, tested+"_0")

	//act
	downloaded, err := Download(config)

	//assert
	var exists bool
	assert.NoError(t, err)
	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE), &exists).Close()
	assert.True(t, exists, ErrResponseMustExist)
	assert.Equal(t, 1, len(downloaded))
	assert.Equal(t, 3, len(downloaded[0].Unzipped))

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.EDT), &exists).Close()
	assert.False(t, exists, ErrEdtFileMustExist)

	assertResponse(t, tested, testData.OutPathSftpBrcls)
}

func TestSuccessfullyMultipleFiles(t *testing.T) {
	const tested1 = "KV1212_T_EDT_Warrants180808.zip"
	const tested2 = "KV1212_T_EDT_Bonds180808.zip"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	_, err := testInit(testData.OutPathSftpBrcls, tested1, tested1+"_0")
	assert.NoError(t, err, ErrExpectedConfiguration)
	config, err := testInit(testData.OutPathSftpCoba, tested2, tested2+"_0")
	assert.NoError(t, err, ErrExpectedConfiguration)

	//act
	downloaded, err := Download(config)

	//assert
	assert.NoError(t, err)
	var exists bool
	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested1+constants.RESPONSE), &exists).Close()
	assert.True(t, exists, ErrResponseMustExist)
	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpCoba, tested2+constants.RESPONSE), &exists).Close()
	assert.True(t, exists, ErrResponseMustExist)
	assert.Equal(t, 2, len(downloaded))
	assert.Equal(t, 2, len(downloaded[0].Unzipped))
	assert.Equal(t, 3, len(downloaded[1].Unzipped))

	assertResponse(t, tested1, testData.OutPathSftpBrcls)
	assertResponse(t, tested1, testData.OutPathSftpCoba)

	exists, err = utils.Exists(path.Join(config.DstPath, tested1))
	assert.NoError(t, err)
	assert.False(t, exists, ErrZipMustBeDeleted)

	exists, err = utils.Exists(path.Join(config.DstPath, tested2))
	assert.NoError(t, err)
	assert.False(t, exists, ErrZipMustBeDeleted)
}

func TestDownloadCorruptedFile(t *testing.T) {
	const tested = "KV0011_T_EDT_Corrupted.zip"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, err := testInit(testData.OutPathSftpBrcls, tested, tested+"_0")
	assert.NoError(t, err, ErrExpectedConfiguration)

	//act
	downloaded, err := Download(config)

	//assert
	assert.NoError(t, err)
	assert.Error(t, downloaded[0].Error)
	assert.Equal(t, 1, len(downloaded))

	exists, err := utils.Exists(path.Join(config.DstPath, tested))
	assert.NoError(t, err)
	assert.False(t, exists, ErrCorruptedFileMustNotExists)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE), &exists).Close()
	assert.False(t, exists, ErrResponseCannotBeSend)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.EDT), &exists).Close()
	assert.True(t, exists, ErrEdtFileMustExist)
}

func TestDownloadEmptyFile(t *testing.T) {

	const tested = "KV0011_T_EDT_Empty.zip"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, err := testInit(testData.OutPathSftpBrcls, tested, tested+"_0")
	assert.NoError(t, err, ErrExpectedConfiguration)

	//act
	downloaded, err := Download(config)

	//assert
	assert.NoError(t, err)
	assert.Error(t, downloaded[0].Error)
	assert.Equal(t, 1, len(downloaded))

	exists, err := utils.Exists(path.Join(config.DstPath, tested))
	assert.NoError(t, err)
	assert.False(t, exists, ErrCorruptedFileMustNotExists)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE), &exists).Close()
	assert.False(t, exists, ErrResponseCannotBeSend)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.EDT), &exists).Close()
	assert.True(t, exists, ErrEdtFileMustExist)
}

func TestDownloadWithoutZeroLenFile(t *testing.T) {
	const tested = "KV0011_T_EDT_Warrant01.zip"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, err := testInit(testData.OutPathSftpBrcls, tested)
	assert.NoError(t, err, ErrExpectedConfiguration)

	//act
	downloaded, err := Download(config)

	//assert
	assert.NoError(t, err)
	assert.Empty(t, downloaded)

	exists, err := utils.Exists(path.Join(config.DstPath, tested))
	assert.NoError(t, err)
	assert.False(t, exists)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested), &exists).Close()
	assert.True(t, exists, ErrExpectedFileExists)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE), &exists).Close()
	assert.False(t, exists, ErrResponseCannotBeSend)

	exists, err = utils.Exists(path.Join(config.DstPath, tested+constants.EDT))
	assert.NoError(t, err)
	assert.False(t, exists, ErrEdtFileMustNotExist)
}

func TestDownloadWithZeroLenFileOnly(t *testing.T) {
	const tested = "KV0011_T_EDT_Warrant01.zip_0"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, err := testInit(testData.OutPathSftpBrcls, tested)
	assert.NoError(t, err, ErrExpectedConfiguration)

	//act
	downloaded, err := Download(config)

	//assert
	assert.NoError(t, err)
	assert.Empty(t, downloaded)

	exists, err := utils.Exists(path.Join(config.DstPath, tested))
	assert.NoError(t, err)
	assert.False(t, exists)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested), &exists).Close()
	assert.True(t, exists, ErrExpectedFileExists)

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE), &exists).Close()
	assert.False(t, exists, ErrResponseCannotBeSend)

	exists, err = utils.Exists(path.Join(config.DstPath, tested+constants.EDT))
	assert.NoError(t, err)
	assert.False(t, exists, ErrEdtFileMustNotExist)
}

func TestAckAlreadyExists(t *testing.T) {
	const tested = "KV1212_T_EDT_Bonds180808.zip"
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, _ := testInit(testData.OutPathSftpBrcls, tested, tested+"_0", tested+constants.RESPONSE)

	//act
	downloaded, err := Download(config)

	//assert
	var exists bool
	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE), &exists).Close()
	assert.True(t, exists, ErrResponseMustExist)
	assert.Equal(t, 3, len(downloaded[0].Unzipped))

	utils.CreateClient().Exists(filepath.Join(testData.OutPathSftpBrcls, tested+constants.EDT), &exists).Close()
	assert.False(t, exists, ErrEdtFileMustExist)

	utils.CreateClient().LinkFromRemoteToLocal(filepath.Join(testData.OutPathSftpBrcls, tested+constants.RESPONSE),
		filepath.Join(testData.OutPathLocal, tested+constants.RESPONSE)).Close()
	bytesa, err := ioutil.ReadFile(filepath.Join(testData.OutPathLocal, tested+constants.RESPONSE))
	assert.NoError(t, err)
	bytesb, err := ioutil.ReadFile(filepath.Join(testData.InPath, tested+constants.RESPONSE))
	assert.NoError(t, err)

	a := string(bytesa)
	b := string(bytesb)
	ok, err := regexp.MatchString(`KV1212_T_EDT_Bonds180808.zip;\d+T\d+`, a)
	assert.NoError(t, err)
	ok, err = regexp.MatchString(`KV1212_T_EDT_Bonds180808.zip;\d+T\d+`, b)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NotEqual(t, a, b, ErrDiffResponses)
}

func TestDownloadEmptyFolders(t *testing.T) {
	//arrange
	//act
	config, _ := testInit(testData.OutPathSftpEmpty)
	downloaded, err := Download(config)

	//assert
	assert.NoError(t, err)
	assert.Empty(t, downloaded)
}

func TestNilConfig(t *testing.T) {
	//arrange
	//act
	downloaded, err := Download(nil)

	//assert
	assert.Error(t, err)
	assert.Empty(t, downloaded)
}

// tests are running against sftp server
func TestMain(m *testing.M) {
	utils.CreateClient().CreateDirIfNotExists(
		testData.OutPathSftpBrcls).CreateDirIfNotExists(
		testData.OutPathSftpCoba).CreateDirIfNotExists(
		testData.OutPathSftpEmpty).Close()
	utils.RemoveAllExcept(filepath.Join(testData.OutPathLocal, ".gitkeep"))
	m.Run()
	utils.RemoveAllExcept(filepath.Join(testData.OutPathLocal, ".gitkeep"))
	utils.CreateClient().RemoveDir(
		testData.OutPathSftpBrcls).RemoveDir(
		testData.OutPathSftpCoba).RemoveDir(
		testData.OutPathSftpEmpty).Close()
}

func testInit(outputDir string, filesNames ...string) (*conf.SftpConfig, error) {
	c := utils.CreateClient()
	c.RemoveDir(outputDir).CreateDirIfNotExists(outputDir)
	for _, f := range filesNames {
		c.LinkFromLocalToRemote(
			filepath.Join(testData.InPath, f),
			filepath.Join(outputDir, f))
	}
	c.Close()
	utils.RemoveAllExcept(filepath.Join(testData.OutPathLocal, ".gitkeep"))
	return conf.NewFactory().Get()
}

func assertResponse(t *testing.T, tested string, remotefolder string) {
	utils.CreateClient().LinkFromRemoteToLocal(filepath.Join(remotefolder, tested+constants.RESPONSE),
		filepath.Join(testData.OutPathLocal, tested+constants.RESPONSE)).Close()
	bytes, err := ioutil.ReadFile(filepath.Join(testData.OutPathLocal, tested+constants.RESPONSE))
	assert.NoError(t, err)
	ok, err := regexp.MatchString(fmt.Sprintf(`%s;\d+T\d+`, tested), string(bytes))
	assert.NoError(t, err)
	assert.True(t, ok)
}

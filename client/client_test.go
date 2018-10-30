package client

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/Deutsche-Boerse/edt-sftp/client/structs"
	"github.com/Deutsche-Boerse/edt-sftp/conf"
	"github.com/Deutsche-Boerse/edt-sftp/constants"
	"github.com/Deutsche-Boerse/edt-sftp/utils"

	"github.com/stretchr/testify/assert"
)

const (
	testedZip      = "KV1212_T_EDT_Bonds180808.zip"
	testedEdt      = "KV1212_T_EDT_Bonds180808.zip" + constants.EDT
	testedResponse = "KV1212_T_EDT_Bonds180808.zip" + constants.RESPONSE
	data1          = "PPCZ01_160101-145332.xml"
	data2          = "PPCZ02_180808-145332.xml"
	tc             = "TCCZ02_180808-145332.pdf"
)

var (
	Skip = "skipping test in short mode"
)

var testData = struct {
	InPath   string
	OutPath  string
	SftpCoba string
}{
	filepath.Join("testdata", "in"),
	filepath.Join("testdata", "out"),
	filepath.Join("/home/ec2-user/COBA/"),
}

func TestSendMultipleFilesRequest(t *testing.T) {
	if testing.Short() {
		t.Skip(Skip)
	}
	//arrange
	config, err := conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)
	testee1 := filepath.Join(testData.InPath, data1)
	testee2 := filepath.Join(testData.InPath, data2)
	termsAndConditions := filepath.Join(testData.InPath, tc)
	downloads := []*structs.DownloadInfo{{Unzipped: []string{testee1, testee2, termsAndConditions}}}

	//act
	err = sftpClient.SendToEdt(downloads)
	//assert
	assert.NoError(t, err)
}

func TestSendOneFilesRequest(t *testing.T) {
	if testing.Short() {
		t.Skip(Skip)
	}
	//arrange
	config, err := conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)
	testee1 := filepath.Join(testData.InPath, data1)
	termsAndConditions := filepath.Join(testData.InPath, tc)
	downloads := []*structs.DownloadInfo{{Unzipped: []string{testee1, termsAndConditions}}}

	//act
	err = sftpClient.SendToEdt(downloads)
	//assert
	assert.NoError(t, err)
}

func TestSendZeroFilesRequest(t *testing.T) {
	if testing.Short() {
		t.Skip(Skip)
	}

	//arrange
	config, err := conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)

	downloads := []*structs.DownloadInfo{{Unzipped: []string{}}}

	//act
	err = sftpClient.SendToEdt(downloads)

	//assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(downloads))
	//assert.Error(t, downloads[0].Error)
}

func TestCleaningWhenEverythingGoesFine(t *testing.T) {
	var err error
	var config *conf.SftpConfig
	var exists bool

	//arrange
	err = copyToRemote(testData.SftpCoba, testedEdt, testedResponse)
	assert.NoError(t, err)
	err = copyToLocal(testData.OutPath, testedZip, data1, data2, tc)
	assert.NoError(t, err)
	config, err = conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)

	downloads := []*structs.DownloadInfo{
		{
			Error:           nil,
			DestinationPath: path.Join(path.Join(testData.OutPath, testedZip)),
			SourcePath:      path.Join(testData.SftpCoba, testedEdt),
			Unzipped: []string{
				path.Join(testData.OutPath, data1),
				path.Join(testData.OutPath, data2),
				path.Join(testData.OutPath, tc),
			},
			ResponsePath:       path.Join(testData.SftpCoba, testedResponse),
			SourcePathOriginal: path.Join(testData.SftpCoba, testedZip),
		},
	}

	//act
	err = sftpClient.Clean(downloads)

	//assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(downloads))
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedResponse), &exists).Close()
	assert.True(t, exists)
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedEdt), &exists).Close()
	assert.False(t, exists)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data1))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data2))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, tc))
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.NoError(t, downloads[0].Error)
}

func TestCleaningWhenSendFailed(t *testing.T) {
	var err error
	var config *conf.SftpConfig
	var exists bool

	//arrange
	err = copyToRemote(testData.SftpCoba, testedEdt)
	assert.NoError(t, err)
	err = copyToLocal(testData.OutPath, testedZip, data1, data2, tc)
	assert.NoError(t, err)
	config, err = conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)

	downloads := []*structs.DownloadInfo{
		{
			Error:           errors.New("fake error"),
			DestinationPath: path.Join(path.Join(testData.OutPath, testedZip)),
			SourcePath:      path.Join(testData.SftpCoba, testedEdt),
			Unzipped: []string{
				path.Join(testData.OutPath, data1),
				path.Join(testData.OutPath, data2),
				path.Join(testData.OutPath, tc),
			},
			ResponsePath:       "",
			SourcePathOriginal: path.Join(testData.SftpCoba, testedZip),
		},
	}
	//act
	err = sftpClient.Clean(downloads)
	//assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(downloads))
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedEdt), &exists).Close()
	assert.True(t, exists)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data1))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data2))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, tc))
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Error(t, downloads[0].Error)
}

func TestCleaningWhenUnzipFailed(t *testing.T) {
	var err error
	var config *conf.SftpConfig
	var exists bool

	//arrange
	err = copyToRemote(testData.SftpCoba, testedEdt)
	assert.NoError(t, err)
	err = copyToLocal(testData.OutPath, testedZip)
	assert.NoError(t, err)
	config, err = conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)

	downloads := []*structs.DownloadInfo{
		{
			Error:              errors.New("fake error"),
			DestinationPath:    path.Join(path.Join(testData.OutPath, testedZip)),
			SourcePath:         path.Join(testData.SftpCoba, testedEdt),
			Unzipped:           []string{},
			ResponsePath:       "",
			SourcePathOriginal: path.Join(testData.SftpCoba, testedZip),
		},
	}
	//act
	err = sftpClient.Clean(downloads)
	//assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(downloads))
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedResponse), &exists).Close()
	assert.False(t, exists)
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedEdt), &exists).Close()
	assert.True(t, exists)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data1))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data2))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, tc))
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Error(t, downloads[0].Error)
}

func TestCleaningWhenDownloadFailed(t *testing.T) {
	var err error
	var config *conf.SftpConfig
	var exists bool

	//arrange
	err = copyToRemote(testData.SftpCoba, testedEdt)
	assert.NoError(t, err)
	config, err = conf.NewFactory().Get()
	assert.NoError(t, err)
	sftpClient, err := NewFactory(ClientOptions{config}).Get()
	assert.NoError(t, err)

	downloads := []*structs.DownloadInfo{
		{
			Error:              errors.New("fake error"),
			DestinationPath:    path.Join(path.Join(testData.OutPath, testedZip)),
			SourcePath:         path.Join(testData.SftpCoba, testedEdt),
			Unzipped:           []string{},
			ResponsePath:       "",
			SourcePathOriginal: path.Join(testData.SftpCoba, testedZip),
		},
	}
	//act
	err = sftpClient.Clean(downloads)
	//assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(downloads))
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedResponse), &exists).Close()
	assert.False(t, exists)
	utils.CreateClient().Exists(filepath.Join(testData.SftpCoba, testedEdt), &exists).Close()
	assert.True(t, exists)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data1))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, data2))
	assert.False(t, exists)
	assert.NoError(t, err)
	exists, err = utils.Exists(filepath.Join(testData.OutPath, tc))
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Error(t, downloads[0].Error)
}

func copyToRemote(remoteDir string, filesNames ...string) error {
	c := utils.CreateClient()
	c.RemoveDir(remoteDir).CreateDirIfNotExists(remoteDir)
	for _, f := range filesNames {
		c.LinkFromLocalToRemote(
			filepath.Join(testData.InPath, f),
			filepath.Join(remoteDir, f))
	}
	c.Close()
	return nil
}

func copyToLocal(outputDir string, filesNames ...string) error {
	if err := utils.RemoveAllExcept(filepath.Join(outputDir, ".gitkeep")); err != nil {
		return err
	}
	for _, f := range filesNames {
		if err := copyFileContents(filepath.Join(testData.InPath, f), filepath.Join(testData.OutPath, f)); err != nil {
			return err
		}
	}
	return nil
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

package util

func MxcUri(origin string, mediaId string) string {
	return "mxc://" + origin + "/" + mediaId
}

// GK CUSTOMIZATION: Supported file types
var supportedFileTypes = []string{
	"docx",
	"doc",
	"xlsx",
	"csv",
	"txt",
	"pdf",
	"ppr",
	"ppt",
	"jpg",
	"png",
	"gif",
	"heic",
	"mpeg4",
	"mpeg",
	"h264",
	"avi",
}

func IsSupportedFileType(fileType string) bool {
	for _, v := range supportedFileTypes {
		if v == fileType {
			return true
		}
	}
	return false
}

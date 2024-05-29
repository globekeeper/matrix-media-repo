package util

func MxcUri(origin string, mediaId string) string {
	return "mxc://" + origin + "/" + mediaId
}

func IsSupportedFileType(fileType string, supportedFileTypes []string) bool {
	for _, v := range supportedFileTypes {
		if v == fileType {
			return true
		}
	}
	return false
}

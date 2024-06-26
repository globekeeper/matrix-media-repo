package config

type MinimumRepoConfig struct {
	DataStores     []DatastoreConfig `yaml:"datastores"`
	Archiving      ArchivingConfig   `yaml:"archiving"`
	Uploads        UploadsConfig     `yaml:"uploads"`
	Identicons     IdenticonsConfig  `yaml:"identicons"`
	Quarantine     QuarantineConfig  `yaml:"quarantine"`
	TimeoutSeconds TimeoutsConfig    `yaml:"timeouts"`
	Features       FeatureConfig     `yaml:"featureSupport"`
	AccessTokens   AccessTokenConfig `yaml:"accessTokens"`
}

func NewDefaultMinimumRepoConfig() MinimumRepoConfig {
	return MinimumRepoConfig{
		DataStores: []DatastoreConfig{},
		Archiving: ArchivingConfig{
			Enabled:            true,
			SelfService:        false,
			TargetBytesPerPart: 209715200, // 200mb
		},
		Uploads: UploadsConfig{
			MaxFilenameLength: 24,
			SupportedFileTypes: []string{
				"docx",
				"doc",
				"xlsx",
				"xls",
				"odt",
				"ods",
				"odp",
				"csv",
				"txt",
				"pdf",
				"ppr",
				"pptx",
				"ppt",
				"jpg",
				"png",
				"gif",
				"heic",
				"bmp",
				"webp",
				"svg",
				"mpeg4",
				"mpeg",
				"h264",
				"webm",
				"mkv",
				"avi",
				"mov",
				"mp3",
				"wav",
				"flac",
				"zip",
			},
			MaxSizeBytes:         104857600, // 100mb
			MinSizeBytes:         100,
			ReportedMaxSizeBytes: 0,
			MaxPending:           5,
			MaxAgeSeconds:        1800, // 30 minutes
			Quota: QuotasConfig{
				Enabled:    false,
				UserQuotas: []QuotaUserConfig{},
			},
		},
		Identicons: IdenticonsConfig{
			Enabled: true,
		},
		Quarantine: QuarantineConfig{
			ReplaceThumbnails: true,
			ReplaceDownloads:  false,
			ThumbnailPath:     "",
			AllowLocalAdmins:  true,
		},
		TimeoutSeconds: TimeoutsConfig{
			UrlPreviews:  10,
			ClientServer: 30,
			Federation:   120,
		},
		Features: FeatureConfig{},
		AccessTokens: AccessTokenConfig{
			MaxCacheTimeSeconds: 0,
			UseAppservices:      false,
			Appservices:         []AppserviceConfig{},
		},
	}
}

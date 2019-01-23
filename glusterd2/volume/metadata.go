package volume

const (
	// BlockHostingAvailableSize is a volume metadata to store available size of hosting volume to create block devices.
	BlockHostingAvailableSize = "_block-hosting-available-size"
	// BlockHostingVolumeAutoCreated is volume metadata which will be set as `yes` if gd2 auto create the hosting volume
	BlockHostingVolumeAutoCreated = "block-hosting-volume-auto-created"
	// BlockHosting is a volume metadata which will be set as `yes' for volumes which are able to host block devices.
	BlockHosting = "block-hosting"
	// BlockPrefix is the prefix of the volume metadata which will contain BlockPrefix + blockname as the key and size of the block as value.
	BlockPrefix = "block-vol:"
)

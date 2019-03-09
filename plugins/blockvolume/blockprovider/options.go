package blockprovider

// BlockVolOption configures various optional parameters for a block operation
type BlockVolOption func(*BlockVolumeOptions)

// BlockVolumeOptions represents various optional params to be used for a block volume operation
type BlockVolumeOptions struct {
	Auth               bool
	FullPrealloc       bool
	Storage            string
	Ha                 int
	RingBufferSizeInMB uint64
	ForceDelete        bool
	UnlinkStorage      bool
	Hosts              []string
	BlockType          string
}

// ApplyOpts applies configured optional parameters on BlockVolumeOptions
func (op *BlockVolumeOptions) ApplyOpts(optFuncs ...BlockVolOption) {
	for _, optFunc := range optFuncs {
		optFunc(op)
	}
}

// WithHaCount configures haCount for block creation
func WithHaCount(count int) BlockVolOption {
	return func(options *BlockVolumeOptions) {
		options.Ha = count
	}
}

// WithStorage configures storage param for block-vol creation
func WithStorage(storage string) BlockVolOption {
	return func(options *BlockVolumeOptions) {
		options.Storage = storage
	}
}

// WithRingBufferSizeInMB configures ring-buffer param (size should in MB units)
func WithRingBufferSizeInMB(size uint64) BlockVolOption {
	return func(options *BlockVolumeOptions) {
		options.RingBufferSizeInMB = size
	}
}

// WithForceDelete configures force param in a block delete req
func WithForceDelete(options *BlockVolumeOptions) {
	options.ForceDelete = true
}

// WithUnlinkStorage configures unlink-storage param in block delete req
func WithUnlinkStorage(options *BlockVolumeOptions) {
	options.UnlinkStorage = true
}

// WithAuthEnabled enables auth for block creation
func WithAuthEnabled(options *BlockVolumeOptions) {
	options.Auth = true
}

// WithFullPrealloc configures "prealloc" param
func WithFullPrealloc(options *BlockVolumeOptions) {
	options.FullPrealloc = true
}

// WithHosts configures required hosts for block creation
func WithHosts(hosts []string) BlockVolOption {
	return func(options *BlockVolumeOptions) {
		options.Hosts = hosts
	}
}

// WithBlockType configures the block type
func WithBlockType(blockType string) BlockVolOption {
	return func(options *BlockVolumeOptions) {
		if blockType == "" {
			options.BlockType = "ext4"
		} else {
			options.BlockType = blockType
		}
	}
}

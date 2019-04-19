package disk

type FileSystemType string

const (
	FileSystemSwap    FileSystemType = "swap"
	FileSystemExt4    FileSystemType = "ext4"
	FileSystemXFS     FileSystemType = "xfs"
	FileSystemLUKS    FileSystemType = "crypto_LUKS"
	FileSystemDefault FileSystemType = ""
)

type Formatter interface {
	Format(partitionPath string, fsType FileSystemType) (err error)
}

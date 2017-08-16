package structer

type PackageKind string

const (
	NoPackage     PackageKind = ""
	VendorPackage             = "vendor"
	SystemPackage             = "system"
	UserPackage               = "user"
)

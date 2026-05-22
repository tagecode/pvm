package win

// EnvStatus summarizes PVM-related user environment configuration.
type EnvStatus struct {
	PVMHome     string
	PHPHome     string
	PHPRC       string
	Path        string
	InPath      bool
	PathIndex   int
	PathAtFront bool // true when current link is the first entry in user PATH
}

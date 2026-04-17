package buildinfo

import "fmt"

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func String() string {
	return fmt.Sprintf("GophKeeper %s (built %s)", Version, BuildDate)
}

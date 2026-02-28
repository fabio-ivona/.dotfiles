package shared

type Config struct {
	Type    string
	TypeSet bool
	Force   bool
	Verbose bool
	BaseDir string
	Token   string
	Repo    string
	OldTag  string
	OldVer  string
	NewVer  string
	NewTag  string
	Changes string
	Release string
}

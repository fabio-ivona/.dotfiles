package shared

type Config struct {
	Type      string
	TypeSet   bool
	Force     bool
	Verbosity int
	BaseDir   string
	Token     string
	Repo      string
	OldTag    string
	OldVer    string
	NewVer    string
	NewTag    string
	Changes   string
	Release   string
}

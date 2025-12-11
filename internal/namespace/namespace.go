package namespace

type Namespace interface {
	Attach() error
	Exec(cmd []string) error
}

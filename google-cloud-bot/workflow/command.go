package workflow

const (
	CommandSudo = "sudo"
)

type Command interface {
	GetCommandName() string
}

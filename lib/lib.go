package lib

const (
	ActionAdd    Action = "add"
	ActionRemove Action = "remove"
	ActionOutput Action = "output"

	CaseRemovePrefix CaseRemove = 0
	CaseRemoveEntry  CaseRemove = 1
)

var ActionsRegistry = map[Action]bool{
	ActionAdd:    true,
	ActionRemove: true,
	ActionOutput: true,
}

type Action string

type CaseRemove int

type Typer interface {
	GetType() string
}

type Actioner interface {
	GetAction() Action
}

type Descriptioner interface {
	GetDescription() string
}

type InputConverter interface {
	Typer
	Actioner
	Descriptioner
	Input(Container) (Container, error)
}

type OutputConverter interface {
	Typer
	Actioner
	Descriptioner
	Output(Container) error
}

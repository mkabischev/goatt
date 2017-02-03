package goatt

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pborman/uuid"
)

type Session map[string]interface{}

func (session Session) UUID() string {
	_, ok := session["uuid"]
	if !ok {
		session["uuid"] = uuid.New()
	}
	return session["uuid"].(string)
}

type Step map[string]string

func (step Step) UUID() string {
	_, ok := step["uuid"]
	if !ok {
		step["uuid"] = uuid.New()
	}
	fmt.Println("\nstep uuid", ok, step)
	return step["uuid"]
}

type Context struct {
	Constants interface{}
	Step      Step
	Session   Session
	Result    interface{}
}

func (ctx Context) UUID() string {
	return uuid.New()
}

func (ctx Context) ClearStep() {
	ctx.Step = Step{}
}

func (ctx Context) Evaluate(tmpl string) (string, error) {
	template, err := template.New(uuid.New()).Parse(tmpl)
	if err != nil {
		panic(fmt.Sprintf("unable parse subject %q as a template", tmpl))
	}

	buffer := bytes.NewBuffer([]byte{})
	if err := template.Execute(buffer, ctx); err != nil {
		panic(fmt.Sprintf("failed to render template [%s]\n[%s]\n", err, ctx))
		return "", err
	}
	return buffer.String(), nil
}

func InitContext(consts map[string]interface{}) *Context {
	ctx := Context{
		Constants: consts,
		Step:      Step{},
		Session: Session{
			"stepIdx": 0,
		},
	}

	return &ctx
}

package checker

import (
	"bufio"
	"fmt"

	"github.com/pkg/errors"
)

var (
	// Foo is readonly
	Foo = 10
	Bar = "hello world"
)

func Hoge() {
	var hello int
	Foo = 100

	hello = 10

	// overwrite
	Bar = "See You." // foo bar

	bufio.ErrBufferFull = fmt.Errorf("bufio: buffer full kamata")
	fmt.Println(hello)

	e := errors.New("error is error")
	fmt.Println(e)

	Foo, Bar = 200, "hoge"
	fmt.Println(Foo, Bar)
}

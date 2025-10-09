package filmdata

import (
	"fmt"
)

type FilmList struct {
	Name    string
	ListUrl string
	Films   []*Film
}

func (fl FilmList) String() string {
	return fmt.Sprintf("{%s: %s}", fl.Name, fl.ListUrl)
}

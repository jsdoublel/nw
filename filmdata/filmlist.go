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

// func MakeFilmList(listUrl string) (FilmList, error) {
// 	name, filmUrls, err := ScrapeList(listUrl)
// 	if err != nil {
// 		return FilmList{}, err
// 	}
// 	films := make([]*Film, 0, len(filmUrls))
// 	for _, url := range filmUrls {
// 		film, err := MakeFilm(url)
// 		if err != nil {
// 			log.Printf("failed to retrieve film %s, %s", url, err)
// 		}
// 		films = append(films, film)
// 	}
// 	return FilmList{Name: name, ListUrl: listUrl, Films: films}, nil
// }

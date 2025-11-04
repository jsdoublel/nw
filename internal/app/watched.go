package app

type WatchedFilms struct {
	Films map[int]*Film
}

func (wf WatchedFilms) Watched(film *Film) bool {
	_, ok := wf.Films[film.LBxdID]
	return ok
}

func (wf WatchedFilms) IsZero() bool {
	return wf.Films == nil
}

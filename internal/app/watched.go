package app

func (a *Application) Watched(film *Film) bool {
	_, ok := a.WatchedFilms[film.LBxdID]
	return ok
}

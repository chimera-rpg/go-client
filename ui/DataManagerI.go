package ui

type DataManagerI interface {
	GetDataPath(...string) string
}

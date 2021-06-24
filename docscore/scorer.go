package main

type scorer interface {
	Score(filenames []fileEntry) []fileEntry
}

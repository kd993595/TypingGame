package wordsystem

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"
)

type WordSystem struct {
	allwords       []string
	usedWords      map[string]struct{}
	WordComponents []Word
	randgen        *rand.Rand
}

//returns wordsystem struct with fields intialized with filename passed to retrieve word list
func InitializeWordSystem(file string) WordSystem {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	t_allwords := strings.Fields(string(f))
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return WordSystem{allwords: t_allwords, usedWords: make(map[string]struct{}), WordComponents: make([]Word, 0, 20), randgen: r1}
}

func (w *WordSystem) GetWord() string {
	randint := w.randgen.Intn(len(w.allwords))
	t_word := w.allwords[randint]
	if _, ok := w.usedWords[t_word]; ok {
		t_word = w.GetWord()
	}
	w.usedWords[t_word] = struct{}{}
	return t_word
}

func (w *WordSystem) RemoveWordFromSet(s string) {
	delete(w.usedWords, s)
}

//removes all word entities that have entitylinked field match their prefix given word
func (w *WordSystem) RemoveWordEntities(word string) {
	b := make([]Word, 0, len(w.WordComponents))
	for i := 0; i < len(w.WordComponents); i++ {
		if !strings.HasPrefix(w.WordComponents[i].EntityLinked, word) {
			b = append(b, w.WordComponents[i])
		}
	}
	w.WordComponents = b
}

func (w *WordSystem) CreateWordEntity(entity string, posx, posy int) {
	w.WordComponents = append(w.WordComponents, Word{WordChar: w.GetWord(), EntityLinked: entity, X: posx, Y: posy, Highlighted: false})
}

func (w *WordSystem) CheckChars(characters string) (bool, string) {
	for i := range w.WordComponents {
		res := strings.HasPrefix(w.WordComponents[i].WordChar, characters)
		if res && characters != "" {
			w.WordComponents[i].Highlighted = true
			if characters == w.WordComponents[i].WordChar {
				tmp := w.WordComponents[i].WordChar
				w.WordComponents[i].WordChar = w.GetWord()
				w.RemoveWordFromSet(tmp)
				return true, w.WordComponents[i].EntityLinked
			}
		} else {
			w.WordComponents[i].Highlighted = false
		}
	}
	return false, ""
}

type Word struct {
	WordChar     string
	EntityLinked string
	X            int
	Y            int
	Highlighted  bool
}

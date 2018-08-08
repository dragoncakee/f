package mimic

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strings"
)

var (
	usersData map[string]map[string]map[string]int
)

const (
	messageStarter = "((!BEGINNING_OF_MESSAGE!))"
)

func init() {
	usersData = make(map[string]map[string]map[string]int)
}

func Build(m *discordgo.Message) {
	text := m.Content

	if len(text) == 0 || text[0] == '!' {
		return
	}

	userData := getUserData(m.Author.Username)

	// TODO: remove chained whitespace
	// TODO: ignore links (or not?)
	words := strings.Split(text, " ")
	putWord(messageStarter, words[0], userData)
	for i := 0; i < len(words)-1; i++ {
		putWord(words[i], words[i+1], userData)
	}
}

func putWord(leading string, trailing string, userData map[string]map[string]int) {
	if wordTrails, ok := userData[leading]; ok {
		if _, ok := wordTrails[trailing]; ok {
			wordTrails[trailing]++
		} else {
			wordTrails[trailing] = 1
		}
	} else {
		m := make(map[string]int)
		m[trailing] = 1
		userData[leading] = m
	}
}

func getUserData(username string) map[string]map[string]int {
	if val, ok := usersData[username]; ok {
		return val
	}
	userData := make(map[string]map[string]int)
	usersData[username] = userData
	return userData
}

func Generate(username string, starter string) string {
	if _, ok := usersData[username]; !ok {
		return ""
	}

	userData := usersData[username]

	var err error
	var word = starter
	if starter == "" {
		word, err = selectWord(userData[messageStarter])
	}

	if err != nil {
		log.Error(err)
	}
	message := []string{word}
	err = nil
	for {
		word, err = selectWord(userData[word])
		if err != nil || len(message) > 1000 {
			break
		}
		message = append(message, word)
	}
	return strings.Join(message, " ")
}

func selectWord(wordData map[string]int) (string, error) {
	weightSum := 0
	for _, weight := range wordData {
		weightSum += weight
	}

	if weightSum <= 0 {
		return "", errors.New("no words to use")
	}
	randomWeight := rand.Intn(weightSum) + 1
	for word, weight := range wordData {
		randomWeight -= weight
		if randomWeight <= 0 {
			return word, nil
		}
	}

	return "", errors.New(fmt.Sprintf("no word selected, length of wordMap %d", len(wordData)))
}

func GetStatus(username string) string {
	if username == "" {
		totalWords := 0
		conts := 0
		for _, v := range usersData {
			totalWords += len(v)
			for _, wordConts := range v {
				conts += len(wordConts)
			}
		}

		return fmt.Sprintf(
			"Total of %d words with %d continuations for %d users. Each word has on average %.2f continuations",
			totalWords,
			conts,
			len(usersData),
			float64(conts)/float64(totalWords),
		)
	} else {
		if _, ok := usersData[username]; !ok {
			return ""
		}
		userData := usersData[username]
		totalWords := len(userData)
		conts := 0
		for _, wordConts := range userData {
			conts += len(wordConts)
		}

		return fmt.Sprintf(
			"%s has a total of %d words with %d continuations, each word has an average of %.2f continuations",
			username,
			totalWords,
			conts,
			float64(conts)/float64(totalWords),
		)
	}
}

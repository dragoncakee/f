package mimic

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"math/rand"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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

func Generate(username string) string {
	userData := usersData[username]

	word, err := selectWord(userData[messageStarter])
	if err != nil {
		log.Error(err)
	}
	message := []string{word}
	err = nil
	for {
		word, err = selectWord(userData[word])
		if err != nil {
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
	//randomWeight = rand(1, sumOfWeights)
	//for each item in array
	//	randomWeight = randomWeight - item.Weight
	//	if randomWeight <= 0
	//		break // done, we select this item
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

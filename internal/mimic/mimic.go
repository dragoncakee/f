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
	messageStarter = "((!BEGIN_MESSAGE!))"
	messageEnder   = "((!END_MESSAGE!))"
)

func init() {
	usersData = make(map[string]map[string]map[string]int)
}

func Build(m *discordgo.Message) {
	text := m.Content

	if len(text) == 0 || text[0] == '!' {
		return
	}

	text = strings.ToLower(text)

	words := strings.Split(text, " ")
	if len(words) < 2 {
		return
	}

	// TODO: remove chained whitespace
	// TODO: ignore links (or not?)
	putWord(messageStarter, words[0], m)
	for i := 0; i < len(words)-1; i++ {
		putWord(words[i], words[i+1], m)
		if i < len(words)-2 {
			putWord(words[i]+" "+words[i+1], words[i+2], m)
		}
	}

	l := len(words)
	putWord(words[l-2]+" "+words[l-1], messageEnder, m)
}

func putWord(leading string, trailing string, m *discordgo.Message) {
	userData := getUserData(m.Author.Username)
	singleUserPutWord(leading, trailing, userData)
	if !m.Author.Bot {
		singleUserPutWord(leading, trailing, getUserData("all"))
	}
}

func singleUserPutWord(leading string, trailing string, userData map[string]map[string]int) {
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

	starter = strings.ToLower(starter)

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
		word, err = selectWord(userData[getState(message)])
		if err != nil || len(message) > 1000 || word == messageEnder {
			break
		}
		message = append(message, word)
	}
	return strings.Join(message, " ")
}

func getState(message []string) string {
	l := len(message)
	if l == 1 {
		return message[0]
	}
	return message[l-2] + " " + message[l-1]
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

func DebugSelectWord(input string) string {
	input = strings.ToLower(input)
	
	splits := strings.Split(input, " ")[1:]
	username := splits[0]
	if _, ok := usersData[username]; !ok {
		return "no such user"
	}
	userData := usersData[username]

	content := strings.Join(splits[1:], " ")
	wordData := userData[content]
	selectedWord, err := selectWord(userData[content])
	if err != nil {
		return fmt.Sprintf(
			"Length of wordsdata was %d, and selectWord returned error: %s",
			len(wordData),
			err.Error(),
		)
	}
	return fmt.Sprintf("Selected word returned was %q, words in data were %s", selectedWord, weightsTostring(wordData))
}

func weightsTostring(m map[string]int) string {
	s := []string{"["}
	for k, v := range m {
		s = append(s, k+":"+fmt.Sprintf("%d", v))
	}
	s = append(s, "]")
	return strings.Join(s, " ")
}

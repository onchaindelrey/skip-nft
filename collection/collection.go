package collection

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"

	"skip-nft/utils"
)

const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"

var Logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

type Collection struct {
	Mutex             sync.RWMutex
	Count             int
	BaseUrl           string
	Name              string
	TraitsList        map[string][]string
	Tokens            []*Token
	TokenRarityScores []RarityScorecard
}

func (col *Collection) fetchToken(tid int, colUrl string) *Token {
	url := fmt.Sprintf("%s/%s/%d.json", col.BaseUrl, colUrl, tid)
	res, err := http.Get(url)
	if err != nil {
		Logger.Println(string(COLOR_RED), fmt.Sprintf("Error getting token %d :", tid), err, string(COLOR_RESET))
		return &Token{}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Logger.Println(string(COLOR_RED), fmt.Sprintf("Error reading response for token %d :", tid), err, string(COLOR_RESET))
		return &Token{}
	}
	attrs := make(map[string]string)
	json.Unmarshal(body, &attrs)

	return &Token{
		ID:    tid,
		Attrs: attrs,
	}
}

func (col *Collection) LoadTokenCollection(processThreads int) {
	Logger.Println(string(COLOR_GREEN), fmt.Sprintf("Loading Token Collection %s, with %d threads", col.Name, processThreads), string(COLOR_RESET))

	tokens := make([]*Token, col.Count)

	chunkedTokenList := utils.ChunkBy(tokens, processThreads)

	for threadNum, chunk := range chunkedTokenList {
		var wg sync.WaitGroup

		for i := 0; i < len(chunk); i++ {
			wg.Add(1)

			tokenIdx := (threadNum * processThreads) + i

			go func(idx int) {

				defer wg.Done()

				token := col.fetchToken(idx, col.Name)

				tokenTraits := token.GetTokenTraits()

				col.Mutex.Lock()
				defer col.Mutex.Unlock()

				col.Tokens[idx] = token
				col.addCollectionTraits(tokenTraits)

			}(tokenIdx)

		}

		wg.Wait()
	}

}

func (col *Collection) addCollectionTraits(trait map[string]string) {
	for key, value := range trait {

		if _, exists := col.TraitsList[key]; exists {
			found := false
			for _, v := range col.TraitsList[key] {
				if v == value {
					found = true
					break
				}
			}
			if !found {
				col.TraitsList[key] = append(col.TraitsList[key], value)
			}
		} else {
			col.TraitsList[key] = append(col.TraitsList[key], value)
		}
	}

}

func (col *Collection) CalculateCollectionTokenRarity(processThreads int) {
	Logger.Println(string(COLOR_GREEN), fmt.Sprintf("Calculating %s collection rarity, with %d threads", col.Name, processThreads), string(COLOR_RESET))

	chunkedTokenList := utils.ChunkBy(col.Tokens, processThreads)

	for threadNum, chunk := range chunkedTokenList {

		var wg sync.WaitGroup

		for i, token := range chunk {

			tokenIdx := (threadNum * processThreads) + i

			wg.Add(1)

			go func(t *Token, idx int) {
				defer wg.Done()

				tokenRarity := t.CalculateTokenRarity(col.Tokens, col.TraitsList)

				col.Mutex.Lock()
				defer col.Mutex.Unlock()

				col.TokenRarityScores[idx] = tokenRarity

			}(token, tokenIdx)

		}

		wg.Wait()

	}
}

func (col *Collection) GetTopFive() []RarityScorecard {
	sortedScores := col.TokenRarityScores

	sort.SliceStable(sortedScores, func(i, j int) bool {
		return sortedScores[i].Rarity > sortedScores[j].Rarity
	})

	return sortedScores[:5]
}

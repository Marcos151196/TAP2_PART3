package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/vistarmedia/gossamr"
)

type Task1 struct{}

func main() {
	task1 := gossamr.NewTask(&Task1{})
	err := gossamr.Run(task1)
	if err != nil {
		log.Fatal(err)
	}
}

// CONVERTS FILE TO [DECADE,ARRAY_WITH_ALL_THE_WORD+COUNT_PAIRS_FOR_THAT_DECADE] JUST FOR WORDS STARTIONG WITH "a"
func (wc *Task1) Map(p int64, line string, c gossamr.Collector) error {
	tokens := strings.Fields(line)
	ngram := tokens[0]
	if strings.HasPrefix(ngram, "a") {
		year, err := strconv.Atoi(tokens[1])
		if err != nil {
			return nil
		}
		decade := (int64(year) / 10) * 10
		_, err = strconv.Atoi(tokens[2])
		if err != nil {
			return nil
		}
		match_count := tokens[2]
		if year >= 1800 {
			c.Collect(decade, ngram+"||"+match_count)
		}
	}

	return nil
}

// THIS IS RUN 1 TIME FOR EACH KEY AND RECEIVES [DECADE,ARRAY_WITH_ALL_THE_WORD+COUNT_PAIRS_FOR_THAT_DECADE]
func (wc *Task1) Reduce(key int64, values chan string, c gossamr.Collector) error {
	mapa := map[string]int64{}
	for value := range values {
		tokens := strings.Split(value, "||")
		ngram := tokens[0]
		count, _ := strconv.Atoi(tokens[1])
		mapa[ngram] = mapa[ngram] + int64(count) // Computes total count for a word (in case the same word is the most repeated one for different years within the decade)
	}
	word, count := mapMax(mapa)
	c.Collect(key, fmt.Sprintf("%s\t%d", word, count))
	return nil
}

// GETS [KEY,VALUE] FOR THE MAXIMUM VALUE OF A MAP[KEY,VALUE]
// IF THERE IS MULTIPLE KEYS THAT HAVE THE MAXIMUM VALUE, RETURNS ONLY THE FIRST ALPHABETICALLY ORDERED KEY
func mapMax(mapa map[string]int64) (string, int64) {
	var maxvalue int64 = 0
	var keySlice []string
	for key, value := range mapa {
		if value > maxvalue {
			maxvalue = value
			keySlice = []string{key}
		} else if value == maxvalue {
			keySlice = append(keySlice, key)
			maxvalue = value
		}
	}
	sort.Strings(keySlice)
	maxKey := keySlice[0]
	return maxKey, maxvalue
}

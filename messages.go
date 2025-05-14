package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type TrainbotData struct {
	Posts    []string
	Stations []string
	lastPost int
	// TrainLines []string
}

func readTrainbotData(path string) (TrainbotData, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return TrainbotData{}, err
	}

	var res TrainbotData
	err = yaml.Unmarshal([]byte(data), &res)

	if err != nil {
		return TrainbotData{}, err
	}

	res.lastPost = -1
	return res, nil
}

// replaceRanges finds any instances of substrings of the form "[a-b]", where a an b are numbers,
// and replaces them with a randomly generated integer in that range.
func replaceRanges(post *string) {
	re := regexp.MustCompile(`\[\d+-\d+\]`)

	match := re.FindStringIndex(*post)
	for match != nil {
		// The bounds are adjusted to remove the square brackets around the range
		slice := (*post)[match[0]+1 : match[1]-1]
		lo, hi, found := strings.Cut(slice, "-")
		if !found {
			panic("Unreachable")
		}
		loInt, err := strconv.Atoi(lo)
		must(err)
		hiInt, err := strconv.Atoi(hi)
		must(err)
		num := rand.Intn(hiInt-loInt) + loInt
		*post = (*post)[:match[0]] + fmt.Sprintf("%d", num) + (*post)[match[1]:]

		match = re.FindStringIndex(*post)
	}
}

// newMessage returns a random message picked using the data in t
func (t *TrainbotData) newMessage() string {
	postId := rand.Intn(len(t.Posts))

	// Ensure we dont get the same post twice in a row
	if t.lastPost != -1 {
		for postId == t.lastPost {
			postId = rand.Intn(len(t.Posts))
		}
	}

	post := t.Posts[postId]
	t.lastPost = postId

	if strings.Contains(post, "[station]") {
		station := t.Stations[rand.Intn(len(t.Stations))]
		post = strings.Replace(post, "[station]", station, -1)
	}

	replaceRanges(&post)

	return post
}

package generator

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	prefixes   = []string{"Fastest", "Brightest", "Byter", "Quickee", "Syncee", "Cloudy", "Deeper", "Neo", "Smart"}
	suffixes   = []string{"User", "Cat", "Parrot", "Dog", "Tiger", "Tacker", "Snake", "Wave", "Bit"}
	vowels     = "aeiouy"
	consonants = "bcdfghjklmnpqrstvwxyz"
)

const (
	maxInt = 1000
)

func GenerateName() (string, error) {
	var builder strings.Builder
	var n int
	var err error

	rand.New(rand.NewSource(time.Now().UnixNano()))

	_ = n

	randNum := rand.Intn(maxInt)
	_, err = builder.WriteString(strconv.Itoa(randNum))
	if err != nil {
		return uuid.NewString(), err
	}

	_, err = builder.WriteString(".")
	if err != nil {
		return uuid.NewString(), err
	}

	randIdxPrefix := rand.Intn(len(prefixes))
	_, err = builder.WriteString(prefixes[randIdxPrefix])
	if err != nil {
		return uuid.NewString(), err
	}

	_, err = builder.WriteString("-")
	if err != nil {
		return uuid.NewString(), err
	}

	randIdxSuffix := rand.Intn(len(suffixes))
	_, err = builder.WriteString(suffixes[randIdxSuffix])
	if err != nil {
		return uuid.NewString(), err
	}

	return builder.String(), nil
}

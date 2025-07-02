package models

import (
	"encoding/json"
	"errors"
	"os"
)

type Question struct {
	QuestionId int
	Ques       string
	OptionA    string
	OptionB    string
	OptionC    string
	OptionD    string
	Answer     string
}

func GetQuestionsFromJSON(filename string) ([]Question, error) {

	file, err := os.Open(filename)

	if err != nil {
		return nil, errors.New("unable to open file")
	}

	defer file.Close()

	var questions []Question
	err = json.NewDecoder(file).Decode(&questions)
	if err != nil {
		return nil, errors.New("unable to read file")
	}

	return questions, err
}

func GetQuestionFromId(questions []Question, qId int) (q Question, err error) {

	for index := range questions {
		if questions[index].QuestionId == qId {
			q = questions[index]
		}

	}

	if q.Ques == "" {
		return q, errors.New("invalid question id")
	}

	return q, err
}

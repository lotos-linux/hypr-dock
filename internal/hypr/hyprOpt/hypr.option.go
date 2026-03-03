package hyprOpt

import (
	"errors"
	"fmt"
	"hypr-dock/pkg/ipc"
	"math"
	"strconv"
	"strings"
)

type gapsOut struct {
	Option string `json:"option"`
	Custom string `json:"custom"`
	Set    bool   `json:"set"`
}

func GetGap() ([]int, error) {
	option := "general:gaps_out"

	gapsVal := gapsOut{}
	err := ipc.GetOption(option, &gapsVal)
	if err != nil {
		err = fmt.Errorf("failed to get option \"%s\": %v", option, err)
		return nil, err
	}

	if !gapsVal.Set {
		errorText := fmt.Sprintf("value \"%s\" is unset", option)
		return nil, errors.New(errorText)
	}

	if gapsVal.Custom == "" {
		errorText := fmt.Sprintf("value \"%s\" is empty", option)
		return nil, errors.New(errorText)
	}

	outValues := []int{}
	gapsVal.Custom = strings.TrimSpace(gapsVal.Custom)
	values := strings.Split(gapsVal.Custom, " ")
	for _, value := range values {
		intValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			err = fmt.Errorf("failed to convert \"%s\" to int: %v", value, err)
			return nil, err
		}

		outValues = append(outValues, int(math.Round(math.Max(intValue, 0))))
	}

	return outValues, nil
}

func GapChangeEvent(handler func(gap []int)) {
	preGaps, err := GetGap()
	if err != nil {
		return
	}

	ipc.AddEventListener("configreloaded", func(e string) {
		gaps, err := GetGap()
		if err != nil {
			return
		}

		for i, gap := range gaps {
			if gap != preGaps[i] {
				preGaps = gaps
				handler(gaps)
				return
			}
		}
	}, true)
}

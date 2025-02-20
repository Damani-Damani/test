package auth

import (
	"fmt"
	"log"
	"net/http"
)

var (
	jellyfishURL = "https://api.clearbot.dev"
)

func IsUserAuthenticated(userToken string) bool {
	checkURL := fmt.Sprintf("%s/user/check", jellyfishURL)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	req.Header.Set("Authorization", userToken)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	if response.StatusCode == 200 {
		return true
	}
	return false
}

func CanUserConnectToRobotID(userToken string, robotId int) bool {
	checkURL := fmt.Sprintf("%s/user/check/robot/%d", jellyfishURL, robotId)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", userToken)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if response.StatusCode == 200 {
		return true
	}
	return false
}

func IsRobotAuthenticated(robotToken string) bool {
	checkURL := fmt.Sprintf("%s/robot/check", jellyfishURL)
	response, err := http.Get(checkURL)
	if err != nil {
		return false
	}
	if response.StatusCode == 200 {
		return true
	}
	return false
}

package parse

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

type User struct {
	Login    string
	Password string
}

func CreateBrowser(userData User) *rod.Browser {

	url := "https://pinterest.com/"
	u := launcher.New().Leakless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	page := browser.MustPage(url)
	page.MustWaitDOMStable()
	if len(userData.Login) > 0 {
		if len(page.MustElements("input#email")) > 0 {
			fmt.Println("\nbegin login")
			page.MustWaitDOMStable()
			page.MustElement("input[data-test-id='emailInputField']").MustInput(userData.Login)
			page.MustElement("input[data-test-id='passwordInputField']").MustInput(userData.Password).MustType(input.Enter)
			page.MustWaitDOMStable()
			fmt.Println("finish login")
		} else {
			fmt.Println("\nbegin login")
			page.MustWaitDOMStable()
			page.MustElement("div[data-test-id='login-button']").MustElement("button").MustClick()
			page.MustWaitDOMStable()
			page.MustElement("input#email").MustInput(userData.Login)
			page.MustElement("input#password").MustInput(userData.Password).MustType(input.Enter)
			page.MustWaitDOMStable()
			fmt.Println("finish login")
		}
		page = browser.MustPage(url)
		page.MustWaitDOMStable()
	}
	return browser
}

func LogOut(browser *rod.Browser) {
	url := "https://pinterest.com/"
	page := browser.MustPage(url)
	page.MustWaitDOMStable()
	if len(page.MustElements("button[data-test-id='header-accounts-options-button']")) > 0 {
		page.MustElement("button[data-test-id='header-accounts-options-button']").MustClick()
		page.MustWaitDOMStable()
		page.MustElement("button[data-test-id='header-menu-options-logout']").MustClick()
		page.MustWaitDOMStable()
		fmt.Println("\nLogout Success")
	} else {
		fmt.Println("\nYou already logout")
	}
}

func GetBoardsList(browser *rod.Browser, url string) ([]string, []string) {

	page := browser.MustPage(url)
	page.MustWaitDOMStable()

	var folders = []string{}
	var folderLinks = []string{}
	var loop bool = true
	var loopStop uint8

	boardsList := page.MustElement("div[data-test-id='masonry-container']")
	boardsList.MustScrollIntoView()
	boardSnapCopy := boardsList.MustElements("div[role='listitem']")

	for loop {
		boardSnap := boardsList.MustElements("div[role='listitem']")
		if boardSnap[0].MustText() == boardSnapCopy[0].MustText() && boardSnap[len(boardSnap)-1].MustText() == boardSnapCopy[len(boardSnapCopy)-1].MustText() {
			loopStop++
			if loopStop > 3 {
				loop = false
			}
		}
		for _, v := range boardSnap {
			if len(v.MustElements("h2.wyEmcc")) > 0 {
				element := v.MustElement("h2.wyEmcc")
				if !slices.Contains(folders, element.MustText()) && element.MustText() != "You are signed out" && element.MustText() != "Все пины" {
					fmt.Printf("\nadd board %#v", element.MustText())
					folders = append(folders, element.MustText())
					folderLinks = append(folderLinks, element.MustParents("a.etmDmh").First().MustProperty("href").String())
					loopStop = 0
				}
			}
		}

		boardSnapCopy = boardSnap
		boardSnap.Last().MustScrollIntoView()
		page.MustWaitDOMStable()
	}

	page.Close()

	return folders, folderLinks
}

func unlockScroll(page *rod.Page) {
	block, err := page.Element("div#__PWS_ROOT__")
	if err != nil {
		log.Fatal(err)
	}
	blockTry := block.MustElements("style")
	for _, v := range blockTry {
		err = v.Remove()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetBoard(url string, browser *rod.Browser) (images []string, boardName string) {

	page := browser.MustPage(url)
	page.MustWaitDOMStable()
	loop := true

	boardName = page.MustElement("h1").MustText()
	fmt.Println("\n\nBoard name: ", boardName)

	pinCount := page.MustElement("div[data-test-id='pin-count']").MustText()
	fmt.Println("Board pin amount: ", pinCount)

	unlockScroll(page)

	pinCountInt, err := strconv.Atoi(regexp.MustCompile(`[^0-9]`).ReplaceAllString(pinCount, ""))

	if err != nil {
		fmt.Println(err)
	}

	board := page.MustElement("div[data-test-id='board-feed']")
	board.MustScrollIntoView()
	loopStop := 0
	var skipedElements []string
	boardSnapCopy := board.MustElements("div[role='listitem']")
	if len(boardSnapCopy) < 1 {
		return nil, ""
	}

	for loop {
		boardSnap := board.MustElements("div[role='listitem']")
		if len(boardSnap) > 0 {
			if boardSnap[0].MustText() == boardSnapCopy[0].MustText() && boardSnap[len(boardSnap)-1].MustText() == boardSnapCopy[len(boardSnapCopy)-1].MustText() {
				loopStop++
				if loopStop > 3 {
					loop = false
				}
			}
		}

		if len(boardSnap) < 3 {
			loop = false
		}
		for _, v := range boardSnap {
			if len(v.MustElements("div[data-test-id='pinrep-image']")) > 0 && len(v.MustElements("img[srcset]")) > 0 {
				element := v.MustElement("img[srcset]")
				if !slices.Contains(images, *element.MustAttribute("srcset")) {
					images = append(images, *element.MustAttribute("srcset"))
					fmt.Printf("\nAdd image to %s, board lenght: %d/%d", boardName, len(images), pinCountInt)
					loopStop = 0
				}
			}

			if len(v.MustElements("div[data-test-id='pinrep-video']")) > 0 {
				if !slices.Contains(skipedElements, *v.MustElement("div[data-test-pin-id]").MustAttribute("data-test-pin-id")) {
					skipedElements = append(skipedElements, *v.MustElement("div[data-test-pin-id]").MustAttribute("data-test-pin-id"))
					pinCountInt--
				}
			}

			if len(v.MustElements("div[data-test-id='unavailable-pin']")) > 0 {
				if !slices.Contains(skipedElements, *v.MustElement("div[data-test-pin-id]").MustAttribute("data-test-pin-id")) {
					skipedElements = append(skipedElements, *v.MustElement("div[data-test-pin-id]").MustAttribute("data-test-pin-id"))
					pinCountInt--
				}
			}
		}
		boardSnapCopy = boardSnap
		page.Mouse.MustScroll(0, 300)
		page.MustWaitDOMStable()
	}

	page.Close()
	return images, boardName
}

func GetPin(url string) (string, string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Read body: %v", err)
	}
	reg := regexp.MustCompile(`(<img.*?src=")(?P<link>.*?)(".*?elementtiming="closeupImage".*?>)`)
	match := reg.FindSubmatch(body)
	if len(match) == 0 {
		return "", ""
	}
	nicknameIndex := reg.SubexpIndex("link")
	pinLinkByte := string(match[nicknameIndex])
	pinLinkSlice := strings.Split(pinLinkByte, "/")
	pinName := pinLinkSlice[len(pinLinkSlice)-1]
	pinLinkSlice[3] = "originals"
	pinLink := strings.Join(pinLinkSlice, "/")
	return pinLink, pinName
}

func CheckPinterestLink(url string, pinterest *rod.Browser) bool {

	urlCheck := strings.Split(url, "/")
	if len(urlCheck) < 3 {
		return false
	}
	if urlCheck[1] == "https:" {
		matched, err := regexp.MatchString(`pinterest.com`, urlCheck[2])
		if err != nil {
			fmt.Print(err)
			return false
		}
		if !matched {
			fmt.Println("URL is not Pinterest")
			return matched
		}
	}

	page := pinterest.MustPage(url)
	page.MustWaitDOMStable()
	info, err := page.Info()
	if err != nil {
		log.Fatal(err)
	}

	CurrentPage := strings.Split(info.URL, "/")[3]
	RequiredPage := strings.Split(url, "/")[3]
	if strings.EqualFold(CurrentPage, RequiredPage) {
		return true
	} else {
		return false
	}
}

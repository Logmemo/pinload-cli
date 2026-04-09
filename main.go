package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	// "syscall"

	"pinterestdownload/config"
	"pinterestdownload/db"
	"pinterestdownload/parse"
	"pinterestdownload/printinfo"
)

var version string = "dev"

// For me in Windows 10 cmd does't show colors, this func help with it
// func init() {
// 	stdout := syscall.Handle(os.Stdout.Fd())
//
// 	var originalMode uint32
// 	syscall.GetConsoleMode(stdout, &originalMode)
// 	originalMode |= 0x0004
//
// 	syscall.MustLoadDLL("kernel32").MustFindProc("SetConsoleMode").Call(uintptr(stdout), uintptr(originalMode))
// }

func optionsLoop() {

	var options = map[int]string{
		1: "Download pin",
		2: "Download board",
		3: "Download profile",
		4: "Change download folder",
		5: "Check logout",
	}
	var resetTextColor string = "\033[0m"
	var blueText string = "\033[37;44m"
	var greenText string = "\033[32m"
	var redText string = "\033[31m"

	fmt.Printf("\n%sSelect option:%s", blueText, resetTextColor)
	for i := range len(options) {
		fmt.Printf("\n%s[%d]%s %s", greenText, i+1, resetTextColor, options[i+1])
	}
	fmt.Printf("\n%s[0]%s Exit\n\n", redText, resetTextColor)

	var answer string
	fmt.Scanln(&answer)

	switch answer {
	case "0":
		fmt.Println("Exit...")
		os.Exit(0)
	case "1":
		var pinLink string

		checkSinglePinFolder()

		fmt.Printf("Enter pin URL: ")
		fmt.Scanln(&pinLink)

		parsePin(pinLink)
		optionsLoop()
	case "2":
		var UserData parse.User
		var board string
		var auth string

		fmt.Println("Enter board url:")
		fmt.Scanln(&board)

		fmt.Printf("Need auth?(y/n) ")
		fmt.Scanln(&auth)

		if auth == "y" || auth == "Y" {

			fmt.Printf("Введите логин: ")
			fmt.Scanln(&UserData.Login)
			fmt.Printf("Введите пароль: ")
			fmt.Scanln(&UserData.Password)

			parseBoard(board, UserData)
		} else {
			parseBoard(board, UserData)
		}
	case "3":
		var UserData parse.User
		var user string
		var auth string

		fmt.Printf("Enter username: ")
		fmt.Scanln(&user)

		fmt.Printf("Need auth?(y/n) ")
		fmt.Scanln(&auth)

		if auth == "y" || auth == "Y" {

			fmt.Printf("Введите логин: ")
			fmt.Scanln(&UserData.Login)
			fmt.Printf("Введите пароль: ")
			fmt.Scanln(&UserData.Password)

			parseAll(user, UserData)
		} else {
			parseAll(user, UserData)
		}
	case "4":
		fmt.Printf("Current path: %s\n", config.GetConfigPath())
		fmt.Println("Enter new path: ")
		var newPath string
		fmt.Scanln(&newPath)

		config.ChangeConfigPath(newPath)
		optionsLoop()
	case "5":
		fmt.Println("Check logout...")
		var UserData parse.User
		pinterest := parse.CreateBrowser(UserData)
		parse.LogOut(pinterest)
		pinterest.Close()
		optionsLoop()
	default:
		fmt.Println("Try again")
		optionsLoop()
	}
}

func CreateFolders(user string, list []string) {

	var userPath, groupPath string

	userPath = filepath.Join(config.GetConfigPath(), "Pins", user)

	_, err := os.Stat(userPath)
	if err != nil {
		err3 := os.MkdirAll(userPath, os.ModePerm)
		if err3 != nil {
			fmt.Printf("error: %v", err3)
		}
	}

	for _, v := range list {
		groupPath = filepath.Join(userPath, v)

		_, err := os.Stat(groupPath)
		if err != nil {
			err4 := os.Mkdir(groupPath, os.ModePerm)
			if err4 != nil {
				fmt.Printf("error: %v", err4)
			}
		}
	}
}

func checkSinglePinFolder() {
	SinglePinFolder := filepath.Join(config.GetConfigPath(), "Pins", "_Single Pins")
	_, errStat := os.Stat(SinglePinFolder)
	if errStat != nil {
		errMkdir := os.Mkdir(SinglePinFolder, os.ModePerm)
		if errMkdir != nil {
			log.Fatal(errMkdir)
		}
	}
}

func DownloadFile(ImageUrl string, pathImage string) {

	response, err := http.Get(ImageUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	file, err := os.Create(pathImage)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Image %s download", pathImage)
}

func DownloadFiles(user string, board string, imageList []string) {

	imagePath := filepath.Join(config.GetConfigPath(), "Pins", user)

	for i, imageURL := range imageList {
		imageName := path.Base(imageURL)
		pathImage := filepath.Join(imagePath, board, imageName)
		fmt.Printf("\n%d/%d: ", i, len(imageList))
		DownloadFile(imageURL, pathImage)
	}
}

func ExtractURL(imageList []string) (URLs []string) {
	for _, v := range imageList {
		URLs = append(URLs, regexp.MustCompile(`https://[^\s]*originals[^\s]*`).FindString(v))
	}
	return URLs
}

func getNicknameFromURL(URL string) string {
	reg := regexp.MustCompile(`\.com/(?P<nickname>.*?)/`)
	match := reg.FindStringSubmatch(URL)
	nicknameIndex := reg.SubexpIndex("nickname")
	user := match[nicknameIndex]
	return user
}

func checkValidLink(url string) bool {

	urlCheck := strings.Split(url, "/")
	if len(urlCheck) < 3 {
		fmt.Println("Invalid URL 1")
		return false
	}
	if urlCheck[0] == "https:" {
		matched, err := regexp.MatchString(`pinterest.com`, urlCheck[2])
		if err != nil {
			fmt.Println("ERROR: ", err)
			return false
		}
		if matched {
			return matched
		} else {
			fmt.Println("Invalid URL 2")
			return matched
		}
	} else {
		fmt.Println("Invalid URL 3")
		return false
	}
}

func parsePin(pinLink string) {
	pinLink, imageName := parse.GetPin(pinLink)
	if pinLink == "" {
		fmt.Println("Pin not found in pinterest")
		optionsLoop()
	}
	imagePath := filepath.Join(config.GetConfigPath(), "Pins", "_Single Pins", imageName)
	DownloadFile(pinLink, imagePath)
	fmt.Println("Pin download success")
}

func parseBoard(boardUrl string, UserData parse.User) {

	if !checkValidLink(boardUrl) {
		optionsLoop()
	}

	pinterest := parse.CreateBrowser(UserData)
	if !parse.CheckPinterestLink(boardUrl, pinterest) {
		fmt.Println("Board not found in pinterest")
		pinterest.Close()
		optionsLoop()
	}
	BoardImages, boardName := parse.GetBoard(boardUrl, pinterest)
	BoardImages = ExtractURL(BoardImages)

	user := getNicknameFromURL(boardUrl)

	CreateFolders(user, []string{boardName})

	userID := db.CheckUserDB(user)
	db.DBAddBoards(userID, []string{boardName}, []string{boardUrl})
	db.DBAddPins(userID, boardName, BoardImages)

	printinfo.PinsOnBoard(BoardImages)

	DownloadFiles(user, boardName, BoardImages)
	if len(UserData.Login) > 0 {
		parse.LogOut(pinterest)
	}
	pinterest.Close()
	fmt.Println("\nDownload Success")
	optionsLoop()
}

func parseAll(user string, UserData parse.User) {
	url := "https://pinterest.com/" + user + "/"
	pinterest := parse.CreateBrowser(UserData)
	if !parse.CheckPinterestLink(url, pinterest) {
		fmt.Println("user not found in pinterest")
		pinterest.Close()
		optionsLoop()
	}
	boardNames, boardLinks := parse.GetBoardsList(pinterest, url)
	CreateFolders(user, boardNames)
	userID := db.CheckUserDB(user)
	db.DBAddBoards(userID, boardNames, boardLinks)
	for i, v := range boardLinks {
		BoardImages, _ := parse.GetBoard(v, pinterest)
		BoardImages = ExtractURL(BoardImages)
		db.DBAddPins(userID, boardNames[i], BoardImages)
		printinfo.PinsOnBoard(BoardImages)
		DownloadFiles(user, boardNames[i], BoardImages)
	}
	if len(UserData.Login) > 0 {
		parse.LogOut(pinterest)
	}
	pinterest.Close()
	fmt.Println("Download Success")
	optionsLoop()
}

func main() {

	printinfo.HelloMessage(version)
	config.CheckConfig()
	optionsLoop()

}

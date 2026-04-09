package printinfo

import (
	"fmt"
)

func HelloMessage(version string) {
	var authorInfo = map[string]string{
		"author": "Logmemo",
		"github": "https://github.com/pinterest-download",
		"email":  "logmemogithub@gmail.com",
	}

	fmt.Printf(`
	██╗      ██████╗ ██╗███╗   ██╗██╗      ██████╗  █████╗ ██████╗ 
	╚██╗     ██╔══██╗██║████╗  ██║██║     ██╔═══██╗██╔══██╗██╔══██╗
	 ╚██╗    ██████╔╝██║██╔██╗ ██║██║     ██║   ██║███████║██║  ██║
	 ██╔╝    ██╔═══╝ ██║██║╚██╗██║██║     ██║   ██║██╔══██║██║  ██║
	██╔╝     ██║     ██║██║ ╚████║███████╗╚██████╔╝██║  ██║██████╔╝
	╚═╝      ╚═╝     ╚═╝╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ 
                                                      
	Version: %s
	Author:  %s
	Github:  %s
	Email:   %s
	`,
		version,
		authorInfo["author"],
		authorInfo["github"],
		authorInfo["email"],
	)
}

func PinsOnBoard(pins []string) {
	fmt.Printf("\nFound %d pins from board:", len(pins))
	for i, v := range pins {
		fmt.Printf("\n%d. %s", i, v)
	}
}

func PrintBoards(boardNames []string, boardLinks []string) {
	fmt.Printf("Found %d board(s):", len(boardNames))
	for i, v := range boardNames {
		fmt.Printf("\n%d. %s (Board URL: %s )", i, v, boardLinks[i])
	}
}

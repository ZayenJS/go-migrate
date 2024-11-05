package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
)

func GetCurrentWorkingDirectory() string {
	currentWorkingDirectory, err := os.Getwd()

	if err != nil {
		fmt.Println("Error getting current working directory")
		os.Exit(1)
	}

	return currentWorkingDirectory
}

func CreateDirectoryIfNotExist(directoryPath string) {
	_, err := os.Stat(directoryPath)

	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(directoryPath, 0777)

		if err != nil {
			fmt.Println("Error creating migrations directory")
			os.Exit(1)
		}
	}
}

func CreateFileIfNotExist(filePath string, content string, mode int) {
	_, err := os.Stat(filePath)

	if err != nil && os.IsNotExist(err) {
		file, err := os.Create(filePath)

		if err != nil {
			fmt.Printf("Error creating file %v\n", filePath)
			os.Exit(1)
		}

		file.Chmod(0666)
		defer file.Close()

		_, err = file.WriteString(content)

		if err != nil {
			fmt.Printf("Error writing to file %v\n", filePath)
			os.Exit(1)
		}

		return
	}

	if err != nil {
		fmt.Printf("Error checking file %v\n", filePath)
		os.Exit(1)
	}

	file, err := os.OpenFile(filePath, mode, 0644)

	if err != nil {
		fmt.Printf("Error opening file %v\n", filePath)
		os.Exit(1)
	}

	defer file.Close()

	envFileContent := GetFileContent(filePath)

	match, _ := regexp.MatchString(`GO_MIGRATE_DATABASE_URL=(postgres|mysql)://.*`, envFileContent)

	if !match {
		_, err := file.WriteString(content)

		if err != nil {
			fmt.Printf("Error writing to file %v\n", filePath)
			os.Exit(1)
		}
	}

}

func GetFileContent(filePath string) string {
	fileContent, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Printf("Error reading file %v, please make sure it exists\n", filePath)
		os.Exit(1)
	}

	return string(fileContent)
}

func GetDirectoryContent(directoryPath string) []fs.DirEntry {
	content, err := os.ReadDir(directoryPath)

	if err != nil {
		fmt.Printf("Error reading directory %v, please make sure it exists\n", directoryPath)
		os.Exit(1)
	}

	return content
}

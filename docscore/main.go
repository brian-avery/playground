package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var maxScore = 15

type testStatus string

var (
	testNotApplicable testStatus = "n/a"
	testNo            testStatus = "no"
	testYes           testStatus = "yes"
	testUnknown       testStatus = "unknown"
)

var excludeDirs = []string{
	"en/about",
	"zh/about",
	"en/blog",
	"zh/blog",
	"en/boilerplates",
	"zh/boilerplates",
	"en/events",
	"zh/events",
	"en/docs/reference/glossary",
	"zh/docs/reference/glossary",
	"zh/news",
	"en/news",
	"en/test",
	"zh/test",
}

type fileEntry struct {
	fullPath string
	relative string
	url      string
	tested   testStatus
	score    int
	title    string
	owner    string
	notes    []string
}

func main() {
	var docsPath, outPath, analyticsPath string
	flag.StringVar(&docsPath, "docspath", "../istio.io", "points to the cloned path of the istio.io site")
	flag.StringVar(&outPath, "outpath", "out.csv", "path to create the spreadsheet CSV at")
	flag.StringVar(&analyticsPath, "analyticspath", "analytics.csv", "path to a file containing the istio.io analytics CSV")
	flag.Parse()
	fmt.Printf("Scoring docs in %s\n", docsPath)

	files, err := getAllFiles(docsPath)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	var scorers = []scorer{
		getHitsScorer(analyticsPath, 15),
	}
	for _, scorerInstance := range scorers {
		files = scorerInstance.Score(files)
	}

	writeTestSpreadsheet(files, outPath)
}

func getPriority(file fileEntry) string {
	priority := 0
	switch score := file.score; {
	case score > maxScore*2/3:
		priority = 0
	case score > maxScore*1/3:
		priority = 1
	case score > 0:
		priority = 2
	case score == 0:
		priority = 3

	}
	return fmt.Sprintf("P%d", priority)
}

func writeTestSpreadsheet(files []fileEntry, path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("unable to create CSV: %s", err.Error())
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"",
		"Owner",
		"Test Cases",
		"Priority",
		"Automated",
		"In Progress",
		"In Progress Last Updated",
		"Done By",
		"Done By Last Updated",
		"GitHub Issue",
		"Comments (e.g. env used)",
		"Automated Sign Up",
		"Automated Last Updated",
		"Automation GitHub Issue",
		"Generator Notes"})

	for _, value := range files {
		err := writer.Write([]string{
			"",
			value.owner,
			fmt.Sprintf("=HYPERLINK(\"%s\",\"%s\")", value.url, strings.ReplaceAll(value.title, "\"", "")),
			getPriority(value),
			string(value.tested),
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			strings.Join(value.notes[:], "\n"),
		})

		if err != nil {
			log.Fatalf("unable to write CSV: %s", err.Error())
		}
	}
	return nil
}

func getURL(path string) string {
	path = strings.TrimPrefix(path, "en/")
	return fmt.Sprintf("https://preliminary.istio.io/latest/%s", filepath.Dir(path))
}
func getAllFiles(path string) ([]fileEntry, error) {
	files := make([]fileEntry, 0)
	rootPath := path
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		relativePath := strings.TrimPrefix(path[len(rootPath):], "/")
		if filepath.Ext(path) == ".md" && !skipFile(relativePath) {
			fileContent, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Unable to open file %s\n", path)
			}

			files = append(files, fileEntry{
				fullPath: path,
				relative: relativePath,
				url:      getURL(relativePath),
				tested:   isTested(string(fileContent)),
				title:    getAndTrimField("title", string(fileContent)),
				owner:    getAndTrimField("owner", string(fileContent)),
				notes:    []string{fmt.Sprintf("Relative path:%s", relativePath)},
			})
		}
		return nil
	})
	return files, err
}

func skipFile(path string) bool {
	for _, value := range excludeDirs {
		if strings.HasPrefix(path, value) {
			return true
		}
	}
	return false
}

func getAndTrimField(field string, body string) string {
	fieldRegex := regexp.MustCompile(fmt.Sprintf("%s:\\s*[\\w|\\/]*.*", field))
	fieldValue := strings.TrimPrefix(fieldRegex.FindString(string(body)), fmt.Sprintf("%s:", field))
	fieldValue = strings.TrimPrefix(fieldValue, " ")
	return fieldValue
}

func isTested(content string) testStatus {
	teststatus := getAndTrimField("test", content)
	if teststatus == "" {
		teststatus = string(testUnknown)
	}
	return testStatus(teststatus)
}
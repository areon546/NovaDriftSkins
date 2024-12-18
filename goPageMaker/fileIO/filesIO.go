package fileIO

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/areon546/NovaDriftCustomSkins/goPageMaker/helpers"
)

// ~~~~~~~~~~~~~~~~ File

type File struct {
	filename      string
	suffix        string
	relPath       string
	contentBuffer []string
	lines         int
	hasBeenRead   bool
}

func NewFileWithSuffix(fn string, suff string, path string) *File {
	return &File{filename: fn, suffix: suff, relPath: path, hasBeenRead: false}
}

func NewFile(fn string) *File {
	fn, suff := splitFileName(fn)
	return &File{filename: fn, suffix: suff, hasBeenRead: false}
}

func splitFileName(filename string) (name, suffix string) {
	stringSections := strings.Split(filename, ".")
	// print(stringSections)

	if len(stringSections) > 1 {
		suffix = stringSections[len(stringSections)-1]
	}

	for i := 0; i < len(stringSections)-1; i++ {
		name += stringSections[i]
	}

	return
}

func (f *File) GetFileName() string {
	return helpers.Format("%s.%s", f.filename, f.suffix)
}

func (f *File) GetContents() []string {
	return f.contentBuffer
}

func (f *File) readFile() []string {
	if !f.hasBeenRead {
		data, err := os.ReadFile(f.GetFileName()) // For read access.
		handle(err)

		oneLine := strings.ReplaceAll(string(data), "\r", "")
		f.contentBuffer = strings.Split(oneLine, "\n")
		f.lines = len(f.contentBuffer)
	}
	return f.contentBuffer
}

func (f *File) isEmpty() bool {
	return len(f.contentBuffer) == 0
}

func (f *File) readLine(lineNum int) (output string, err error) {
	lineNum -= 1 // converted to index notation

	if f.isEmpty() {
		f.readFile()
	}

	if lineNum > f.lines {
		return "", errors.New("Index out of bounds for File length")
	}

	output = f.contentBuffer[lineNum]
	print(output)

	return
}

func (f *File) WriteFile() {
	if err := os.WriteFile(f.GetFileName(), []byte(f.bufferToString()), 0666); err != nil {
		log.Fatal(err)
	}
}

func (f *File) appendLines(arr []string) {

	// f.contentBuffer = append(f.contentBuffer, arr...)
	for _, v := range arr {
		f.contentBuffer = append(f.contentBuffer, v)
	}
}

func (f *File) Append(s string) {
	f.AppendLine(s, len(f.contentBuffer), true)
}

func (f *File) AppendNewLine() {
	f.Append("")
}

func (f *File) bufferLines(arr []string) {

	if f.isEmpty() {
		f.contentBuffer = make([]string, len(arr))
	}

	f.contentBuffer = append(f.contentBuffer, arr...)

}

func (f *File) clearFile() {
	if err := os.WriteFile(f.GetFileName(), make([]byte, 0), 0666); err != nil {
		log.Fatal(err)
	}
}

func (f *File) String() string {
	return f.GetFileName()
}

func (f *File) bufferToString() string {
	s := ""
	for _, v := range f.contentBuffer {
		s += v
	}

	return s
}

func (f *File) AppendLine(s string, i int, nl bool) {

	for i >= len(f.contentBuffer) {
		f.contentBuffer = append(f.contentBuffer, "")
	}

	if nl {
		s += "\n"
	}

	f.contentBuffer[i] = s
}

func ConstructPath(preffix, directory, fileName string) (s string) {
	if !reflect.DeepEqual(preffix, "") {
		s += preffix + "/"
	}
	s += directory + "/" + fileName
	return s
}

// ~~~~~~~~~~~~~~~~~~~~ MarkdownFile
type MarkdownFile struct {
	File
}

func NewMarkdownFile(name, path string) *MarkdownFile {
	return &MarkdownFile{File: *NewFileWithSuffix(name, "md", path)}
}

func (m *MarkdownFile) AppendMarkdownLink(displayText, path string) {
	m.Append(ConstructMarkDownLink(false, displayText, path))
}

func (m *MarkdownFile) AppendMarkdownEmbed(path string) {
	m.Append(ConstructMarkDownLink(true, "", path))
}

func ConstructMarkDownLink(embed bool, displayText, path string) (s string) {
	if embed {
		s += "!"
	}
	s += fmt.Sprintf("[%s](%s)", displayText, path)
	return
}

// ~~~~~~~~~~~~~~~~~~~~ CSVFile

type CSVFile struct {
	File
	headings []string
	contents [][]string
}

// returns an array of headings and a 2d array of
func ReadCSV(fileName string) (csv CSVFile) {
	file := File{filename: fileName, suffix: "csv"}
	// read fileName into CSVFile

	// file := makeFile(fileName)
	fileContents := file.readFile()

	// go through each line in CSV and
	for i, csvCell := range fileContents {
		// print("csv:", csvCell)
		if i == 0 { // adds headings to headings attribute
			csv.headings = strings.Split(csvCell, ",")
		} else { // ads csv items to contents attribute

			// check if the string is empty, if so skip
			if reflect.DeepEqual(csvCell, "") {
				continue
			}

			csv.contents = append(csv.contents, strings.Split(csvCell, ","))
		}
	}

	return
}

func (c *CSVFile) GetIndexOfColumn(header string) (index int) {
	for i, heading := range c.headings {
		if reflect.DeepEqual(heading, header) {
			index = i
		}
	}

	return
}

func (c *CSVFile) NumHeaders() int {
	return len(c.headings)
}

func (c *CSVFile) Rows() int {
	return len(c.contents)
}

func (c *CSVFile) GetContents() [][]string {
	return c.contents
}

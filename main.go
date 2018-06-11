package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type byFileName []os.FileInfo

func (nf byFileName) Len() int           { return len(nf) }
func (nf byFileName) Swap(i, j int)      { nf[i], nf[j] = nf[j], nf[i] }
func (nf byFileName) Less(i, j int) bool { return nf[i].Name() < nf[j].Name() }

func filterFiles(infos []os.FileInfo) []os.FileInfo {
	filtered := make([]os.FileInfo, 0)
	for _, info := range infos {
		if info.IsDir() {
			filtered = append(filtered, info)
		}
	}
	return filtered
}

func filePrefix(isLast bool) string {
	if isLast {
		return "└"
	}
	return "├"
}

func nextPrefix(currentPrefix string, isLast bool) string {
	nextPrefix := currentPrefix

	if !isLast {
		nextPrefix += "│"
	}

	return nextPrefix + "\t"
}

func fileSize(info os.FileInfo) string {
	if info.IsDir() {
		return ""
	}
	var sizeText string
	if size := info.Size(); size == 0 {
		sizeText = "empty"
	} else {
		sizeText = fmt.Sprintf("%db", size)
	}
	return fmt.Sprintf(" (%s)", sizeText)
}

func printInfo(out io.Writer, info os.FileInfo, prefix string, isLast bool) {
	format := prefix + filePrefix(isLast) + "───%s" + fileSize(info) + "\n"
	fmt.Fprintf(out, format, info.Name())
}

func readDir(out io.Writer, path string, printFiles bool, prefix string) error {
	dir, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("Can't open path %v: %v", path, err)
	}

	if stat, err := dir.Stat(); err != nil || !stat.IsDir() {
		return fmt.Errorf("Path %v is not directory", path)
	}

	infos, err := dir.Readdir(-1)

	if err != nil {
		return fmt.Errorf("Error while reading directories: %v", err)
	}

	if !printFiles {
		infos = filterFiles(infos)
	}

	sort.Sort(byFileName(infos))

	for i, info := range infos {
		isLast := i == len(infos)-1
		printInfo(out, info, prefix, isLast)

		if !info.IsDir() {
			continue
		}

		nextPath := filepath.Join(path, info.Name())
		err := readDir(out, nextPath, printFiles, nextPrefix(prefix, isLast))
		if err != nil {
			return err
		}
	}

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return readDir(out, path, printFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

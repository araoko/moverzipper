package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var dFlag = flag.String("d", "", "Destination folder for zipped files")
	var sFlag = flag.String("s", "", "folder containing files to zip")
	var tFlag = flag.String("t", "", "extension of files to zip")
	var mFlag = flag.Int("m", 0, "least modification date from now in days")
	flag.Parse()

	if err := validateFlags(*dFlag, *sFlag, *tFlag, *mFlag); err != nil {
		log.Fatalln(err)
	}
	maxModDate := time.Now().AddDate(0, 0, -*mFlag)
	files, err := ioutil.ReadDir(*sFlag)
	if err != nil {
		log.Fatal(err)
	}
	r := make([]string, 0)
	for _, file := range files {
		if file.IsDir() || !strings.EqualFold(filepath.Ext(file.Name()), "."+*tFlag) || file.ModTime().After(maxModDate) {
			continue
		}
		r = append(r, filepath.Join(*sFlag, file.Name()))
	}

	err = ZipAndDelete(*dFlag, r)
	fmt.Println(err)
	fmt.Println("Done")

}

// ZipAndDelete compresses list of files individually into a folder,
//appending .zip to the original names and deleting the original files
// modified from  ZipFiles function on https://golangcode.com/create-zip-files-in-go/
func ZipAndDelete(zipFolder string, files []string) error {
	var fileInfo os.FileInfo
	var err error
	if fileInfo, err = os.Stat(zipFolder); err == nil && !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a folder", zipFolder)
	}
	if err != nil {
		return err
	}

	// Add files to zip
	for _, file := range files {
		s := fmt.Sprintf("\rFile: %s -  ", filepath.Base(file))
		fmt.Print(s)
		zippedFileName := filepath.Join(zipFolder, filepath.Base(file)+".zip")
		zippedFile, err := os.Create(zippedFileName)
		if err != nil {
			//s := fmt.Sprintf("\x1B\x5B%sDFailed")
			fmt.Println(s, "Error Creating zip file ")
			return err
		}

		zipWriter := zip.NewWriter(zippedFile)
		fmt.Print(s, "Opening..   ")
		file2Zip, err := os.Open(file)
		if err != nil {
			fmt.Println(s, "Failed      ")
			return err
		}
		s = fmt.Sprint(s, "Opened   ")
		fmt.Print(s)
		defer file2Zip.Close()

		// Get the file information
		info, err := file2Zip.Stat()
		if err != nil {
			fmt.Println(s, "Failed      (error getting file info)")
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			fmt.Println(s, "Failed      (error getting header)")
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			fmt.Println(s, "Failed      (error creating header)")
			return err
		}
		fmt.Print(s, "Zipping..   ")
		_, err = io.Copy(writer, file2Zip)
		if err != nil {
			fmt.Println(s, "Failed      ")
			return err
		}
		s = fmt.Sprint(s, "Zipped   ")
		fmt.Print(s)
		zipWriter.Close()
		zippedFile.Close()
		file2Zip.Close()
		fmt.Print(s, "Deleting..  ")
		err = os.Remove(file)

		if err != nil {
			fmt.Println(s, "Failed      ")
			return err
		}
		fmt.Println(s, "Deleted     ")
	}
	return nil
}

func validateFlags(dFlag, sFlag, tFlag string, mFlag int) error {
	if dFlag == "" || sFlag == "" || tFlag == "" {
		return fmt.Errorf("all arguments except -m are mandatory. use -h to get argument list")
	}
	if mFlag < 0 {
		return fmt.Errorf("-d value must be positive")
	}

	if !isFolderExisting(dFlag) {
		return fmt.Errorf("cannot find destination folder %s", dFlag)
	}

	if !isFolderExisting(sFlag) {
		return fmt.Errorf("cannot find source folder %s", sFlag)
	}

	return nil
}

func isFolderExisting(s string) bool {
	if f, err := os.Stat(s); err == nil {
		if f.IsDir() {
			return true
		}
		log.Println(s, "is not a directory")
		return false
	}
	log.Println("error running os.Stat() on", s)
	return false
}

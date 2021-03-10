package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	var file, output string

	fmt.Println()
	//if a tar.gz file and extraction path defined
	if len(os.Args) == 3 {
		file, output = os.Args[1], os.Args[2]
		//if only tar.gz file specified
	} else if len(os.Args) == 2 {
		file, output = os.Args[1], strings.TrimSuffix(os.Args[1], ".tar.gz")
		//if no arguments specified get mad
	} else if len(os.Args) == 1 {
		fmt.Println("No tar.gz archive specified!")
		os.Exit(1)
	}

	err := Untar(file, output)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Untarred " + file + " to: " + output + " successfully")

}

// Untar will decompress a zip archive, moving all files and folders
// within the tar file (parameter 1) to an output directory (parameter 2).
func Untar(src string, dest string) error {

	//f is the main konvoy bundle
	f, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//gzf is the gzip reader for the main bundle
	gzf, err := gzip.NewReader(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tarReader := tar.NewReader(gzf)

	//destination folder is the name of the bundle minus tar.gz
	err = os.Mkdir(dest, 0755)

	//Main loop, extract the various top level tar.gz files, helm, cluster-data and master/worker
	for true {

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if header.Typeflag == tar.TypeReg {
			//delete after creating this target
			target := filepath.Join(dest, header.Name)
			targetFolder := filepath.Join(dest, strings.TrimSuffix(header.Name, ".tar.gz"))
			err = os.MkdirAll(targetFolder, 0755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tarReader); err != nil {
				return err
			}
			f.Close()
			//now take the file.tar.gz from the main bundle and extract it into its own folder created above
			ftarget, err := os.Open(target)
			fileGZ, err := gzip.NewReader(ftarget)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fileGZReader := tar.NewReader(fileGZ)

			for true {
				fileheader, err := fileGZReader.Next()
				if err == io.EOF {
					break
				}
				if header.Typeflag == tar.TypeReg {

					targetFile := filepath.Join(targetFolder, fileheader.Name)
					fmt.Println(targetFile)

					cut := -1
					if runtime.GOOS == "windows" {
						cut = strings.LastIndex(targetFile, "\\")

					} else {

						cut = strings.LastIndex(targetFile, "/")
					}

					if cut == -1 {
						fmt.Println("Error: no path separator in filepath for: " + targetFile)
					} else {
						err = os.MkdirAll(targetFile[0:cut], 0755)

						if _, err := os.Stat("targetFile"); os.IsNotExist(err) {
							// path/to/whatever does not exist

							file, err := os.OpenFile(targetFile, os.O_CREATE|os.O_RDWR, os.FileMode(fileheader.Mode))
							if err != nil {
								return err
							}

							// copy over contents
							if _, err := io.Copy(file, fileGZReader); err != nil {
								return err
							}
						}
					}
				}

			}
			ftarget.Close()
			e := os.Remove(target)
			if e != nil {
				log.Fatal(e)
			}
		}

	}
	f.Close()
	return err
}

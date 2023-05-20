package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var srcDir string
var destDir string
var numWorkers int

type Job struct {
	srcFile  string
	destFile string
	isFlac   bool
}

func init() {
	flag.StringVar(&srcDir, "src", "", "Source directory containing the music files.")
	flag.StringVar(&destDir, "dest", "", "Destination directory where the converted files will be saved.")
	flag.IntVar(&numWorkers, "workers", 5, "Number of workers for processing files.")
}

func copyFile(srcFile, destFile string) (err error) {
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("unable to open source file: %w", err)
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing source file: %w", closeErr)
		}
	}()

	dest, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("unable to create destination file: %w", err)
	}
	defer func() {
		if closeErr := dest.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing destination file: %w", closeErr)
		}
	}()

	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("unable to copy source to destination: %w", err)
	}

	return nil
}

func convertFile(srcFile, destFile string) error {
	cmd := exec.Command("qaac64.exe", "--cvbr", "192", "--ignorelength", "--copy-artwork", srcFile, "-o", destFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	return nil
}

func worker(jobs <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		if _, err := os.Stat(job.destFile); err == nil {
			fmt.Printf("Skipping (file already exists): %s\n", job.destFile)
			continue
		} else if !os.IsNotExist(err) {
			fmt.Printf("Error checking file: %s\n", err)
			continue
		}

		if job.isFlac {
			fmt.Printf("Converting: %s\n", job.srcFile)
			if err := convertFile(job.srcFile, job.destFile); err != nil {
				fmt.Printf("Error converting file: %s\n", err)
			}
		} else {
			fmt.Printf("Copying: %s\n", job.srcFile)
			if err := copyFile(job.srcFile, job.destFile); err != nil {
				fmt.Printf("Error copying file: %s\n", err)
			}
		}
	}
}

func main() {
	flag.Parse()

	if srcDir == "" || destDir == "" {
		fmt.Println("Both source and destination directories should be specified.")
		os.Exit(1)
	}

	fmt.Printf("Source directory: %s\n", srcDir)
	fmt.Printf("Destination directory: %s\n", destDir)
	fmt.Printf("Number of workers: %d\n", numWorkers)

	jobs := make(chan Job)
	var wg sync.WaitGroup

	// Start the workers.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error visiting file: %w", err)
		}

		if f.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("unable to determine relative path: %w", err)
		}

		destFile := filepath.Join(destDir, rel)
		destDirPath := filepath.Dir(destFile)
		if err := os.MkdirAll(destDirPath, 0755); err != nil {
			return fmt.Errorf("unable to create destination directory: %w", err)
		}

		if strings.ToLower(filepath.Ext(path)) == ".flac" {
			destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".m4a"
			jobs <- Job{srcFile: path, destFile: destFile, isFlac: true}
		} else if strings.ToLower(filepath.Ext(path)) == ".mp3" {
			jobs <- Job{srcFile: path, destFile: destFile, isFlac: false}
		}

		return nil
	})

	close(jobs)

	// Wait for all jobs to finish.
	wg.Wait()

	if err != nil {
		fmt.Printf("Error processing files: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Conversion complete.")
	}
}

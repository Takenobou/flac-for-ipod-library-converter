package main

import (
	"flag"
	"fmt"
	"github.com/disintegration/imaging"
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
var codec string
var bitrate int

type Job struct {
	srcFile  string
	destFile string
	isFlac   bool
	bitrate  int
}

func init() {
	flag.StringVar(&srcDir, "src", "", "Source directory containing the music files.")
	flag.StringVar(&destDir, "dest", "", "Destination directory where the converted files will be saved.")
	flag.IntVar(&numWorkers, "workers", 5, "Number of workers for processing files.")
	flag.StringVar(&codec, "codec", "aac", "Codec to use for conversion. Options are 'aac' and 'opus'.")
	flag.IntVar(&bitrate, "bitrate", 192, "Bitrate for the output file.")
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

func convertFile(srcFile, destFile string, codec string, bitrate int) error {
	var cmd *exec.Cmd
	if codec == "aac" {
		cmd = exec.Command("qaac64.exe", "--cvbr", fmt.Sprintf("%d", bitrate), "--ignorelength", "--copy-artwork", srcFile, "-o", destFile)
	} else if codec == "opus" {
		cmd = exec.Command("opusenc", "--bitrate", fmt.Sprintf("%d", bitrate), srcFile, destFile)
	} else {
		return fmt.Errorf("unsupported codec: %s", codec)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	return nil
}

func resizeAndConvertImage(srcFile, destFile string) error {
	if _, err := os.Stat(destFile); err == nil {
		fmt.Printf("Skipping (file already exists): %s\n", destFile)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking file: %w", err)
	}

	src, err := imaging.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	src = imaging.Resize(src, 320, 320, imaging.Lanczos)

	err = imaging.Save(src, destFile, imaging.JPEGQuality(90))
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Printf("Cover saved successfully: %s\n", destFile)
	return nil
}

func worker(jobs <-chan Job, wg *sync.WaitGroup, codec string) {
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
			if err := convertFile(job.srcFile, job.destFile, codec, job.bitrate); err != nil {
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

	if codec == "opus" {
		bitrate = 160
	}

	jobs := make(chan Job)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg, codec)
	}

	err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error visiting file: %w", err)
		}

		if f.IsDir() {
			if cover, err := os.Stat(filepath.Join(path, "cover.png")); err == nil && !cover.IsDir() {
				relDir, err := filepath.Rel(srcDir, path)
				if err != nil {
					return fmt.Errorf("unable to determine relative directory path: %w", err)
				}
				destDir := filepath.Join(destDir, relDir)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					return fmt.Errorf("unable to create destination directory: %w", err)
				}
				coverDestFile := filepath.Join(destDir, "cover.jpg")
				if err := resizeAndConvertImage(filepath.Join(path, "cover.png"), coverDestFile); err != nil {
					fmt.Printf("Error resizing and converting image: %s\n", err)
				}
			}
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("unable to determine relative path: %w", err)
		}

		destFile := filepath.Join(destDir, rel)
		if strings.ToLower(filepath.Ext(path)) == ".flac" {
			if codec == "aac" {
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".m4a"
			} else if codec == "opus" {
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".opus"
			}
			jobs <- Job{srcFile: path, destFile: destFile, isFlac: true, bitrate: bitrate}
		} else if strings.ToLower(filepath.Ext(path)) == ".mp3" {
			jobs <- Job{srcFile: path, destFile: destFile, isFlac: false, bitrate: bitrate}
		}

		return nil
	})

	close(jobs)

	wg.Wait()

	if err != nil {
		fmt.Printf("Error processing files: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Conversion complete.")
	}
}

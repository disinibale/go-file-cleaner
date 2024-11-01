package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	var dirPath string
	if len(os.Args) < 2 {
		fmt.Print("Please enter a directory path to scan: ")
		fmt.Scanln(&dirPath)
	} else {
		dirPath = os.Args[1]
	}

	if dirPath == "" {
		fmt.Println("No directory path provided. Exiting.")
		time.Sleep(3 * time.Second)
		return
	}

	sizeMap := make(map[int64][]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		// Filter only .mp4 files
		if !info.IsDir() && strings.ToLower(filepath.Ext(info.Name())) == ".mp4" {
			size := info.Size()
			sizeMap[size] = append(sizeMap[size], path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", dirPath, err)
		time.Sleep(3 * time.Second)
		return
	}

	fmt.Println("MP4 files with the same size:")
	printDuplicatesInTable(sizeMap)

	totalSavedSpace := calculatePotentialSpaceSaved(sizeMap)
	fmt.Printf("\nTotal space that can be saved by deleting duplicates: %d bytes\n", totalSavedSpace)

	var choice string
	fmt.Print("\nDo you want to delete duplicate MP4 files with the same size? (y/n): ")
	fmt.Scanln(&choice)

	if choice == "y" || choice == "Y" {
		deleteDuplicates(sizeMap)
	}

	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}

func printDuplicatesInTable(sizeMap map[int64][]string) {
	// Sort the sizes in ascending order
	sizes := make([]int64, 0, len(sizeMap))
	for size := range sizeMap {
		sizes = append(sizes, size)
	}
	sort.Slice(sizes, func(i, j int) bool { return sizes[i] < sizes[j] })

	// Print header
	fmt.Printf("%-15s %-50s\n", "File Size (bytes)", "File Paths")
	fmt.Println(strings.Repeat("-", 65))

	// Print sorted file sizes and their paths
	for _, size := range sizes {
		files := sizeMap[size]
		if len(files) > 1 {
			for _, file := range files {
				fmt.Printf("%-15d %-50s\n", size, file)
			}
			fmt.Println(strings.Repeat("-", 65))
		}
	}
}

func calculatePotentialSpaceSaved(sizeMap map[int64][]string) int64 {
	var totalSavedSpace int64

	for _, files := range sizeMap {
		if len(files) > 1 {
			totalSavedSpace += int64(len(files)-1) * getFileSize(files[0])
		}
	}

	return totalSavedSpace
}

func deleteDuplicates(sizeMap map[int64][]string) {
	for size, files := range sizeMap {
		if len(files) > 1 {
			fmt.Printf("\nProcessing files of size %d bytes:\n", size)
			hashMap := make(map[string][]string)

			for _, file := range files {
				hash, err := getFileHash(file)
				if err != nil {
					fmt.Printf("Error hashing file %s: %v\n", file, err)
					continue
				}
				hashMap[hash] = append(hashMap[hash], file)
			}

			for hash, duplicates := range hashMap {
				if len(duplicates) > 1 {
					fmt.Printf("Duplicate files with hash %s:\n", hash)
					for i, file := range duplicates {
						if i == 0 {
							fmt.Printf("  Keeping: %s\n", file)
						} else {
							var deleteChoice string
							fmt.Printf("  Do you want to delete: %s? (y/n): ", file)
							fmt.Scanln(&deleteChoice)
							if deleteChoice == "y" || deleteChoice == "Y" {
								err := os.Remove(file)
								if err != nil {
									fmt.Printf("Error deleting file %s: %v\n", file, err)
								} else {
									fmt.Printf("  Deleted: %s\n", file)
								}
							} else {
								fmt.Printf("  Skipped: %s\n", file)
							}
						}
					}
				}
			}
		}
	}
}

func getFileSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

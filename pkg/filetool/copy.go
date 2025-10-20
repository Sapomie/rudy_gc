package filetool

import (
	"fmt"
	"io"
	"os"
)

// CopyFileWithProgress copies a file from src to dst and displays the progress.
func CopyFileWithProgress(src, dst string) error {
	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %w", err)
	}
	defer sourceFile.Close()

	// 获取源文件的大小
	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("无法获取源文件信息: %w", err)
	}
	totalSize := sourceFileInfo.Size()

	// 创建目标文件
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("无法创建目标文件: %w", err)
	}
	defer destinationFile.Close()

	buffer := make([]byte, 4096) // 缓冲区4KB
	var totalCopied int64

	for {
		// 从源文件读取数据到缓冲区
		n, readErr := sourceFile.Read(buffer)
		if n > 0 {
			// 将缓冲区的数据写入到目标文件
			_, writeErr := destinationFile.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("写入目标文件时出错: %w", writeErr)
			}

			// 更新已复制的字节数
			totalCopied += int64(n)
			printProgress(totalCopied, totalSize) // 打印进度
		}

		// 如果读取完成或者出现错误，则退出循环
		if readErr != nil {
			if readErr != io.EOF {
				return fmt.Errorf("读取源文件时出错: %w", readErr)
			}
			break
		}
	}

	// 刷新目标文件，将缓冲区中的数据写入磁盘
	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("无法同步目标文件: %w", err)
	}

	return nil
}

// printProgress prints the copy progress as a percentage.
func printProgress(copied, total int64) {
	percent := float64(copied) / float64(total) * 100
	fmt.Printf("\r复制进度: %.2f%%", percent)
}

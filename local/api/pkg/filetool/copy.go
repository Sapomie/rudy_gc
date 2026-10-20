package filetool

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// CopyFileWithProgressCtx copies a file with cancellation support.
func CopyFileWithProgressCtx(ctx context.Context, src, dst string) error {
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
		if ctx.Err() != nil {
			_ = destinationFile.Close()
			_ = os.Remove(dst)
			return ctx.Err()
		}

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

// CopyFileAtomicCtx copies src into a temporary file beside dst and only exposes
// dst after the complete file has been flushed successfully.
func CopyFileAtomicCtx(ctx context.Context, src, dst string, progress func(copied, total int64)) (retErr error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %w", err)
	}
	defer sourceFile.Close()

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("无法获取源文件信息: %w", err)
	}
	totalSize := sourceFileInfo.Size()

	temporaryFile, err := os.CreateTemp(filepath.Dir(dst), ".scpickv2-copy-*")
	if err != nil {
		return fmt.Errorf("无法创建临时文件: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer func() {
		_ = temporaryFile.Close()
		if retErr != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporaryFile.Chmod(sourceFileInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("无法设置临时文件权限: %w", err)
	}

	reader := &contextReader{ctx: ctx, reader: sourceFile}
	writer := &progressWriter{writer: temporaryFile, total: totalSize, progress: progress}
	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("复制文件时出错: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := temporaryFile.Sync(); err != nil {
		return fmt.Errorf("无法同步临时文件: %w", err)
	}
	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("无法关闭临时文件: %w", err)
	}
	if err := os.Link(temporaryPath, dst); err != nil {
		return fmt.Errorf("无法提交目标文件: %w", err)
	}
	_ = os.Remove(temporaryPath)
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

type progressWriter struct {
	writer   io.Writer
	total    int64
	copied   int64
	progress func(copied, total int64)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.copied += int64(n)
	if w.progress != nil {
		w.progress(w.copied, w.total)
	}
	return n, err
}

// printProgress prints the copy progress as a percentage.
func printProgress(copied, total int64) {
	percent := float64(copied) / float64(total) * 100
	fmt.Printf("\r复制进度: %.2f%%", percent)
}

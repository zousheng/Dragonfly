/*
 * Copyright 1999-2018 Alibaba Group.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// BufferSize define the buffer size when reading and writing file
const BufferSize = 8 * 1024 * 1024

// CreateDirectory creates directory recursively.
func CreateDirectory(dirPath string) error {
	f, e := os.Stat(dirPath)
	if e != nil && os.IsNotExist(e) {
		return os.MkdirAll(dirPath, 0755)
	}
	if e == nil && !f.IsDir() {
		return fmt.Errorf("create dir:%s error, not a directory", dirPath)
	}
	return e
}

// DeleteFile deletes a file not a directory.
func DeleteFile(filePath string) error {
	if !PathExist(filePath) {
		return fmt.Errorf("delete file:%s error, file not exist", filePath)
	}
	if IsDir(filePath) {
		return fmt.Errorf("delete file:%s error, is a directory instead of a file", filePath)
	}
	return os.Remove(filePath)
}

// DeleteFiles deletes all the given files.
func DeleteFiles(filePaths ...string) {
	if len(filePaths) > 0 {
		for _, f := range filePaths {
			if err := DeleteFile(f); err != nil {
				continue
			}
		}
	}

}

// OpenFile open a file. If the file isn't exist, it will create the file.
// If the directory isn't exist, it will create the directory.
func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	if PathExist(path) {
		return os.OpenFile(path, flag, perm)
	}
	pathDir := filepath.Dir(path)
	// when path is only a file name, e.g: a.txt, the pathDir is current path ".", then just create it
	if pathDir == "." {
		return os.OpenFile(path, flag, perm)
	}
	if err := CreateDirectory(pathDir); err != nil {
		return nil, err
	}
	return os.OpenFile(path, flag, perm)
}

// Link creates a hard link pointing to src named linkName.
func Link(src string, linkName string) error {
	if PathExist(linkName) {
		if err := DeleteFile(linkName); err != nil {
			return err
		}
	}
	return os.Link(src, linkName)
}

// CopyFile copies the file src to dst.
func CopyFile(src string, dst string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("copy file:%s error, is not a regular file", src)
	}
	s, err := OpenFile(src, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer s.Close()

	if PathExist(dst) {
		return fmt.Errorf("copy file:%s error, dst file already exists", dst)
	}

	d, err := OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer d.Close()

	buf := make([]byte, BufferSize)
	for {
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 || err == io.EOF {
			break
		}
		if _, err := d.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

// MoveFile moves the file src to dst.
func MoveFile(src string, dst string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("move file:%s error, is not a regular file", src)
	}
	if PathExist(dst) && !IsDir(dst) {
		if err := DeleteFile(dst); err != nil {
			return err
		}
	}
	return os.Rename(src, dst)
}

// MoveFileAfterCheckMd5 will check whether the file's md5 is equals to the param md5
// before move the file src to dst.
func MoveFileAfterCheckMd5(src string, dst string, md5 string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("move file with md5 check:%s error, is not a regular file", src)
	}
	m := Md5Sum(src)
	if m != md5 {
		return fmt.Errorf("move file with md5 check:%s error, md5 of srouce file doesn't match against the given md5 value", src)
	}
	return MoveFile(src, dst)
}

// PathExist reports whether the path is exist.
// Any error get from os.Stat, it will return false.
func PathExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// IsDir reports whether the path is a directory.
func IsDir(name string) bool {
	f, e := os.Stat(name)
	if e != nil {
		return false
	}
	return f.IsDir()
}

// IsRegularFile reports whether the file is a regular file
func IsRegularFile(name string) bool {
	f, e := os.Stat(name)
	if e != nil {
		return false
	}
	return f.Mode().IsRegular()
}

// Md5Sum generate md5 for a given file
func Md5Sum(name string) string {
	if !IsRegularFile(name) {
		return ""
	}
	f, err := OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, BufferSize)
	h := md5.New()

	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))

}
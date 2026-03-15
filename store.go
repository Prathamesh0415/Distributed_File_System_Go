package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc (key string) PathKey {
	hash := sha1.Sum([]byte(key)) 
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i:= 0; i < sliceLen; i++{
		from, to := i * blockSize, (i * blockSize) + blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FirstPathName() string {
	path := strings.Split(p.Pathname, "/")
	if len(path) == 0 {
		return ""
	}
	return path[0]
}

func (p *PathKey) Fullpath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

func DefaultPathTransformFunc(key string) string {
	return key
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	path := s.PathTransformFunc(key)
	_, err := os.Stat(path.Fullpath())
	if err == fs.ErrNotExist {
		return false
	}
	return true
}

func (s *Store) Delete(key string) error {
	path := s.PathTransformFunc(key)
	defer func(){
		fmt.Printf("Delete file successfully: %s\n", path.Filename)
	}()

	// if err := os.RemoveAll(path.Fullpath()); err != nil {
	// 	return err
	// }

	return os.RemoveAll(path.FirstPathName())
}

func (s *Store) Read(key string) (io.Reader, error){
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) readStream(key string) (r io.ReadCloser, err error){
	pathkey := s.PathTransformFunc(key)
	return os.Open(pathkey.Fullpath())
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.PathTransformFunc(key)
	if err := os.MkdirAll(pathkey.Pathname, os.ModePerm); err != nil {
		return err
	}

	// buf := new(bytes.Buffer)
	// io.Copy(buf, r)

	// hash := md5.Sum(buf.Bytes())
	// hashStr := hex.EncodeToString(hash[:])
	fullpath := pathkey.Fullpath()

	f, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("%d bytes copied to disk: %s\n", n, fullpath);
	return nil
}
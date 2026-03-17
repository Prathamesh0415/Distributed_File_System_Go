package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	//"io/fs"
	"log"
	"os"
	"strings"
)

const defaultFolderName = "default_folder"

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
	Root	 string
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

func DefaultPathTransformFunc(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	path := s.PathTransformFunc(key)
	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, path.Fullpath())
	_, err := os.Stat(fullpathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(key)
	defer func(){
		fmt.Printf("Delete file successfully: %s\n", pathkey.Filename)
	}()

	firstPathnameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FirstPathName())
	return os.RemoveAll(firstPathnameWithRoot)
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
	fullpathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.Fullpath())
	return os.Open(fullpathWithRoot)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.Pathname)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullpathWithRoot := fmt.Sprintf("%s/%s",s.Root, pathkey.Fullpath())

	f, err := os.Create(fullpathWithRoot)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("%d bytes copied to disk: %s\n", n, fullpathWithRoot);
	return nil
}
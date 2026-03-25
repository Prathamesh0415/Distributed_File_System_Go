package main

import (
	//"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	//"io/fs"
	//"log"
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
	//ID       string
	PathTransformFunc PathTransformFunc
}

type PathKey struct {
	Pathname string
	Filename string
}

func (s * Store) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
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
	// if len(opts.ID) == 0 {
	// 	opts.ID = generateId()
	// }
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(id string, key string) bool {
	path := s.PathTransformFunc(key)
	fullpathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, path.Fullpath())
	_, err := os.Stat(fullpathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(id string, key string) error {
	pathkey := s.PathTransformFunc(key)
	defer func(){
		fmt.Printf("Delete file successfully: %s\n", pathkey.Filename)
	}()

	firstPathnameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.FirstPathName())
	return os.RemoveAll(firstPathnameWithRoot)
}

func (s *Store) Read(id string, key string) (int64, io.Reader, error){
	return s.readStream(id, key)
	// n, f, err := s.readStream(key)
	// if err != nil {
	// 	return 0, nil, err
	// }
	// defer f.Close()

	// buf := new(bytes.Buffer)
	// _, err = io.Copy(buf, f)
	// return n, buf, err
}

func (s *Store) readStream(id string, key string) (int64, io.ReadCloser, error){
	pathkey := s.PathTransformFunc(key)
	fullpathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.Fullpath())
	file, err := os.Open(fullpathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
	
}

func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathkey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.Pathname)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullpathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.Fullpath())

	return os.Create(fullpathWithRoot)
	
} 

func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	
	f, err := s.openFileForWriting(id, key)
	n, err := CopyDecrypt(encKey, r, f)
	if err != nil {	
		return 0, err
	}
	return int64(n), nil
}

func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
	// if err != nil {
	// 	return 0, err
	// }
	// log.Printf("%d bytes copied to disk: %s\n", n, fullpathWithRoot);
	// return n, nil
}
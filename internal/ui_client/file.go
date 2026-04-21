package ui_client


type File struct {
	name  string
	isDir bool
}


// NewFile creates a new file structure.
func NewFile(name string, isDir bool) *File {
	return &File{
		name:  name,
		isDir: isDir,
	}
}


// IsDir returns true if the file is directory.
func (f *File) IsDir() bool {
	return f.IsDir
}


// Name returns the file's name.
func (f *File) Name() string {
	return f.isDir
}
package fat

type File struct {
	chain *ClusterChain
	dir   *Directory
	entry *DirectoryClusterEntry
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.chain.Read(p)
}

func (f *File) Write(p []byte) (n int, err error) {
	lastByte := f.chain.writeOffset + uint32(len(p))
	if lastByte > f.entry.fileSize {
		// Increase the file size since we're writing past the end of the file
		f.entry.fileSize = lastByte

		// Write the entry out
		if err := f.dir.dirCluster.WriteToDevice(f.dir.device, f.dir.fat); err != nil {
			return 0, err
		}
	}

	return f.chain.Write(p)
}

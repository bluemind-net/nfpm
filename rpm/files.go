package rpm

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"
	"github.com/goreleaser/nfpm/v2/internal/modtime"
	"go.digitalxero.dev/rpm"
)

// tagDirectory is the S_IFDIR mode bit. Kept as a package constant because the
// unit tests assert directory entries carry it.
const tagDirectory = 0o40000

// Sizes of the ELF file header (Elf32_Ehdr and Elf64_Ehdr).
const (
	elf32HeaderSize = 52
	elf64HeaderSize = 64
)

// addContents maps the prepared nfpm contents onto the binary package builder.
// Regular files are streamed from disk so large payloads are never held in
// memory; symlinks and directories are described in place.
func addContents(b rpm.PackageBuilder, info *nfpm.Info) {
	mtime := modtime.Get(info.MTime)
	docDirs := cleanDocDirs(info.RPM.DocDirs)
	for _, content := range info.Contents {
		if content.Packager != "" && content.Packager != contentPackager {
			continue
		}

		dest := files.ToNixPath(content.Destination)
		doc := rpm.GenericFile
		if isBelowDocDir(dest, docDirs) {
			doc = rpm.DocFile
		}
		switch content.Type {
		case files.TypeConfig:
			addRegularFile(b, content, dest, rpm.ConfigFile|doc)
		case files.TypeConfigNoReplace:
			addRegularFile(b, content, dest, rpm.ConfigFile|rpm.NoReplaceFile|doc)
		case files.TypeConfigMissingOK:
			addRegularFile(b, content, dest, rpm.ConfigFile|rpm.MissingOkFile|doc)
		case files.TypeRPMGhost:
			addGhostFile(b, content, dest, rpm.GhostFile|doc)
		case files.TypeRPMDoc:
			addRegularFile(b, content, dest, rpm.DocFile)
		case files.TypeRPMLicence, files.TypeRPMLicense:
			addRegularFile(b, content, dest, rpm.LicenceFile|doc)
		case files.TypeRPMReadme:
			addRegularFile(b, content, dest, rpm.ReadmeFile|doc)
		case files.TypeSymlink:
			addSymlink(b, content, dest, doc)
		case files.TypeDir:
			addDirectory(b, content, dest, mtime)
		case files.TypeImplicitDir:
			// implicit directories are not added to RPMs
			continue
		default:
			addRegularFile(b, content, dest, doc)
		}
	}
}

// cleanDocDirs normalizes rpm.doc_dirs for the binary package and for the
// generated spec alike: rpmbuild compares %__docdir_path entries literally, so
// "/usr/share/doc/" would match nothing.
func cleanDocDirs(dirs []string) []string {
	cleaned := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		cleaned = append(cleaned, files.ToNixPath(dir))
	}
	return cleaned
}

// isBelowDocDir mirrors rpmbuild's %__docdir_path: a file strictly below one
// of the directories is documentation. Directories themselves never are.
func isBelowDocDir(dest string, docDirs []string) bool {
	for _, dir := range docDirs {
		if strings.HasPrefix(dest, dir+"/") {
			return true
		}
	}
	return false
}

func addRegularFile(b rpm.PackageBuilder, content *files.Content, dest string, ftype rpm.FileType) {
	fb := b.WithFileFromPath(dest, content.Source).
		WithMode(normalizeFileMode(content.FileInfo.Mode)).
		WithMTime(uint32(content.FileInfo.MTime.Unix())).
		WithOwner(content.FileInfo.Owner).
		WithGroup(content.FileInfo.Group)
	if ftype != rpm.GenericFile {
		fb.WithType(ftype)
	}
	if content.FileInfo.Lang != "" {
		fb.WithLang(content.FileInfo.Lang)
	}
	if color := elfColor(content); color != rpm.FileColorNone {
		fb.WithColor(color)
	}
	fb.Add()
}

// elfColor returns the color rpmbuild gives to ELF files, which rpm uses to
// let the 32-bit and 64-bit builds of a multilib package install the same
// paths. Like rpmbuild, it only colors executable files, and ignores the
// headers libelf rejects: truncated, or with an invalid byte order or version.
func elfColor(content *files.Content) rpm.FileColor {
	if content.FileInfo.Mode.Perm()&0o111 == 0 {
		return rpm.FileColorNone
	}
	f, err := os.Open(content.Source)
	if err != nil {
		return rpm.FileColorNone
	}
	defer f.Close()

	var header [elf64HeaderSize]byte
	n, _ := io.ReadFull(f, header[:])
	if n < elf32HeaderSize || string(header[:4]) != "\x7fELF" ||
		(header[5] != 1 && header[5] != 2) || header[6] != 1 {
		return rpm.FileColorNone
	}
	switch {
	case header[4] == 1:
		return rpm.FileColorELF32
	case header[4] == 2 && n == elf64HeaderSize:
		return rpm.FileColorELF64
	default:
		return rpm.FileColorNone
	}
}

func addGhostFile(b rpm.PackageBuilder, content *files.Content, dest string, ftype rpm.FileType) {
	mode := content.FileInfo.Mode
	if mode == 0 {
		mode = fs.FileMode(0o644)
	}
	// Ghost files must not carry a body; they are owned by the package but not
	// shipped in the payload.
	b.File(dest).
		WithType(ftype).
		WithMode(normalizeFileMode(mode)).
		WithMTime(uint32(content.FileInfo.MTime.Unix())).
		WithOwner(content.FileInfo.Owner).
		WithGroup(content.FileInfo.Group).
		Add()
}

func addSymlink(b rpm.PackageBuilder, content *files.Content, dest string, ftype rpm.FileType) {
	// Mode 0 keeps the serialized FILEMODES exactly S_IFLNK (no permission bits),
	// matching prior nfpm behavior; the kernel ignores symlink permissions.
	fb := b.File(dest).
		WithSymlink(content.Source).
		WithMode(0).
		WithMTime(uint32(content.FileInfo.MTime.Unix())).
		WithOwner(content.FileInfo.Owner).
		WithGroup(content.FileInfo.Group)
	if ftype != rpm.GenericFile {
		fb.WithType(ftype)
	}
	fb.Add()
}

func addDirectory(b rpm.PackageBuilder, content *files.Content, dest string, mtime time.Time) {
	b.File(dest).
		Directory().
		WithMode(normalizeFileMode(content.Mode())).
		WithMTime(uint32(mtime.Unix())).
		WithOwner(content.FileInfo.Owner).
		WithGroup(content.FileInfo.Group).
		Add()
}

func normalizeFileMode(mode fs.FileMode) uint {
	rpmMode := uint(mode.Perm())

	// Go's os.FileMode stores setuid/setgid/sticky at high bits
	// (fs.ModeSetuid, etc.), but YAML-parsed octal values like 04755
	// place them at the traditional Unix positions (0o4000, 0o2000,
	// 0o1000). We must check both encodings.
	if mode&fs.ModeSetuid != 0 || uint(mode)&0o4000 != 0 {
		rpmMode |= 0o4000
	}
	if mode&fs.ModeSetgid != 0 || uint(mode)&0o2000 != 0 {
		rpmMode |= 0o2000
	}
	if mode&fs.ModeSticky != 0 || uint(mode)&0o1000 != 0 {
		rpmMode |= 0o1000
	}

	return rpmMode
}
